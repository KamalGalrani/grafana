// SPDX-License-Identifier: AGPL-3.0-only
// Provenance-includes-location: https://github.com/tilt-dev/tilt-apiserver/blob/main/pkg/storage/filepath/watchset.go
// Provenance-includes-license: Apache-2.0
// Provenance-includes-copyright: The Kubernetes Authors.

package file

import (
	"sync"

	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/apiserver/pkg/storage"
)

// Keeps track of which watches need to be notified
type WatchSet struct {
	mu      sync.RWMutex
	nodes   map[int]*watchNode
	counter int
}

func NewWatchSet() *WatchSet {
	return &WatchSet{
		nodes:   make(map[int]*watchNode, 20),
		counter: 0,
	}
}

// Creates a new watch with a unique id, but
// does not start sending events to it until start() is called.
func (s *WatchSet) newWatch() *watchNode {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counter++

	return &watchNode{
		id:       s.counter,
		s:        s,
		updateCh: make(chan watch.Event),
		outCh:    make(chan watch.Event),
	}
}

func (s *WatchSet) cleanupWatchers() {
	// Doesn't protect the below access on nodes slice since it won't ever be modified during cleanup
	for _, w := range s.nodes {
		w.Stop()
	}
}

func (s *WatchSet) notifyWatchers(ev watch.Event) {
	s.mu.RLock()
	for _, w := range s.nodes {
		w.updateCh <- ev
	}
	s.mu.RUnlock()
}

type watchNode struct {
	s        *WatchSet
	id       int
	updateCh chan watch.Event
	outCh    chan watch.Event
}

// Start sending events to this watch.
func (w *watchNode) Start(p storage.SelectionPredicate, initEvents []watch.Event) {
	w.s.mu.Lock()
	w.s.nodes[w.id] = w
	w.s.mu.Unlock()

	go func() {
		for _, e := range initEvents {
			w.outCh <- e
		}

		for e := range w.updateCh {
			ok, err := p.Matches(e.Object)
			if err != nil {
				continue
			}

			if !ok {
				continue
			}
			w.outCh <- e
		}
		close(w.outCh)
	}()
}

func (w *watchNode) Stop() {
	w.s.mu.Lock()
	delete(w.s.nodes, w.id)
	w.s.mu.Unlock()

	close(w.updateCh)
}

func (w *watchNode) ResultChan() <-chan watch.Event {
	return w.outCh
}
