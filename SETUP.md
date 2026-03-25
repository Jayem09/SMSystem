# SMSystem Setup Guide

## Prerequisites
- Node.js (v18+)
- Go (v1.22+)
- MySQL

---

## Quick Setup

### Backend

```bash
cd SMSystem/backend
cp .env.example .env
# Edit .env with your DB credentials
go mod download
go run cmd/server/main.go
```

Server runs on http://localhost:8080

### Frontend

```bash
cd SMSystem/frontend
npm install
npm run dev
```

App opens at http://localhost:5173

---

## Environment Variables

### backend/.env
```
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=yourpassword
DB_NAME=smsystem
SERVER_PORT=8080
JWT_SECRET=generate_a_secret_key
```

### frontend/.env (optional)
```
VITE_API_BASE_URL=http://localhost:8080
```

---

## Build Desktop App

```bash
cd frontend
npm run tauri build
```

---

## Troubleshooting

- If frontend shows no data, make sure backend is running
- Check backend terminal for errors
- Ensure MySQL is running and credentials are correct
