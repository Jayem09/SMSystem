import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { get, post, put, del } from '../api/axios';
import type { QueryParams } from '../types/api';
import { getIsOfflineMode } from '../context/AuthContext';
import offlineStorage, { type LocalCustomer } from '../services/offlineStorage';
import {
  attachCategoriesToProducts,
  buildCachedPosProducts,
  dedupePosCustomers,
  type CachedPosCategory,
  type CachedPosCustomer,
  type CachedPosData,
  mapLocalCustomerToCachedPosCustomer,
} from '../services/posDataNormalization';

export function useProductsQuery(params?: QueryParams) {
  return useQuery({
    queryKey: ['products', params],
    queryFn: () => get('/api/products', { params }).then(res => res.data),
  });
}

export function useCategoriesQuery(params?: QueryParams) {
  return useQuery({
    queryKey: ['categories', params],
    queryFn: () => get('/api/categories', { params }).then(res => res.data),
  });
}

export function useBrandsQuery(params?: QueryParams) {
  return useQuery({
    queryKey: ['brands', params],
    queryFn: () => get('/api/brands', { params }).then(res => res.data),
  });
}

export function useOrdersQuery(params?: QueryParams) {
  return useQuery({
    queryKey: ['orders', params],
    queryFn: () => get('/api/orders', { params }).then(res => res.data),
  });
}

export function useCustomersQuery(params?: QueryParams) {
  return useQuery({
    queryKey: ['customers', params],
    queryFn: () => get('/api/customers', { params }).then(res => res.data),
  });
}

export function useSuppliersQuery(params?: QueryParams) {
  return useQuery({
    queryKey: ['suppliers', params],
    queryFn: () => get('/api/suppliers', { params }).then(res => res.data),
  });
}

export function useDashboardStats(days: number = 30, branchId?: string) {
  return useQuery({
    queryKey: ['dashboard', 'stats', days, branchId],
    queryFn: () => {
      const params: Record<string, string> = { days: days.toString() };
      if (branchId && branchId !== 'ALL') {
        params.branch_id = branchId;
      }
      return get('/api/dashboard', { params }).then(res => res.data);
    },
    staleTime: 0, // Always refetch on branch change
    refetchOnWindowFocus: false,
  });
}

export function useCreateProduct() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: unknown) => post('/api/products', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['products'] });
    },
  });
}

export function useUpdateProduct() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: unknown }) => put(`/api/products/${id}`, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['products'] });
    },
  });
}

export function useDeleteProduct() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => del(`/api/products/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['products'] });
    },
  });
}

export function useCreateOrder() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: unknown) => post('/api/orders', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['orders'] });
      queryClient.invalidateQueries({ queryKey: ['dashboard'] });
    },
  });
}

export function useCreateCustomer() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: unknown) => post('/api/customers', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['customers'] });
    },
  });
}

export function useUpdateCustomer() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: unknown }) => put(`/api/customers/${id}`, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['customers'] });
    },
  });
}

export function useCreateCategory() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: unknown) => post('/api/categories', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['categories'] });
    },
  });
}

export function useUpdateCategory() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: unknown }) => put(`/api/categories/${id}`, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['categories'] });
    },
  });
}

export function useCreateBrand() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: unknown) => post('/api/brands', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['brands'] });
    },
  });
}

export function useUpdateBrand() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: unknown }) => put(`/api/brands/${id}`, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['brands'] });
    },
  });
}

export function usePOSData(branchId?: string) {
  // React Query handles caching - try online first, fallback to cache on error
  // POS is always branch-scoped. If no concrete branch is selected, do not fetch products.
  return useQuery({
    queryKey: ['pos', 'data', branchId],
    queryFn: async (): Promise<CachedPosData> => {
      const categories = offlineStorage.getCategories() as CachedPosCategory[];
      const customers = offlineStorage
        .getCustomers()
        .map((customer) => mapLocalCustomerToCachedPosCustomer(customer))
        .filter((customer): customer is CachedPosCustomer => customer !== null);

      if (!branchId) {
        return {
          products: [],
          categories,
          customers,
        };
      }
      
      // Try online first
      if (!getIsOfflineMode()) {
        try {
          const [pRes, cRes, custRes] = await Promise.all([
            get(`/api/products?all=1&branch_id=${branchId}`),
            get('/api/categories'),
            get('/api/customers'),
          ]);
          
          const rawProducts = (pRes.data as { products?: Record<string, unknown>[] }).products 
            || (pRes.data as { data?: Record<string, unknown>[] }).data 
            || [];
          const categories = (cRes.data as { categories?: CachedPosCategory[] }).categories || [];
          const rawCustomers = (custRes.data as { customers?: Record<string, unknown>[] }).customers || [];
          const products = buildCachedPosProducts(rawProducts, categories);
          const uniqueCustomers = dedupePosCustomers(rawCustomers);
          
          // Cache for offline
          offlineStorage.saveProductsByBranch(branchId, products);
          offlineStorage.saveCategories(categories);
          offlineStorage.saveCustomers(uniqueCustomers);
          
          return {
            products,
            categories,
            customers: uniqueCustomers,
          };
        } catch (err) {
          console.warn('[usePOSData] API fetch failed, falling back to cache:', err);
          // Fall through to cache below
        }
      }
      
      // OFFLINE or API failed - load from cached storage
      const products = attachCategoriesToProducts(offlineStorage.getProductsByBranch(branchId), categories);
      
      return {
        products,
        categories,
        customers,
      };
    },
    staleTime: 5 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
  });
}

// Function to find customer by RFID card (works offline!)
export function findCustomerByRfid(rfidCode: string) {
  const customers = offlineStorage.getCustomers();
  // Look for RFID match in cached customers
  return customers.find((customer) => {
    const legacyCustomer = customer as LocalCustomer & { rfid_card_id?: string };
    return legacyCustomer.rfid_card_id === rfidCode || customer.rfidCardId === rfidCode;
  });
}
