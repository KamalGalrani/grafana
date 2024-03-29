import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import lodash from 'lodash'; // eslint-disable-line lodash/import-scope
import React from 'react';

import * as analytics from '../../Analytics';

import { MatcherFilter } from './MatcherFilter';

const logInfoSpy = jest.spyOn(analytics, 'logInfo');

describe('Analytics', () => {
  beforeEach(() => {
    lodash.debounce = jest.fn().mockImplementation((fn) => {
      fn.cancel = () => {};
      return fn;
    });
  });

  it('Sends log info when filtering alert instances by label', async () => {
    render(<MatcherFilter onFilterChange={jest.fn()} />);

    const searchInput = screen.getByTestId('search-query-input');
    await userEvent.type(searchInput, 'job=');

    expect(logInfoSpy).toHaveBeenCalledWith(analytics.LogMessages.filterByLabel);
  });

  it('should call onChange handler', async () => {
    const onFilterMock = jest.fn();

    render(<MatcherFilter defaultQueryString="foo" onFilterChange={onFilterMock} />);

    const searchInput = screen.getByTestId('search-query-input');
    await userEvent.type(searchInput, '=bar');

    expect(onFilterMock).toHaveBeenLastCalledWith('foo=bar');
  });
});
