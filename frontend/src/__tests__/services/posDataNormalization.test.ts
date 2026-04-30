import { describe, expect, it } from 'vitest';

import {
  buildCachedPosProducts,
  dedupePosCustomers,
  mapCachedProductsToPosProducts,
  normalizePosCustomerRecord,
} from '../../services/posDataNormalization';

describe('normalizePosCustomerRecord', () => {
  it('normalizes snake_case customer fields into cached camelCase fields', () => {
    const customer = normalizePosCustomerRecord({
      id: 7,
      name: 'Jane Doe',
      phone: '09171234567',
      email: 'jane@example.com',
      address: 'Lipa City',
      rfid_card_id: 'RF-123',
      loyalty_points: 12,
    });

    expect(customer).toEqual({
      id: 7,
      name: 'Jane Doe',
      phone: '09171234567',
      email: 'jane@example.com',
      address: 'Lipa City',
      rfidCardId: 'RF-123',
      loyaltyPoints: 12,
      loyalty_points: 12,
      synced: true,
    });
  });
});

describe('dedupePosCustomers', () => {
  it('deduplicates normalized cached customers by phone number', () => {
    const customers = dedupePosCustomers([
      { id: 1, name: 'Jane Doe', phone: '09171234567', loyalty_points: 5 },
      { id: 2, name: 'Jane Duplicate', phone: '09171234567', loyalty_points: 9 },
      { id: 3, name: 'John Doe', phone: '09998887777', loyalty_points: 2 },
    ]);

    expect(customers).toHaveLength(2);
    expect(customers[0].id).toBe(1);
    expect(customers[1].id).toBe(3);
  });
});

describe('buildCachedPosProducts', () => {
  it('builds cached products with numeric branch stock and attached categories', () => {
    const products = buildCachedPosProducts(
      [
        {
          id: 1,
          name: 'Tire A',
          price: 100,
          branch_stock: '7',
          category_id: 4,
          brand_id: 2,
        },
      ],
      [{ id: 4, name: 'TIRES' }],
    );

    expect(products).toEqual([
      {
        id: 1,
        name: 'Tire A',
        price: 100,
        branch_stock: 7,
        category_id: 4,
        brand_id: 2,
        category: { id: 4, name: 'TIRES' },
      },
    ]);
  });
});

describe('mapCachedProductsToPosProducts', () => {
  it('maps cached products into the slimmer POS product shape', () => {
    const products = mapCachedProductsToPosProducts([
      {
        id: 1,
        name: 'Tire A',
        price: 100,
        branch_stock: 7,
        category_id: 4,
        brand_id: 2,
        category: { id: 4, name: 'TIRES' },
      },
    ]);

    expect(products).toEqual([
      {
        id: 1,
        name: 'Tire A',
        price: 100,
        branch_stock: 7,
        category_id: 4,
        category: { name: 'TIRES' },
      },
    ]);
  });
});
