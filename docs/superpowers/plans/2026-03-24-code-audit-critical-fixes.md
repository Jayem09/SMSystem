# Code Audit Critical Fixes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix critical bugs and security vulnerabilities identified in the comprehensive code audit to make the SMSytem production-ready for high-scale usage.

**Architecture:** This plan addresses critical issues across the full stack:
- Frontend: React/TypeScript (Auth race conditions, memory leaks, input validation)
- Backend: Go/Gin (Race conditions, stock data integrity, JWT security)
- Shared: API communication patterns

**Tech Stack:** React 19, TypeScript, Go, GORM, MySQL, JWT, Tauri

---

## PHASE 1: CRITICAL BUG FIXES (Data Integrity & Security)

### Task 1.1: Fix Race Condition in Stock Deduction (Backend)

**Files:**
- Modify: `backend/internal/handlers/order_handler.go:130-295`

- [ ] **Step 1: Read the current order_handler.go to understand the Create function**

Run: `read backend/internal/handlers/order_handler.go`

- [ ] **Step 2: Add FOR UPDATE lock to prevent race conditions**

In the `Create` function, find this code around line 142-145:
```go
tx.Model(&models.Batch{}).
    Where("product_id = ? AND branch_id = ?", product.ID, order.BranchID).
    Select("COALESCE(SUM(quantity), 0)").
    Row().Scan(&currentStock)
```

Replace with:
```go
// Lock rows to prevent concurrent overselling
var batches []models.Batch
if err := tx.Model(&models.Batch{}).
    Where("product_id = ? AND branch_id = ? AND quantity > 0", product.ID, order.BranchID).
    ForUpdate().
    Order("expiry_date ASC, created_at ASC").
    Find(&batches).Error; err != nil {
    return fmt.Errorf("failed to lock batches: %v", err)
}

// Calculate total available stock from locked batches
var currentStock int
for _, b := range batches {
    currentStock += b.Quantity
}
```

- [ ] **Step 3: Apply same fix to UpdateStatus function**

Find the UpdateStatus function (around line 340-390) and apply the same ForUpdate() lock pattern when reading batches.

- [ ] **Step 4: Remove "self-healing" legacy batch code**

Remove lines 247-281 (the legacy batch creation magic). This code masks data integrity issues. Instead, the transaction should fail explicitly if there's insufficient stock.

- [ ] **Step 5: Run backend tests**

Run: `cd backend && go test ./... -v 2>&1 | head -50`

- [ ] **Step 6: Commit**

```bash
git add backend/internal/handlers/order_handler.go
git commit -m "fix: add FOR UPDATE locks to prevent stock race conditions"
```

---

### Task 1.2: Fix Order Delete to Restore Stock

**Files:**
- Modify: `backend/internal/handlers/order_handler.go:458-490`

- [ ] **Step 1: Read the Delete function**

Run: `read backend/internal/handlers/order_handler.go:458-490`

- [ ] **Step 2: Add stock restoration logic**

Replace the Delete function with one that restores stock in a transaction:

```go
func (h *OrderHandler) Delete(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
        return
    }

    var order models.Order
    if err := database.DB.First(&order, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
        return
    }

    // Only restore stock if order was completed
    if order.Status != "completed" {
        err = database.DB.Transaction(func(tx *gorm.DB) error {
            if err := tx.Where("order_id = ?", id).Delete(&models.OrderItem{}).Error; err != nil {
                return err
            }
            if err := tx.Delete(&models.Order{}, id).Error; err != nil {
                return err
            }
            return nil
        })
    } else {
        // Restore stock for completed orders
        err = database.DB.Transaction(func(tx *gorm.DB) error {
            var items []models.OrderItem
            if err := tx.Where("order_id = ?", id).Find(&items).Error; err != nil {
                return err
            }

            userID, _ := GetUintFromContext(c, "userID")

            for _, item := range items {
                var product models.Product
                if err := tx.First(&product, item.ProductID).Error; err != nil {
                    return errors.New("product not found: " + strconv.Itoa(int(item.ProductID)))
                }

                if !product.IsService {
                    // Restore stock
                    if err := tx.Model(&product).Update("stock", gorm.Expr("stock + ?", item.Quantity)).Error; err != nil {
                        return fmt.Errorf("failed to restore stock for %s", product.Name)
                    }

                    // Create reverse stock movement
                    movement := models.StockMovement{
                        ProductID: product.ID,
                        WarehouseID: 1, // Default warehouse
                        BranchID: order.BranchID,
                        UserID: &userID,
                        Type: models.MovementTypeIn,
                        Quantity: item.Quantity,
                        Reference: fmt.Sprintf("Order #%d deleted - stock restored", order.ID),
                    }
                    if err := tx.Create(&movement).Error; err != nil {
                        return fmt.Errorf("failed to record stock movement: %v", err)
                    }
                }
            }

            if err := tx.Where("order_id = ?", id).Delete(&models.OrderItem{}).Error; err != nil {
                return err
            }
            if err := tx.Delete(&models.Order{}, id).Error; err != nil {
                return err
            }
            return nil
        })
    }

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete order: " + err.Error()})
        return
    }

    userID, _ := GetUintFromContext(c, "userID")
    h.LogService.Record(userID, "DELETE", "Order", strconv.Itoa(int(id)), fmt.Sprintf("Deleted order #%d (stock %s)", id, map[bool]string{true: "restored", false: "no change"}[order.Status == "completed"]), c.ClientIP())

    c.JSON(http.StatusOK, gin.H{"message": "Order deleted successfully"})
}
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/handlers/order_handler.go
git commit -m "fix: restore stock when deleting completed orders"
```

