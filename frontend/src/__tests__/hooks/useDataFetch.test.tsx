import React, { act } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createRoot, type Root } from 'react-dom/client';
import { useDataFetch } from '../../hooks/useDataFetch';
import type { ApiResponse } from '../../types/api';

const { mockGetIsOfflineMode } = vi.hoisted(() => ({
  mockGetIsOfflineMode: vi.fn(),
}));

vi.mock('../../context/AuthContext', () => ({
  getIsOfflineMode: mockGetIsOfflineMode,
}));

interface HarnessProps {
  queryKey: unknown[];
  responseData: { value: string };
  onQuery: () => void;
}

function FetchHarness({ queryKey, responseData, onQuery }: HarnessProps) {
  const { data, isLoading } = useDataFetch({
    queryKey,
    queryFn: async (): Promise<ApiResponse<{ value: string }>> => {
      onQuery();
      return {
        data: responseData,
        status: 200,
        statusText: 'OK',
        headers: {},
      };
    },
  });

  return <div>{isLoading ? 'loading' : data?.value ?? 'empty'}</div>;
}

describe('useDataFetch', () => {
  let root: Root | null = null;

  beforeEach(() => {
    mockGetIsOfflineMode.mockReturnValue(false);
    document.body.innerHTML = '<div id="root"></div>';
  });

  afterEach(() => {
    if (root) {
      act(() => {
        root?.unmount();
      });
      root = null;
    }
    document.body.innerHTML = '';
  });

  async function renderHarness(props: HarnessProps) {
    const container = document.getElementById('root');
    if (!container) {
      throw new Error('Missing root container');
    }

    if (!root) {
      root = createRoot(container);
    }

    await act(async () => {
      root?.render(<FetchHarness {...props} />);
      await Promise.resolve();
      await Promise.resolve();
    });
  }

  it('fetches only once for a stable query key even with an inline query function', async () => {
    const querySpy = vi.fn();
    const sharedResponse = { value: 'ready' };

    await renderHarness({
      queryKey: ['products', 'all'],
      responseData: sharedResponse,
      onQuery: querySpy,
    });

    expect(querySpy).toHaveBeenCalledTimes(1);
    expect(document.body.textContent).toContain('ready');
  });

  it('refetches when the query key changes', async () => {
    const querySpy = vi.fn();

    await renderHarness({
      queryKey: ['products', 'all'],
      responseData: { value: 'all' },
      onQuery: querySpy,
    });

    await renderHarness({
      queryKey: ['products', 'promo'],
      responseData: { value: 'promo' },
      onQuery: querySpy,
    });

    expect(querySpy).toHaveBeenCalledTimes(2);
    expect(document.body.textContent).toContain('promo');
  });
});
