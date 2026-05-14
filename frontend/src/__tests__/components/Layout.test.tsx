import React, { act } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createRoot, type Root } from 'react-dom/client';
import { MemoryRouter } from 'react-router-dom';

const { mockApiGet, mockLogout, mockNavigate } = vi.hoisted(() => ({
  mockApiGet: vi.fn(),
  mockLogout: vi.fn(),
  mockNavigate: vi.fn(),
}));

vi.mock('../../api/axios', () => ({
  default: {
    get: mockApiGet,
  },
}));

vi.mock('../../hooks/useAuth', () => ({
  useAuth: () => ({
    user: { email: 'admin@example.com', role: 'admin', branch_id: 1 },
    logout: mockLogout,
  }),
}));

vi.mock('../../context/AuthContext', () => ({
  getIsOfflineMode: () => false,
}));

vi.mock('../../utils/branchDisplay', () => ({
  getBranchDisplayName: () => 'Main Branch',
}));

vi.mock('../../components/GlobalSearch', () => ({
  default: () => <div>Search</div>,
}));

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    Outlet: () => <div>Outlet content</div>,
  };
});

import Layout from '../../components/Layout';

describe('Layout shell', () => {
  let root: Root | null = null;

  beforeEach(() => {
    vi.clearAllMocks();
    mockApiGet.mockResolvedValue({ data: { total_actionable: 0 } });
    document.body.innerHTML = '<div id="root"></div>';
  });

  afterEach(() => {
    if (root) {
      act(() => root?.unmount());
      root = null;
    }
    document.body.innerHTML = '';
  });

  it('keeps the content column constrained beside the fixed sidebar', async () => {
    const container = document.getElementById('root');
    if (!container) throw new Error('Missing root container');

    root = createRoot(container);
    await act(async () => {
      root?.render(
        <MemoryRouter>
          <Layout />
        </MemoryRouter>,
      );
      await Promise.resolve();
      await Promise.resolve();
    });

    const main = document.querySelector('main');

    expect(main?.className).toContain('ml-56');
    expect(main?.className).toContain('w-[calc(100%-14rem)]');
    expect(main?.className).toContain('overflow-x-hidden');
  });
});
