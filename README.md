# Real-time AI Meeting Platform

A Google-Meet-inspired platform: create a meeting, share a link, join from the browser
with video/audio, and get **live speech-to-text** and **live translated captions**.

> MVP-first, built on a foundation that scales. See `docs/` for the full design.

## Stack
- **Frontend:** SvelteKit + TypeScript (WebRTC + WebSocket client, captions UI)
- **Backend:** Go (clean architecture) — REST API + WebSocket signaling/events hub
- **Database:** PostgreSQL 16 (migrations, indexing)
- **Proxy:** nginx (single entry, WebSocket upgrade, future TLS/LB)
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
- Frontend: `cd frontend && npm install && npm run dev`.

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
