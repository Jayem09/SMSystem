# Project Context: SMSystem

## Repo layout
- `frontend/`: React + Vite + TypeScript SPA
- `frontend/src-tauri/`: Tauri 2.x desktop shell (Rust)
- `backend/`: Go + Gin API
- Root `package.json`: Playwright only

## Runtime entry points
- Backend: `backend/cmd/server/main.go`
- Frontend: `frontend/src/main.tsx`

## Typical local commands
- Frontend: `cd frontend && npm run dev`
- Backend: `cd backend && make run`

## Important implementation notes
- Backend migrations auto-run on startup.
- Frontend includes online + offline flows.
- SSE endpoint: `/api/events` (token passed via query string).
