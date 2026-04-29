import type { POSProduct } from '../hooks/usePOS';
import type { LocalCustomer, LocalProduct } from './offlineStorage';

export interface CachedPosCategory {
  id: number;
  name: string;
}

export type CachedPosProduct = LocalProduct & {
  category?: CachedPosCategory | null;
};

export type CachedPosCustomer = LocalCustomer & {
  id: number;
  loyalty_points?: number;
};

export interface CachedPosData {
  products: CachedPosProduct[];
  categories: CachedPosCategory[];
  customers: CachedPosCustomer[];
}

function toFiniteNumber(value: unknown, fallback: number): number {
  const numericValue = typeof value === 'number' ? value : Number(value);
  return Number.isFinite(numericValue) ? numericValue : fallback;
}

function toOptionalString(value: unknown): string | undefined {
  return typeof value === 'string' && value.trim() !== '' ? value : undefined;
}

export function normalizePosCustomerRecord(record: Record<string, unknown>): CachedPosCustomer | null {
  const id = toFiniteNumber(record.id, Number.NaN);
  if (!Number.isFinite(id)) {
    return null;
  }

  const loyaltyPoints = toFiniteNumber(record.loyaltyPoints ?? record.loyalty_points, 0);

  return {
    id,
    name: typeof record.name === 'string' ? record.name : '',
    phone: typeof record.phone === 'string' ? record.phone : '',
    email: toOptionalString(record.email),
    address: toOptionalString(record.address),
    rfidCardId: toOptionalString(record.rfidCardId ?? record.rfid_card_id),
    loyaltyPoints,
    loyalty_points: loyaltyPoints,
    synced: true,
  };
}

export function dedupePosCustomers(records: Record<string, unknown>[]): CachedPosCustomer[] {
  const customersByPhone = new Map<string, CachedPosCustomer>();

  for (const record of records) {
    const customer = normalizePosCustomerRecord(record);
    if (!customer || !customer.phone || customersByPhone.has(customer.phone)) {
      continue;
    }

    customersByPhone.set(customer.phone, customer);
  }

  return Array.from(customersByPhone.values());
}

export function attachCategoriesToProducts(
  products: LocalProduct[],
  categories: CachedPosCategory[],
): CachedPosProduct[] {
  return products.map((product) => ({
    ...product,
    category: categories.find((category) => category.id === product.category_id) ?? null,
  }));
}

export function buildCachedPosProducts(
  records: Record<string, unknown>[],
  categories: CachedPosCategory[],
): CachedPosProduct[] {
  const products: LocalProduct[] = records.map((record) => ({
    id: toFiniteNumber(record.id, 0),
    name: typeof record.name === 'string' ? record.name : '',
    description: toOptionalString(record.description),
    price: toFiniteNumber(record.price, 0),
    cost_price: record.cost_price === undefined ? undefined : toFiniteNumber(record.cost_price, 0),
    branch_stock: toFiniteNumber(record.branch_stock, 0),
    stock: record.stock === undefined ? undefined : toFiniteNumber(record.stock, 0),
    category_id: toFiniteNumber(record.category_id, 0),
    brand_id: toFiniteNumber(record.brand_id, 0),
    image_url: toOptionalString(record.image_url),
    size: toOptionalString(record.size),
    is_service: typeof record.is_service === 'boolean' ? record.is_service : undefined,
    is_reward: typeof record.is_reward === 'boolean' ? record.is_reward : undefined,
    points_required: record.points_required === undefined ? undefined : toFiniteNumber(record.points_required, 0),
    category_name: toOptionalString(record.category_name),
    brand_name: toOptionalString(record.brand_name),
  }));

  return attachCategoriesToProducts(products, categories);
}

export interface PosCustomer {
  id: number;
  name: string;
  email?: string;
  phone?: string;
  address?: string;
  rfidCardId?: string;
  loyaltyPoints?: number;
  loyalty_points?: number;
}

export function mapLocalCustomerToCachedPosCustomer(customer: LocalCustomer): CachedPosCustomer | null {
  if (typeof customer.id !== 'number') {
    return null;
  }

  const loyaltyPoints = customer.loyaltyPoints ?? 0;

  return {
    ...customer,
    id: customer.id,
    loyaltyPoints,
    loyalty_points: loyaltyPoints,
  };
}

export function mapCachedCustomerToPosCustomer(customer: LocalCustomer): PosCustomer | null {
  if (typeof customer.id !== 'number') {
    return null;
  }

  const loyaltyPoints = customer.loyaltyPoints ?? 0;

  return {
    id: customer.id,
    name: customer.name,
    email: customer.email,
    phone: customer.phone,
    address: customer.address,
    rfidCardId: customer.rfidCardId,
    loyaltyPoints,
    loyalty_points: loyaltyPoints,
  };
}

export function mapCachedProductsToPosProducts<T extends LocalProduct & { category?: { name: string } | null }>(products: T[]): POSProduct[] {
  return products.map((product) => ({
    id: product.id,
    name: product.name,
    price: product.price,
    branch_stock: product.branch_stock,
    is_service: product.is_service,
    is_reward: product.is_reward,
    points_required: product.points_required,
    category_id: product.category_id,
    category: product.category ? { name: product.category.name } : undefined,
  }));
}
