import React, { act } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createRoot, type Root } from 'react-dom/client';

const { mockApiGet } = vi.hoisted(() => ({
  mockApiGet: vi.fn(),
}));

vi.mock('../api/axios', () => ({
  default: {
    get: mockApiGet,
  },
}));

import Transactions from './Transactions';

describe('Transactions', () => {
  let root: Root | null = null;

  beforeEach(() => {
    vi.clearAllMocks();
    document.body.innerHTML = '<div id="root"></div>';
    mockApiGet.mockResolvedValue({
      data: {
        transactions: [{
          date: '2026-04-29',
          order_id: 42,
          receipt_type: 'SI',
          branch_name: 'LIPA A',
          customer_name: 'John Doe',
          service_advisor_name: 'Joel',
          mechanic_name: 'Jomar',
          item_name: 'Accelera Tire',
          unit_of_measure: 'pc',
          category_name: 'Tires',
          quantity: 2,
          unit_price: 1029,
          subtotal: 2058,
          payment_method: 'cash',
          order_status: 'completed',
        }],
      },
    });
  });

  afterEach(() => {
    if (root) {
      act(() => root?.unmount());
      root = null;
    }
    document.body.innerHTML = '';
  });

  async function renderPage() {
    const container = document.getElementById('root');
    if (!container) throw new Error('Missing root');

    root = createRoot(container);
    await act(async () => {
      root?.render(<Transactions />);
      await Promise.resolve();
      await Promise.resolve();
    });
  }

  it('renders both Service Advisor and Mechanic columns', async () => {
    await renderPage();

    expect(document.body.textContent).toContain('Service Advisor');
    expect(document.body.textContent).toContain('Mechanic');
    expect(document.body.textContent).toContain('Joel');
    expect(document.body.textContent).toContain('Jomar');
    
    // Check placeholder via input element
    const searchInput = document.querySelector('input[placeholder*="service advisor"]');
    expect(searchInput).toBeTruthy();
  });
});