# WiFi Share - local instructions

Before working here, read `~/codex-workspace/AGENTS.md`.

## Project context

- Local-network file sharing with a Go backend and a React/Vite frontend under `web/`.
- Preserve path containment, authentication, safe media handling, and LAN-only assumptions at every filesystem or network boundary.
- `config.local.json`, shared data, passwords, and local paths remain untracked; update `config.example.json` only with safe placeholders.
- Do not edit generated executables or frontend build output manually.

## Checks

- Backend: `go test ./...`
- Frontend: from `web/`, run `npm run lint` and `npm run build`

Read only the relevant architecture/security document and AI Dev Team rule; do not preload all rules, SPEC files, or `LEARNING_LOG.md`.
