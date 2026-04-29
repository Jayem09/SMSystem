import React, { act } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createRoot, type Root } from 'react-dom/client';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

const { mockApiGet } = vi.hoisted(() => ({ mockApiGet: vi.fn() }));

vi.mock('../api/axios', () => ({
  default: { get: mockApiGet },
}));

vi.mock('../hooks/useAuth', () => ({
  useAuth: () => ({ user: { id: 1, role: 'cashier', branch_id: 4 } }),
}));

vi.mock('../context/ToastContext', () => ({
  useToast: () => ({ showToast: vi.fn() }),
}));

vi.mock('../services/offlineStorage', () => ({
  default: {
    getProducts: () => [],
    getCustomers: () => [],
    getCategories: () => [],
  },
}));

vi.mock('../hooks/usePOS', () => ({
  usePOS: () => ({
    state: {
      products: [],
      categories: [],
      customers: [],
      cart: [],
      search: '',
      selectedCategory: null,
      loading: false,
      error: null,
    },
    dispatch: vi.fn(),
  }),
}));

vi.mock('../hooks/useQueries', () => ({
  usePOSData: () => ({
    data: null,
    isLoading: false,
    error: null,
  }),
}));

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { enabled: false },
  },
});

import POS from './POS';

describe('POS checkout staff attribution', () => {
  let root: Root | null = null;

  beforeEach(() => {
    vi.clearAllMocks();
    document.body.innerHTML = '<div id="root"></div>';
    mockApiGet.mockImplementation((url: string) => {
      if (url === '/api/settings') {
        return Promise.resolve({
          data: {
            staff_directory: [
              { name: 'Mike', type: 'service_advisor' },
              { name: 'Jun', type: 'mechanic' },
              { name: 'Ken', type: 'tintner' },
              { name: 'Paul', type: 'carwasher' },
            ],
          },
        });
      }
      return Promise.resolve({ data: {} });
    });
  });

  afterEach(() => {
    if (root) {
      act(() => root?.unmount());
      root = null;
    }
    document.body.innerHTML = '';
  });

  async function renderPOSPage() {
    const container = document.getElementById('root');
    if (!container) throw new Error('Missing root container');

    root = createRoot(container);
    await act(async () => {
      root?.render(
        <QueryClientProvider client={queryClient}>
          <POS />
        </QueryClientProvider>
      );
      await Promise.resolve();
      await Promise.resolve();
    });
  }

  it('renders separate staff attribution selectors for Mechanic, Tintner, and Carwasher', async () => {
    await renderPOSPage();

    const openButton = Array.from(document.querySelectorAll('button')).find((button) =>
      button.textContent?.includes('PROCESS CHECKOUT'),
    ) as HTMLButtonElement;

    openButton.click();

    await act(async () => {
      await Promise.resolve();
    });

    expect(document.body.textContent).toContain('Service Advisor');
    expect(document.body.textContent).toContain('Mechanic');
    expect(document.body.textContent).toContain('Tintner');
    expect(document.body.textContent).toContain('Carwasher');
  });
});