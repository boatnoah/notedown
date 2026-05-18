# Notedown

Feats:

- [x] Detect change to the editor.
- [x] Send raw array buffer over websocket.
  - [ ] Avoid redudant packets from being applied.
- [x] Go distributes it to all clients
- [x] On message convert array buffer to Uint8array
- [x] Apply changes
- [x] Implement awareness protocol
- [ ] when a new client joins figure out a way to render the state for the joined client
- [ ] create an enpoint that gives you all the connections from the connections map

Chores:

- [ ] fix write error

## Backend Development

Install [Air](https://github.com/cosmtrek/air) for hot reloading:

```
go install github.com/cosmtrek/air@latest
```

Run the API server with file watching from the repo root:

```
cd backend && air
```

Air reads `backend/.air.toml`, builds the binary into `backend/tmp/notedown`, and restarts automatically when files in `cmd`, `internal`, or `pkg` change.

### Required Environment Variables

Create a `.env` file inside `backend/` (loaded automatically) or export the following variables before running the server:

```
GOOGLE_CLIENT_ID=...
GOOGLE_CLIENT_SECRET=...
FRONTEND_URL=http://localhost:5173
AUTH_CALLBACK_URL=http://localhost:3000/auth/google/callback
SESSION_SECRET=some-long-random-string
```

Without `SESSION_SECRET`, the server falls back to an insecure default meant only for local testing, and OAuth session validation may fail after restarts.
