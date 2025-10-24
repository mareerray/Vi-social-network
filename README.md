# Social Network (minimal)

This repository is a small social network demo with a Go backend and a Vue frontend.

Quick start (dev)

1. Choose the DB file to use (recommended: backend/socialnetwork.db)

PowerShell:

```powershell
$env:DB_PATH = 'C:\Users\techgirl\OneDrive\Desktop\social-network\backend\socialnetwork.db'
cd 'C:\Users\techgirl\OneDrive\Desktop\social-network'
go run ./backend
```

2. Start frontend (optional for dev)

```powershell
cd frontend
npm install
npm run dev
```

3. Smoke test (requires backend running)

```powershell
.\scripts\smoke_test.ps1
```

Build for production

- Build frontend and place the `dist` folder under `frontend/dist`. The backend will serve `frontend/dist` if present; otherwise it serves `frontend/public`.

Notes

- The backend reads DB path from `DB_PATH` environment variable. If not set it defaults to `./backend/socialnetwork.db`.
- Session cleanup runs every 10 minutes in background.
- For a minimal demo, keep the DB under `backend/` to avoid duplicate files.
