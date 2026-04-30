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

vi.mock('../../api/axios', () => ({
  default: {
    get: mockApiGet,
    post: vi.fn(),
  },
}));

vi.mock('@tanstack/react-query', async () => {
  const actual = await vi.importActual<typeof import('@tanstack/react-query')>('@tanstack/react-query');
  return {
    ...actual,
    useQueryClient: () => mockQueryClient,
    useQuery: mockUseQuery,
  };
});

vi.mock('../../hooks/useAuth', () => ({
  useAuth: () => ({ user: { id: 1, role: 'admin', branch_id: 4 } }),
}));

vi.mock('../../hooks/usePOS', () => ({
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

vi.mock('../../hooks/useQueries', () => ({
  usePOSData: () => ({
    data: null,
    isLoading: false,
    error: null,
    refetch: mockRefetch,
  }),
}));

vi.mock('../../context/ToastContext', () => ({
  useToast: () => ({ showToast: mockShowToast }),
}));

vi.mock('../../context/AuthContext', () => ({
  getIsOfflineMode: () => false,
}));

vi.mock('../../components/Modal', () => ({
  default: ({ open, title, children, maxWidth }: { open: boolean; title: string; children: React.ReactNode; maxWidth?: string }) =>
    open ? <div><h2>{title}</h2>{children}</div> : null,
}));

vi.mock('../../components/Receipt', () => ({ printReceipt: vi.fn() }));
vi.mock('../../components/DeliveryReceipt', () => ({ printDeliveryReceipt: vi.fn() }));
vi.mock('../../services/offlineStorage', () => ({
  default: {
    getCustomers: vi.fn(() => []),
    saveOrder: vi.fn(),
  },
}));
vi.mock('../../services/syncQueue', () => ({
  createSyncQueueItem: vi.fn(),
  enqueueSyncItem: vi.fn(),
  findLatestEntitySyncItem: vi.fn(() => null),
}));
vi.mock('../../services/offlinePosStock', () => ({
  mergeOfflineBranchProductsIntoPosData: vi.fn((data) => data),
  persistOfflineBranchStockDeduction: vi.fn(),
}));
vi.mock('../../services/dashboardRefresh', () => ({
  invalidateDashboardQueries: vi.fn(),
}));

vi.mock('../../utils/staffDirectory', () => ({
  filterStaffDirectoryByType: vi.fn((staff, type) => {
    const staffByType: Record<string, { name: string; type: string }> = {
      service_advisor: [{ name: 'Mike', type: 'service_advisor' }],
      mechanic: [{ name: 'Jun', type: 'mechanic' }],
      tintner: [{ name: 'Ken', type: 'tintner' }],
      carwasher: [{ name: 'Paul', type: 'carwasher' }],
    };
    return staffByType[type] || [];
  }),
  normalizeStaffDirectorySettings: vi.fn((settings) => settings?.staff_directory || []),
}));

import POS from '../../pages/POS';

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

  it('hides extra staff roles behind a More roles trigger', async () => {
    mockUseQuery.mockReturnValue({
      data: {
        staff_directory: [
          { name: 'Mike', type: 'service_advisor' },
          { name: 'Jun', type: 'mechanic' },
          { name: 'Ken', type: 'tintner' },
          { name: 'Paul', type: 'carwasher' },
        ],
      },
    });

    await renderPOSPage();

    const openButton = Array.from(document.querySelectorAll('button')).find((button) =>
      button.textContent?.includes('PROCESS CHECKOUT'),
    ) as HTMLButtonElement;

    openButton.click();

    await act(async () => {
      await Promise.resolve();
    });

    // Service Advisor should be visible by default
    expect(document.body.textContent).toContain('Service Advisor');
    // "Personnel" trigger should be visible
    expect(document.body.textContent).toContain('Personnel');
    // Extra roles should NOT be visible by default
    expect(document.body.textContent).not.toContain('Mechanic');
    expect(document.body.textContent).not.toContain('Tintner');
    expect(document.body.textContent).not.toContain('Carwasher');

    // Click "Personnel" to expand
    const moreRolesButton = Array.from(document.querySelectorAll('button')).find((button) =>
      button.textContent?.includes('Personnel'),
    ) as HTMLButtonElement;

    moreRolesButton.click();

    await act(async () => {
      await Promise.resolve();
    });

    // Now all roles should be visible
    expect(document.body.textContent).toContain('Mechanic');
    expect(document.body.textContent).toContain('Tintner');
    expect(document.body.textContent).toContain('Carwasher');
  });
});