---

### Task 1.3: Fix Auth Context Race Condition (Frontend)

**Files:**
- Modify: `frontend/src/context/AuthContext.tsx`

- [ ] **Step 1: Read the current AuthContext.tsx**

Run: `read frontend/src/context/AuthContext.tsx`

- [ ] **Step 2: Fix the race condition by adding initialization check**

Replace the AuthProvider component with:

```tsx
export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(() => localStorage.getItem('token'));
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const initAuth = async () => {
      const savedToken = localStorage.getItem('token');
      if (!savedToken) {
        setIsLoading(false);
        return;
      }
      
      try {
        const response = await api.get('/api/auth/me');
        setUser((response.data as any).user);
      } catch {
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        setToken(null);
      } finally {
        setIsLoading(false);
      }
    };
    initAuth();
  }, []);

  const login = async (email: string, password: string) => {
    const response = await api.post('/api/auth/login', { email, password });
    const data = response.data as { token: string; user: User };
    if (!data || !data.token || !data.user) {
      throw new Error('Invalid login response');
    }
    localStorage.setItem('token', data.token);
    localStorage.setItem('user', JSON.stringify(data.user));
    setToken(data.token);
    setUser(data.user);
  };

  const register = async (name: string, email: string, password: string) => {
    const response = await api.post('/api/auth/register', { name, email, password });
    const data = response.data as { token: string; user: User };
    localStorage.setItem('token', data.token);
    localStorage.setItem('user', JSON.stringify(data.user));
    setToken(data.token);
    setUser(data.user);
  };

  const logout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    setToken(null);
    setUser(null);
  };

  // Use useMemo to prevent unnecessary re-renders
  const value = useMemo(() => ({
    user,
    token,
    isLoading,
    login,
    register,
    logout,
    isAuthenticated: !!token && !!user,
  }), [user, token, isLoading]);

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
}
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/context/AuthContext.tsx
git commit -m "fix: prevent auth race condition and unnecessary re-renders"
```

---

### Task 1.4: Fix ProtectedRoute Loading State

**Files:**
- Modify: `frontend/src/components/ProtectedRoute.tsx`

- [ ] **Step 1: Read the current ProtectedRoute.tsx**

Run: `read frontend/src/components/ProtectedRoute.tsx`

- [ ] **Step 2: Fix to show full-screen loader during auth check**

Replace with:

```tsx
import { Navigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';

interface ProtectedRouteProps {
  children: React.ReactNode;
  requiredRole?: string | string[];
}

export default function ProtectedRoute({ children, requiredRole }: ProtectedRouteProps) {
  const { isAuthenticated, isLoading, user } = useAuth();

  // Show full-screen spinner during authentication check
  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="flex flex-col items-center gap-4">
          <div className="w-10 h-10 border-4 border-indigo-200 border-t-indigo-600 rounded-full animate-spin" />
          <span className="text-sm font-medium text-gray-500">Loading...</span>
        </div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  if (requiredRole) {
    const roles = Array.isArray(requiredRole) ? requiredRole : [requiredRole];
    const hasAccess = user?.role === 'super_admin' || roles.includes(user?.role || '');
    
    if (!hasAccess) {
      return <Navigate to="/dashboard" replace />;
    }
  }

  return <>{children}</>;
}
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/ProtectedRoute.tsx
git commit -m "fix: show full-screen loader during auth check"
```

