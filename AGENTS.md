# AGENTS.md - SMSytem Development Guide

This file provides context for agentic coding agents operating in this repository.

## Project Overview

SMSytem is a Tauri-based desktop application with:
- **Frontend**: React 19 + TypeScript + Vite + Tailwind CSS
- **Backend**: External REST API (defaults to `http://168.144.46.137:8080`)
- **Desktop**: Tauri 2.x with Rust backend

## Build Commands

### Frontend (React)

```bash
# Navigate to frontend directory
cd frontend

# Development server (port 5173)
npm run dev

# Production build
npm run build

# Lint code
npm run lint

# Preview production build
npm run preview
```

### Tauri Desktop App

```bash
cd frontend

# Development mode with Tauri
npm run tauri dev

# Build production executable
npm run tauri build
```

### Testing

```bash
# Run Playwright tests (root directory)
npx playwright test

# Run specific test file
npx playwright test tests/login.spec.ts

# Run tests in headed mode
npx playwright test --headed
```

## Code Style Guidelines

### TypeScript Configuration

- TypeScript 5.9 (strict mode via `tsconfig.app.json`)
- Use explicit types for function parameters and return types
- Use `type` for interfaces and unions, `interface` for object shapes
- Avoid `any` - use `unknown` when type is truly unknown

### ESLint Configuration

The project uses ESLint 9 with:
- `@eslint/js` (recommended)
- `typescript-eslint` (recommended)
- `eslint-plugin-react-hooks` (flat config recommended)
- `eslint-plugin-react-refresh`

Run `npm run lint` to check for issues. ESLint is configured to ignore `dist`, `src-tauri/target`, `.vite`, and `node_modules`.

### React Patterns

**Imports:**
```tsx
// React core imports first
import { useState, useEffect, type FormEvent } from 'react';

// Third-party imports (alphabetical)
import { Link, useNavigate } from 'react-router-dom';

// Local imports (relative paths)
import { useAuth } from '../hooks/useAuth';
import { useToast } from '../context/ToastContext';
import api from '../api/axios';

// Type imports last
import type { ReactNode } from 'react';
```

**Component Structure:**
```tsx
// 1. Imports
import { useState } from 'react';
import { SomeIcon } from 'lucide-react';

// 2. Type definitions (if needed)
interface Column<T> {
  key: string;
  label: string;
  render?: (item: T) => ReactNode;
}

// 3. Component definition
export default function MyComponent() {
  // 4. Hooks first
  const [state, setState] = useState('');
  
  // 5. Callbacks
  const handleClick = () => {};
  
  // 6. Render
  return <div>...</div>;
}
```

### Naming Conventions

- **Files**: PascalCase for components (`Login.tsx`, `DataTable.tsx`), camelCase for utilities (`axios.ts`)
- **Components**: PascalCase (`export default function Login()`)
- **Functions**: camelCase (`handleSubmit`, `checkHealthNative`)
- **Interfaces**: PascalCase (`interface DataTableProps<T>`)
- **Constants**: SCREAMING_SNAKE_CASE for config values

### Tailwind CSS

- Use utility classes for styling
- Consistent spacing: `px-4 py-2`, `gap-2`, `space-y-4`
- Use `animate-pulse` for loading skeletons
- Responsive classes: `md:`, `lg:` prefixes

### Error Handling

```tsx
// Always handle errors with proper typing
try {
  await login(email, password);
} catch (err: unknown) {
  const axiosError = err as { response?: { data?: { error?: string } } };
  showToast(axiosError.response?.data?.error || 'Login failed', 'error');
}
```

### API Integration

- Use the centralized `api` instance from `src/api/axios.ts`
- Add Bearer token via request interceptor
- Handle 401 errors (redirect to login)
- Use environment variable `VITE_API_BASE_URL` for backend URL

### State Management

- Use React Context for global state (`AuthContext`, `ToastContext`)
- Use TanStack Query (`@tanstack/react-query`) for server state
- Local state with `useState`, derived state with `useMemo`

## File Organization

```
frontend/
├── src/
│   ├── api/          # Axios configuration and API calls
│   ├── components/  # Reusable UI components
│   ├── context/     # React Context providers
│   ├── hooks/       # Custom React hooks
│   ├── pages/       # Page-level components
│   ├── App.tsx     # Main app component
│   └── main.tsx    # Entry point
├── src-tauri/       # Rust backend
│   ├── src/
│   │   ├── lib.rs
│   │   └── main.rs
│   └── Cargo.toml
└── package.json
```

## Key Dependencies

- **UI**: React 19, Lucide React (icons), Recharts (charts)
- **Data**: Axios, TanStack Query, XLSX (Excel export)
- **Routing**: React Router DOM 7
- **Security**: DOMPurify (HTML sanitization)
- **Tauri Plugins**: log, shell, http, updater

## Common Tasks

### Adding a New Page

1. Create `frontend/src/pages/PageName.tsx`
2. Add route in `App.tsx`: `<Route path="/page-name" element={<PageName />} />`
3. Add navigation link in `Layout.tsx`

### Adding a New API Endpoint

1. Add function to appropriate module in `frontend/src/api/`
2. Use the `api` instance with proper typing
3. Handle errors and return typed responses

### Running a Single Test

```bash
npx playwright test tests/filename.spec.ts
```

## Notes for Agents

- This is a Tauri desktop app - some features may require Tauri APIs
- The backend URL defaults to `http://168.144.46.137:8080` - check if backend is online
- Always run `npm run lint` before committing
- Use `npm run build` to verify TypeScript compiles correctly
