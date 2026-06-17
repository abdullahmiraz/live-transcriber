# Real-time AI Meeting Platform

A Google-Meet-inspired platform: create a meeting, share a link, join from the browser
with video/audio, **realtime text chat**, **live speech-to-text**, and **live translated
captions**.

> MVP-first, built on a foundation that scales. See `docs/` for the full design.

## Stack
- **Frontend:** SvelteKit + TypeScript, **Tailwind CSS v4 + shadcn-svelte** (WebRTC +
  WebSocket client, chat, captions UI; light/dark themes, design system in `docs/design-system.md`)
- **Backend:** Go (clean architecture) — REST API + WebSocket signaling/events hub
- **Database:** PostgreSQL 16 (migrations, indexing) — source of truth
- **Realtime:** WebSockets + **Redis pub/sub** (chat fan-out, multi-instance ready)
- **Proxy:** nginx (single public entry, WebSocket upgrade, future TLS/LB)
- **AI:** pluggable STT + translation providers (`mock` by default, no keys needed)
- **Observability:** Prometheus + Grafana + Loki + OpenTelemetry (Phase 5)

## Quick start (dev — hot reload, no rebuild on code changes)

```bash
cp .env.example .env
bash scripts/generate-dev-certs.sh   # once — needed for phone camera over Wi‑Fi (HTTPS)
docker compose up --build            # first time only; after that: docker compose up
```

| Where | URL |
|---|---|
| **This PC** | [http://localhost/](http://localhost/) |
| **Phone / tablet (same Wi‑Fi)** | `https://<your-pc-lan-ip>/` (accept cert warning once) |
| **Production build (no hot reload)** | `docker compose -f docker-compose.prod.yml up --build` |

Code changes under `frontend/` and `backend/` reload automatically (Vite HMR + Air).

**Full URL guide** (Docker vs local dev, Grafana overlay, troubleshooting):
**[`docs/local-urls.md`](docs/local-urls.md)**

**Verify the stack** (API, WebSocket, chat, frontend bundles):

```bash
node scripts/smoke-test.mjs http://localhost
```

## Repository layout
```
frontend/   SvelteKit app
backend/    Go API + WS hub (cmd/, internal/, migrations/)
infra/      nginx config, monitoring overlay
docs/       architecture, roadmap, db, api, docker, local-urls, observability, stt, project-memory
agents/     AI agent role instructions
skills/     reusable agent skills
```

## Development phases
See `docs/roadmap.md`. Phase 0 (planning) and Phase 1 (foundation) onward.

## Local development (without Docker)
See **`docs/local-urls.md`** §4 for all dev URLs. Summary:

- Backend: `cd backend && go run ./cmd/server` → `http://localhost:8080` (needs Postgres + `DATABASE_URL`).
  `REDIS_URL` is optional locally — if unset, an in-memory pub/sub broker is used so chat
  works on a single instance. Set `REDIS_URL` to use Redis (required for multi-instance).
- Frontend: `cd frontend && npm install && npm run dev` → [http://localhost:3000](http://localhost:3000)
  (proxies `/api`, `/healthz`, `/ws` to `:8080`).

## Documentation map
| Topic | File |
|---|---|
| **Local URLs (start here after `docker compose up`)** | [`docs/local-urls.md`](docs/local-urls.md) |
| Architecture + diagrams | `docs/architecture.md` |
| Roadmap | `docs/roadmap.md` |
| Database schema | `docs/database-design.md` |
| API + WS contracts | `docs/api-design.md` |
| Docker | `docs/docker-architecture.md` |
| **Change history (rollback)** | `docs/change-history/INDEX.md` |
| Observability | `docs/observability.md` |
| STT/translation decision | `docs/stt-decision.md` |
| Design system (tokens, typography, motion) | `docs/design-system.md` |
| Project memory (decisions/status) | `docs/project-memory.md` |