---

### Task 1.5: Add JWT Secret Validation on Startup

**Files:**
- Create: `backend/internal/config/validator.go`
- Modify: `backend/cmd/server/main.go`

- [ ] **Step 1: Read current config setup**

Run: `glob backend/internal/config/*.go`

- [ ] **Step 2: Create validator.go**

Create `backend/internal/config/validator.go`:

```go
package config

import (
    "fmt"
    "os"
)

func Validate(cfg *Config) error {
    if cfg.JWTSecret == "" {
        return fmt.Errorf("JWT_SECRET environment variable is required")
    }
    
    if len(cfg.JWTSecret) < 32 {
        return fmt.Errorf("JWT_SECRET must be at least 32 characters for security")
    }
    
    if cfg.DBHost == "" {
        return fmt.Errorf("DB_HOST environment variable is required")
    }
    
    return nil
}

func MustValidate(cfg *Config) {
    if err := Validate(cfg); err != nil {
        fmt.Fprintf(os.Stderr, "Configuration validation failed: %v\n", err)
        os.Exit(1)
    }
}
```

- [ ] **Step 3: Read main.go to find where config is loaded**

Run: `read backend/cmd/server/main.go`

- [ ] **Step 4: Add validation call in main.go**

Add after config loading:
```go
config.MustValidate(cfg)
```

- [ ] **Step 5: Commit**

```bash
git add backend/internal/config/validator.go backend/cmd/server/main.go
git commit -m "fix: add JWT secret validation on startup"
```

---

### Task 1.6: Add Input Validation for Guest Checkout

**Files:**
- Modify: `frontend/src/pages/POS.tsx`

- [ ] **Step 1: Read POS.tsx lines 470-490**

Run: `read frontend/src/pages/POS.tsx:470-490`

- [ ] **Step 2: Add validation state and handlers**

Add these states after the existing useState declarations:
```tsx
const [guestNameError, setGuestNameError] = useState('');
const [guestPhoneError, setGuestPhoneError] = useState('');
```

Add validation function:
```tsx
const validateGuestDetails = (): boolean => {
  let isValid = true;
  
  if (!customerId) {
    if (!guestName.trim()) {
      setGuestNameError('Guest name is required');
      isValid = false;
    } else if (guestName.trim().length < 2) {
      setGuestNameError('Name must be at least 2 characters');
      isValid = false;
    } else if (!/^[A-Za-z\s]{2,50}$/.test(guestName.trim())) {
      setGuestNameError('Name can only contain letters and spaces');
      isValid = false;
    } else {
      setGuestNameError('');
    }

    if (!guestPhone.trim()) {
      setGuestPhoneError('Contact number is required');
      isValid = false;
    } else if (!/^09\d{9}$/.test(guestPhone.trim())) {
      setGuestPhoneError('Invalid phone format (e.g., 09123456789)');
      isValid = false;
    } else {
      setGuestPhoneError('');
    }
  }
  
  return isValid;
};
```

- [ ] **Step 3: Update FormField components to show errors**

Replace the guest name field:
```tsx
<FormField 
  label="Guest Name" 
  value={guestName} 
  onChange={(v) => { setGuestName(v); setGuestNameError(''); }} 
  placeholder="Enter full name"
  error={guestNameError}
/>
```

Replace the guest phone field:
```tsx
<FormField 
  label="Contact" 
  value={guestPhone} 
  onChange={(v) => { setGuestPhone(v); setGuestPhoneError(''); }} 
  placeholder="09XX XXX XXXX"
  error={guestPhoneError}
/>
```

- [ ] **Step 4: Update handleCheckout to validate**

At the start of handleCheckout function, add:
```tsx
if (!validateGuestDetails()) {
  showToast('Please fix the guest details errors', 'error');
  return;
}
```

- [ ] **Step 5: Check if FormField supports error prop**

Run: `read frontend/src/components/FormField.tsx`

If it doesn't support error prop, update FormField to accept and display error:
```tsx
interface FormFieldProps {
  label: string;
  value: string;
  onChange: (value: string) => void;
  type?: 'text' | 'number' | 'email' | 'select' | 'password';
  placeholder?: string;
  options?: { value: string | number; label: string }[];
  required?: boolean;
  disabled?: boolean;
  error?: string; // Add this
}

export default function FormField({ label, value, onChange, type = 'text', placeholder, options, required, disabled, error }: FormFieldProps) {
  // ... existing code ...
  
  // Add error display before the closing tag
  {error && <p className="mt-1 text-xs text-red-500">{error}</p>}
}
```

