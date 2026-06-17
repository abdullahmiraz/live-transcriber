# Real-time AI Meeting Platform

A Google-Meet-inspired platform: create a meeting, share a link, join from the browser
with video/audio, **realtime text chat**, **live speech-to-text**, and **live translated
captions**.

> MVP-first, built on a foundation that scales. See `docs/` for the full design.

## Stack
- **Frontend:** SvelteKit + TypeScript, **Tailwind CSS v4 + shadcn-svelte** (WebRTC +
  WebSocket client, chat, captions UI)
- **Backend:** Go (clean architecture) — REST API + WebSocket signaling/events hub
- **Database:** PostgreSQL 16 (migrations, indexing) — source of truth
- **Realtime:** WebSockets + **Redis pub/sub** (chat fan-out, multi-instance ready)
- **Proxy:** nginx (single public entry, WebSocket upgrade, future TLS/LB)
- **AI:** pluggable STT + translation providers (`mock` by default, no keys needed)
- **Observability:** Prometheus + Grafana + Loki + OpenTelemetry (Phase 5)

## Quick start
```bash
cp .env.example .env
docker compose up --build
```
Then open http://localhost

- Frontend: `http://localhost/`
- API health: `http://localhost/healthz`
- Metrics: `http://localhost/metrics`

## Repository layout
```
frontend/   SvelteKit app
backend/    Go API + WS hub (cmd/, internal/, migrations/)
infra/      nginx config, monitoring overlay
docs/       architecture, roadmap, db, api, docker, observability, stt, project-memory
agents/     AI agent role instructions
skills/     reusable agent skills
```

## Development phases
See `docs/roadmap.md`. Phase 0 (planning) and Phase 1 (foundation) onward.

## Local development (without Docker)
- Backend: `cd backend && go run ./cmd/server` (needs a reachable Postgres + `DATABASE_URL`).
  `REDIS_URL` is optional locally — if unset, an in-memory pub/sub broker is used so chat
  works on a single instance. Set `REDIS_URL` to use Redis (required for multi-instance).
- Frontend: `cd frontend && npm install && npm run dev` (proxies `/api` + `/ws` to `:8080`).

## Documentation map
| Topic | File |
|---|---|
| Architecture + diagrams | `docs/architecture.md` |
| Roadmap | `docs/roadmap.md` |
| Database schema | `docs/database-design.md` |
| API + WS contracts | `docs/api-design.md` |
| Docker | `docs/docker-architecture.md` |
| Observability | `docs/observability.md` |
| STT/translation decision | `docs/stt-decision.md` |
| Project memory (decisions/status) | `docs/project-memory.md` |
