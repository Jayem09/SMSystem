# AGENTS.md - SMSytem Development Guide

This file provides context for agentic coding agents operating in this repository.

## Project Overview

SMSytem is a Tauri-based desktop application with:
- **Frontend**: React 19 + TypeScript + Vite + Tailwind CSS
- **Backend**: External REST API (defaults to `http://168.144.46.137:8080`)
- **Desktop**: Tauri 2.x with Rust backend
- **Testing**: Playwright (e2e), no unit tests currently

## Build Commands

### Frontend (React)

```bash
cd frontend

# Development server (port 5173)
npm run dev

# Type checking
npx tsc --noEmit

# Production build
npm run build

# Lint code
npm run lint
```

### Tauri Desktop App

```bash
cd frontend
npm run tauri dev     # Development mode (opens desktop window)
npm run tauri build   # Production executable
```

### Running Tests

```bash
# Playwright e2e tests (from root)
npx playwright test

# Run specific test file
npx playwright test tests/login.spec.ts

# Run with UI (headed mode)
npx playwright test --headed
```

## Code Style Guidelines

### TypeScript

- TypeScript ~5.9 with Vite configuration
- Relaxed strict mode (not fully strict - see `tsconfig.app.json`)
- Use `type` for unions, `interface` for objects
- Avoid `any` - use `unknown` for untyped catch blocks
- Use `erasableSyntaxOnly: true` (no private keyword, use `#` fields)

### Imports Order

```tsx
// 1. React → 2. Third-party → 3. Local → 4. Types
import { useState, type FormEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import api from '../api/axios';
import type { User } from '../types';
```

### Component Structure

```tsx
// Hooks → state → callbacks → render
export default function MyComponent() {
  const [state, setState] = useState('');
  const navigate = useNavigate();

  const handleClick = () => navigate('/dashboard');

  return <button onClick={handleClick}>Click</button>;
}
```

### Tailwind CSS

- Use utility classes: `flex`, `gap-4`, `p-4`, `rounded-lg`, `bg-white`
- Dark mode: use `dark:` prefix

### Naming Conventions

- Files: PascalCase components, camelCase utils
- Interfaces: PascalCase
- Constants: SCREAMING_SNAKE_CASE

### Error Handling

```tsx
try {
  await login(email, password);
} catch (err: unknown) {
  const axiosError = err as { response?: { data?: { error?: string } } };
  showToast(axiosError.response?.data?.error || 'Login failed', 'error');
}
```

### Tauri/Rust

Webview blocks JS fetch/XHR - always use Rust commands via `invoke`. Use `reqwest` in Rust. Commands go in `src-tauri/src/lib.rs`. See `frontend/src/api/axios.ts` for wrapper pattern.

### API Integration

Use `api` from `src/api/axios.ts` (pre-configured with interceptors). Bearer token via auth interceptor. Handle 401 (redirect to login). Backend URL via `VITE_API_BASE_URL`.

### State Management

- React Context: global state (AuthContext, ToastContext)
- TanStack Query: server state, caching
- `useState`: local state, `useMemo`: derived values

## File Organization

```
frontend/src/
├── api/         # Axios, API calls
├── components/ # Reusable UI
├── context/    # React Context
├── hooks/      # Custom hooks
├── pages/      # Page components
├── types/      # TypeScript interfaces
└── App.tsx     # Routes
```

## Key Dependencies

- **UI**: React 19, Lucide React (icons), Recharts
- **Data**: Axios, TanStack Query, XLSX
- **Routing**: React Router DOM 7
- **Tauri**: log, shell, http, updater plugins

## Common Tasks

### Adding a Page
1. Create `frontend/src/pages/PageName.tsx`
2. Add route in `App.tsx`
3. Add link in `Layout.tsx`

### Adding API Endpoint
1. Add function in `api/*.ts`
2. Use `api.post()` or `api.get()` with typing
3. Handle errors with try/catch

## Notes for Agents

- Backend defaults to `http://168.144.46.137:8080` - check if online
- Run `npm run lint` and `npx tsc --noEmit` before committing
- Use `showToast()` from ToastContext for user feedback