- [ ] **Step 6: Commit**

```bash
git add frontend/src/pages/POS.tsx frontend/src/components/FormField.tsx
git commit -m "fix: add input validation for guest checkout"
```

---

## PHASE 2: HIGH PRIORITY PERFORMANCE FIXES

### Task 2.1: Add Debounce to Search Input

**Files:**
- Create: `frontend/src/hooks/useDebounce.ts`
- Modify: `frontend/src/pages/POS.tsx`

- [ ] **Step 1: Create useDebounce hook**

Create `frontend/src/hooks/useDebounce.ts`:

```tsx
import { useState, useEffect } from 'react';

export function useDebounce<T>(value: T, delay: number): T {
  const [debouncedValue, setDebouncedValue] = useState<T>(value);

  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedValue(value);
    }, delay);

    return () => {
      clearTimeout(handler);
    };
  }, [value, delay]);

  return debouncedValue;
}
```

- [ ] **Step 2: Update POS.tsx to use debounce**

Add import:
```tsx
import { useDebounce } from '../hooks/useDebounce';
```

Add after useState declarations:
```tsx
const debouncedSearch = useDebounce(search, 300);
```

Update filteredProducts to use debouncedSearch:
```tsx
const filteredProducts = products.filter(p => {
  const matchesSearch = p.name?.toLowerCase()?.includes(debouncedSearch.toLowerCase()) || false;
  const matchesCategory = selectedCategory ? p.category_id === selectedCategory : true;
  return matchesSearch && matchesCategory;
});
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/hooks/useDebounce.ts frontend/src/pages/POS.tsx
git commit -m "perf: add debounce to search input"
```

---

### Task 2.2: Add AbortController for Data Fetching

**Files:**
- Modify: `frontend/src/pages/POS.tsx`

- [ ] **Step 1: Update fetchData to accept signal**

Replace the fetchData function:
```tsx
const fetchData = async (signal?: AbortSignal) => {
  setLoading(true);
  setError(null);
  try {
    const [pRes, cRes, custRes, settingsRes] = await Promise.all([
      api.get('/api/products?all=1'),
      api.get('/api/categories'),
      api.get('/api/customers'),
      api.get('/api/settings'),
    ]);
    
    if (signal?.aborted) return;
    
    setProducts(pRes.data.products || []);
    setCategories(cRes.data.categories || []);
    setCustomers(custRes.data.customers || []);
    
    if (settingsRes.data?.service_advisors) {
      try {
        const parsed = Array.isArray(settingsRes.data.service_advisors)
          ? settingsRes.data.service_advisors
          : JSON.parse(settingsRes.data.service_advisors);
        setServiceAdvisors(parsed);
      } catch (e) {
        console.error("Failed to parse SAs", e);
      }
    }
  } catch (err) {
    if (signal?.aborted) return;
    console.error('POS data fetch failed', err);
    setError('Failed to sync with inventory system. Please check your connection.');
  } finally {
    if (!signal?.aborted) {
      setLoading(false);
    }
  }
};
```

- [ ] **Step 2: Update useEffect to use AbortController**

Replace the useEffect:
```tsx
useEffect(() => {
  const controller = new AbortController();
  fetchData(controller.signal);
  return () => controller.abort();
}, []);
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/pages/POS.tsx
git commit -m "fix: add AbortController to prevent memory leaks"
```

---

### Task 2.3: Fix Discount Validation

**Files:**
- Modify: `frontend/src/pages/POS.tsx`

- [ ] **Step 1: Add validation for discount input**

Add after useState declarations:
```tsx
const [discountError, setDiscountError] = useState('');
```

Add validation function:
```tsx
const validateDiscount = (value: string): boolean => {
  const num = parseFloat(value);
  if (isNaN(num) || num < 0) {
    setDiscountError('Discount must be a positive number');
    return false;
  }
  if (num > subtotal) {
    setDiscountError('Discount cannot exceed subtotal');
    return false;
  }
  setDiscountError('');
  return true;
};
```

Update the discount input onChange:
```tsx
onChange={(e) => {
  setDiscount(e.target.value);
  if (e.target.value) validateDiscount(e.target.value);
}}
```

- [ ] **Step 2: Update handleCheckout to validate discount**

Add at the start of handleCheckout:
```tsx
if (discount && !validateDiscount(discount)) {
  showToast('Please enter a valid discount amount', 'error');
  return;
}
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/pages/POS.tsx
git commit -m "fix: add discount validation"
```

