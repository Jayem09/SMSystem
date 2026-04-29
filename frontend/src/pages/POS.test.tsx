import React, { act } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createRoot, type Root } from 'react-dom/client';

const {
  mockApiGet,
  mockShowToast,
  mockRefetch,
  mockUseQuery,
  mockQueryClient,
} = vi.hoisted(() => ({
  mockApiGet: vi.fn(),
  mockShowToast: vi.fn(),
  mockRefetch: vi.fn(),
  mockUseQuery: vi.fn(),
  mockQueryClient: { invalidateQueries: vi.fn() },
}));

vi.mock('../api/axios', () => ({
  default: {
    get: mockApiGet,
    post: vi.fn(),
  },
}));

vi.mock('@tanstack/react-query', () => ({
  useQueryClient: () => mockQueryClient,
  useQuery: mockUseQuery,
}));

vi.mock('../hooks/useAuth', () => ({
  useAuth: () => ({ user: { id: 1, role: 'admin', branch_id: 4 } }),
}));

vi.mock('../hooks/usePOS', () => ({
  usePOS: () => ({
    state: {
      products: [],
      categories: [],
      customers: [],
      cart: [
        {
          id: 1,
          name: 'Test Tire',
          price: 2058,
          quantity: 1,
          branch_stock: 10,
          category_id: 1,
        },
      ],
      search: '',
      selectedCategory: '',
      loading: false,
      error: null,
    },
    dispatch: vi.fn(),
    addToCart: vi.fn(),
    removeFromCart: vi.fn(),
    updateQuantity: vi.fn(),
    clearCart: vi.fn(),
    setSearch: vi.fn(),
    setCategory: vi.fn(),
    subtotal: 2058,
    filteredProducts: [],
    lastAddBlocked: false,
  }),
}));

vi.mock('../hooks/useQueries', () => ({
  usePOSData: () => ({
    data: null,
    isLoading: false,
    error: null,
    refetch: mockRefetch,
  }),
}));

vi.mock('../context/ToastContext', () => ({
  useToast: () => ({ showToast: mockShowToast }),
}));

vi.mock('../context/AuthContext', () => ({
  getIsOfflineMode: () => false,
}));

vi.mock('../components/Modal', () => ({
  default: ({ open, title, children }: { open: boolean; title: string; children: React.ReactNode }) =>
    open ? <div><h2>{title}</h2>{children}</div> : null,
}));

vi.mock('../components/Receipt', () => ({ printReceipt: vi.fn() }));
vi.mock('../components/DeliveryReceipt', () => ({ printDeliveryReceipt: vi.fn() }));
vi.mock('../services/offlineStorage', () => ({
  default: {
    getCustomers: vi.fn(() => []),
    saveOrder: vi.fn(),
  },
}));
vi.mock('../services/syncQueue', () => ({
  createSyncQueueItem: vi.fn(),
  enqueueSyncItem: vi.fn(),
  findLatestEntitySyncItem: vi.fn(() => null),
}));
vi.mock('../services/offlinePosStock', () => ({
  mergeOfflineBranchProductsIntoPosData: vi.fn((data) => data),
  persistOfflineBranchStockDeduction: vi.fn(),
}));
vi.mock('../services/dashboardRefresh', () => ({
  invalidateDashboardQueries: vi.fn(),
}));

import POS from './POS';

describe('POS checkout modal', () => {
  let root: Root | null = null;

  beforeEach(() => {
    vi.clearAllMocks();
    document.body.innerHTML = '<div id="root"></div>';
    mockApiGet.mockResolvedValue({ data: { staff_directory: [] } });
    mockUseQuery.mockReturnValue({ data: { staff_directory: [] } });
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
      root?.render(<POS />);
      await Promise.resolve();
      await Promise.resolve();
    });
  }

  it('does not show the confusing payment preview card in the finalize sale modal', async () => {
    await renderPOSPage();

    const openButton = Array.from(document.querySelectorAll('button')).find((button) =>
      button.textContent?.includes('PROCESS CHECKOUT'),
    ) as HTMLButtonElement;

    openButton.click();

    await act(async () => {
      await Promise.resolve();
    });

    expect(document.body.textContent).toContain('Finalize Sale');
    expect(document.body.textContent).toContain('Amount Paid');
    expect(document.body.textContent).not.toContain('If Confirmed Now');
    expect(document.body.textContent).not.toContain('Payment Status:');
    expect(document.body.textContent).not.toContain('Hold Sale keeps unpaid amounts as receivables until the order is completed.');
  });
});
