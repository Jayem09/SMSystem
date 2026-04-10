# Data Flow Improvements — 2026-04-10

## Context

SMSytem has a React frontend with direct `api.get()` calls in every page. Issues:
- No caching — same data refetched on every mount
- Inconsistent loading/error states
- POS.tsx has 30+ useState calls (hard to maintain)
- No optimistic updates — UI waits for server response
- UI flicker on data load

The app has a solid foundation (AuthContext, useAuth hook, TauriApi class) but lacks proper server state management.

## Goal

Improve data flow with:
1. TanStack Query for caching/deduplication/optimistic updates
2. Standardized data fetching hooks
3. Loading skeletons
4. POS state refactor (useReducer)

## Architecture

```
Pages → TanStack Query → useDataFetch → useApi → TauriApi → Backend
         (caching)      (loading/error) (typed)    (HTTP)
```

## Changes

### 1. Re-add TanStack Query

**File:** `frontend/package.json`
**Action:** Re-add `@tanstack/react-query`

**File:** `frontend/src/lib/queryClient.ts` (new)
```typescript
import { QueryClient } from '@tanstack/react-query';

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000,      // 5 minutes
      gcTime: 10 * 60 * 1000,        // 10 minutes
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
});
```

**File:** `frontend/src/App.tsx`
**Action:** Wrap app with QueryClientProvider

### 2. Create useDataFetch Hook

**File:** `frontend/src/hooks/useDataFetch.ts` (new)

Provides consistent loading/error/toast handling across all pages.

```typescript
interface UseDataFetchOptions<T> {
  queryKey: string[];
  queryFn: () => Promise<ApiResponse<T>>;
  onError?: (error: unknown) => void;
  enabled?: boolean;
}

interface UseDataFetchResult<T> {
  data: T | null;
  isLoading: boolean;
  error: unknown;
  refetch: () => void;
}
```

Usage:
```typescript
// Before (current)
const [loading, setLoading] = useState(true);
useEffect(() => {
  api.get('/api/products').then(res => {
    setProducts(res.data);
    setLoading(false);
  }).catch(err => {
    showToast('Failed to load', 'error');
  });
}, []);

// After (improved)
const { data: products, isLoading, error, refetch } = useDataFetch({
  queryKey: ['products'],
  queryFn: () => api.get('/api/products'),
  onError: () => showToast('Failed to load', 'error'),
});
```

### 3. Create Loading Skeleton Component

**File:** `frontend/src/components/Skeleton.tsx` (new)

Reusable skeleton for consistent loading UI.

```typescript
interface SkeletonProps {
  variant: 'text' | 'card' | 'table' | 'product';
  count?: number;
}

// Usage
<Skeleton variant="card" count={6} />
<Skeleton variant="table" />
```

### 4. Create usePOS Reducer

**File:** `frontend/src/hooks/usePOS.ts` (new)

Replaces 30+ useState calls in POS.tsx with a single reducer.

```typescript
interface POSState {
  products: Product[];
  categories: Category[];
  customers: Customer[];
  cart: CartItem[];
  search: string;
  selectedCategory: number | null;
  loading: boolean;
  error: string | null;
  // ... checkout states
}

type POSAction =
  | { type: 'SET_PRODUCTS'; payload: Product[] }
  | { type: 'ADD_TO_CART'; payload: Product }
  | { type: 'REMOVE_FROM_CART'; payload: number }
  | { type: 'UPDATE_QUANTITY'; payload: { id: number; delta: number } }
  | { type: 'SET_LOADING'; payload: boolean }
  // ... more actions
```

### 5. Update useApi Hook

**File:** `frontend/src/hooks/useApi.ts`
**Action:** Update to work with TanStack Query patterns

```typescript
// Add mutation helpers
export const useApi = () => {
  const queryClient = useQueryClient();
  
  return {
    // Queries (cached)
    products: {
      list: (params?: QueryParams) => 
        useQuery({
          queryKey: ['products', params],
          queryFn: () => get('/api/products', { params }),
        }),
      // ...
    },
    
    // Mutations (with invalidation)
    products: {
      create: (data: ProductInput) =>
        useMutation({
          mutationFn: (d) => post('/api/products', d),
          onSuccess: () => queryClient.invalidateQueries({ queryKey: ['products'] }),
        }),
      // ...
    },
  };
};
```

## Migration Order

1. Re-add `@tanstack/react-query` to package.json
2. Create `queryClient.ts`
3. Update `App.tsx` with QueryClientProvider
4. Create `useDataFetch.ts` hook
5. Create `Skeleton.tsx` component
6. Update `useApi.ts` with TanStack Query integration
7. Create `usePOS.ts` reducer
8. Migrate Dashboard.tsx to new patterns
9. Migrate Products.tsx to new patterns
10. Migrate POS.tsx to usePOS reducer

## Files to Create

| File | Purpose |
|------|---------|
| `frontend/src/lib/queryClient.ts` | TanStack Query config |
| `frontend/src/hooks/useDataFetch.ts` | Consistent fetch hook |
| `frontend/src/hooks/usePOS.ts` | POS state reducer |
| `frontend/src/components/Skeleton.tsx` | Loading skeleton |

## Files to Modify

| File | Change |
|------|--------|
| `frontend/package.json` | Re-add @tanstack/react-query |
| `frontend/src/App.tsx` | Add QueryClientProvider |
| `frontend/src/hooks/useApi.ts` | Update for TanStack Query |
| `frontend/src/pages/Dashboard.tsx` | Migrate to useDataFetch |
| `frontend/src/pages/Products.tsx` | Migrate to useDataFetch |
| `frontend/src/pages/POS.tsx` | Migrate to usePOS reducer |

## Verification

- [ ] `npm run build` succeeds
- [ ] `npm run lint` passes
- [ ] `npx tsc --noEmit` passes
- [ ] App boots without errors
- [ ] Dashboard loads with skeleton, then data
- [ ] Products page loads with skeleton, then data
- [ ] POS checkout works with reducer state