---

### Task 2.4: Fix Hardcoded API URL

**Files:**
- Modify: `frontend/src/api/axios.ts`

- [ ] **Step 1: Read current axios.ts**

Run: `read frontend/src/api/axios.ts`

- [ ] **Step 2: Update to use environment variable**

Replace line 3:
```tsx
const API_BASE = import.meta.env.VITE_API_BASE_URL || 'http://168.144.46.137:8080';
```

- [ ] **Step 3: Create .env file with the variable**

Run: `read frontend/.env.production.example`

Add to `.env`:
```
VITE_API_BASE_URL=http://168.144.46.137:8080
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/api/axios.ts frontend/.env
git commit -m "fix: use environment variable for API base URL"
```

---

## PHASE 3: CODE ORGANIZATION IMPROVEMENTS

### Task 3.1: Create Storage Utility

**Files:**
- Create: `frontend/src/utils/storage.ts`
- Modify: `frontend/src/context/AuthContext.tsx`

- [ ] **Step 1: Create storage utility**

Create `frontend/src/utils/storage.ts`:

```tsx
const STORAGE_KEYS = {
  TOKEN: 'token',
  USER: 'user',
} as const;

export const storage = {
  getToken: (): string | null => {
    return localStorage.getItem(STORAGE_KEYS.TOKEN);
  },
  
  setToken: (token: string): void => {
    localStorage.setItem(STORAGE_KEYS.TOKEN, token);
  },
  
  removeToken: (): void => {
    localStorage.removeItem(STORAGE_KEYS.TOKEN);
  },
  
  getUser: <T>(): T | null => {
    const user = localStorage.getItem(STORAGE_KEYS.USER);
    return user ? JSON.parse(user) : null;
  },
  
  setUser: <T>(user: T): void => {
    localStorage.setItem(STORAGE_KEYS.USER, JSON.stringify(user));
  },
  
  removeUser: (): void => {
    localStorage.removeItem(STORAGE_KEYS.USER);
  },
  
  clearAuth: (): void => {
    storage.removeToken();
    storage.removeUser();
  },
};
```

- [ ] **Step 2: Update AuthContext to use storage utility**

Replace localStorage calls:
```tsx
const [token, setToken] = useState<string | null>(() => storage.getToken());

// In initAuth:
const savedToken = storage.getToken();
if (!savedToken) {
  setIsLoading(false);
  return;
}

// In catch block:
storage.clearAuth();

// In login:
storage.setToken(data.token);
storage.setUser(data.user);

// In logout:
storage.clearAuth();
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/utils/storage.ts frontend/src/context/AuthContext.tsx
git commit -m "refactor: create storage utility abstraction"
```

---

### Task 3.2: Create Error Types

**Files:**
- Create: `frontend/src/types/errors.ts`
- Modify: `frontend/src/api/axios.ts`

- [ ] **Step 1: Create error types**

Create `frontend/src/types/errors.ts`:

```tsx
export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public code?: string,
    public details?: unknown
  ) {
    super(message);
    this.name = 'ApiError';
  }
  
  static fromResponse(data: unknown, status: number): ApiError {
    const errorData = data as { error?: string; details?: unknown };
    return new ApiError(
      errorData.error || 'An error occurred',
      status,
      undefined,
      errorData.details
    );
  }
}

export class ValidationError extends ApiError {
  constructor(message: string, public fieldErrors?: Record<string, string>) {
    super(message, 400, 'VALIDATION_ERROR', fieldErrors);
    this.name = 'ValidationError';
  }
}

export class AuthError extends Error {
  constructor(message = 'Authentication failed') {
    super(message);
    this.name = 'AuthError';
  }
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/types/errors.ts
git commit -m "refactor: create error types"
```

---

## PLAN SUMMARY

| Phase | Tasks | Priority |
|-------|-------|----------|
| Phase 1 | 1.1-1.6 | CRITICAL - Data integrity & security |
| Phase 2 | 2.1-2.4 | HIGH - Performance & input validation |
| Phase 3 | 3.1-3.2 | MEDIUM - Code organization |

**Estimated total tasks:** 12 main tasks with ~40 steps

**Before implementing:** Run full test suite to establish baseline:
```bash
cd frontend && npm run build
cd backend && go build ./...
```

---

**Plan complete and saved to `docs/superpowers/plans/2026-03-24-code-audit-critical-fixes.md`**

Two execution options:

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

Which approach?
