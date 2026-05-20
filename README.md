# Notedown

A collaborative markdown editor. Multiple users edit the same document in real time over WebSockets, with live preview, cursor presence, and shareable links.

## Features

- **Real-time collaboration** — CodeMirror editor syncs insert/delete operations through a server-authoritative OT layer; clients receive versioned snapshots.
- **Live markdown preview** — Split editor and rendered preview pane.
- **Presence** — Remote cursors show each collaborator’s name, avatar, and selection.
- **Documents** — Create, list, and open documents; download as `.md`.
- **Sharing** — Per-document `private`, `read`, or `edit` modes; share links require sign-in; read-only guests cannot apply edits.
- **Auth** — Email/password registration and JWT sessions (short-lived access token + rotating refresh cookie).
- **Resilient connections** — WebSocket auto-reconnect with exponential backoff; buffered ops replay after reconnect.

## Architecture

| Layer        | Stack                                                                                         |
| ------------ | --------------------------------------------------------------------------------------------- |
| **Frontend** | React, Vite, TypeScript, TanStack Router, TanStack Query, Tailwind CSS, shadcn/ui, CodeMirror |
| **Backend**  | Go, chi, gorilla/websocket, Goose, sqlc, pgx                                                  |
| **Database** | PostgreSQL (Neon)                                                                             |
| **CI/CD**    | GitHub Actions; deploy to Render on merge to `main`                                           |

```
frontend/          React app (Vite)
  src/features/    auth, editor, documents
  src/lib/         config, typed WebSocket protocol
backend/
  cmd/api/         HTTP + WebSocket server
  internal/
    auth/          register, login, refresh, logout
    ot/          OT manager (authoritative document state)
    documents/     document service + repository interfaces
    realtime/      WebSocket hub, presence, typed protocol
    storage/
      postgres/    production adapters (sqlc)
      memory/      in-memory adapters (tests)
```

On connect, the server sends a document **snapshot** and a **presence snapshot**. Clients send `operation`, `sync`, and `presence` messages; the server broadcasts updated snapshots and presence updates to the room.

## Local development

### Prerequisites

- Go 1.24+
- Node.js (LTS)
- PostgreSQL (or a [Neon](https://neon.tech) database)
- [Air](https://github.com/air-verse/air) (optional, for backend hot reload)

### Backend

```bash
cp backend/.env.example backend/.env   # fill in values (see below)
cd backend && air                      # or: go run ./cmd/api
```

Air reads `backend/.air.toml`, builds into `backend/tmp/notedown`, and restarts on changes under `cmd/`, `internal/`, and `pkg/`.

The API listens on `:3000` by default (`HTTP_ADDR`).

### Frontend

```bash
cp frontend/.env.example frontend/.env
cd frontend && npm install && npm run dev
```

Vite serves the app at `http://localhost:5173`. Set `VITE_API_URL` if the backend is not on `localhost:3000`.

### Checks

From the repo root:

```bash
bin/check
```

Runs `go vet`, `go test`, `golangci-lint` (if installed), `tsc --noEmit`, and `vite build` — the same gates as CI.

## Environment variables

Copy the example files and adjust for your environment. Actual `.env` files are gitignored.

### Backend (`backend/.env`)

| Variable         | Description                                                           |
| ---------------- | --------------------------------------------------------------------- |
| `HTTP_ADDR`      | HTTP listen address (default `:3000`)                                 |
| `DATABASE_URL`   | PostgreSQL connection string (Neon)                                   |
| `JWT_SECRET`     | Signing key for access tokens                                         |
| `SESSION_SECRET` | Signing key for refresh-token cookies                                 |
| `FRONTEND_URL`   | Frontend origin for CORS and redirects (e.g. `http://localhost:5173`) |

Goose migrations run automatically on startup. The server fails fast if `DATABASE_URL` is missing.

### Frontend (`frontend/.env`)

| Variable       | Description                                        |
| -------------- | -------------------------------------------------- |
| `VITE_API_URL` | Backend HTTP origin (e.g. `http://localhost:3000`) |
| `VITE_WS_URL`  | Optional WebSocket origin override                 |

### CI / deployment (GitHub Actions secrets)

| Secret                   | Description                               |
| ------------------------ | ----------------------------------------- |
| `NEON_API_KEY`           | Neon API key for per-PR database branches |
| `NEON_PROJECT_ID`        | Neon project ID                           |
| `RENDER_DEPLOY_HOOK_URL` | Render deploy hook (CD on `main`)         |

## API overview

### REST

| Method | Path              | Description                            |
| ------ | ----------------- | -------------------------------------- |
| `GET`  | `/healthz`        | Health check                           |
| `POST` | `/auth/register`  | Create account                         |
| `POST` | `/auth/login`     | Issue access token + refresh cookie    |
| `POST` | `/auth/refresh`   | Rotate refresh token, new access token |
| `POST` | `/auth/logout`    | Clear session                          |
| `POST` | `/documents`      | Create document (authenticated)        |
| `GET`  | `/documents`      | List documents for current user        |
| `GET`  | `/documents/{id}` | Document snapshot                      |

### WebSocket (`/ws?documentId={id}`)

Requires a valid access token. Enforces document `share_mode` (`private` / `read` / `edit`).

**Client → server:** `operation`, `sync`, `presence`  
**Server → client:** `snapshot`, `presenceSnapshot`, `presenceUpdate`, `error`

## Routes (frontend)

| Path                | Description                              |
| ------------------- | ---------------------------------------- |
| `/login`            | Sign in                                  |
| `/register`         | Create account                           |
| `/documents`        | Document list (landing page after login) |
| `/editor?room={id}` | Collaborative editor                     |

Unauthenticated users hitting protected routes are redirected to `/login` with the return URL preserved.

## CI

On every push and PR to `main`, GitHub Actions runs:

- **pr-title** — Conventional Commits format
- **backend** — `go vet`, `go test`, golangci-lint; optional Neon branch per PR
- **frontend** — `tsc --noEmit`, `vite build`

Merges to `main` trigger a Render deploy after CI passes.
