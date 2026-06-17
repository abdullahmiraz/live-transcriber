# Execution Roadmap — MVP

Work in phases. Do not jump ahead. Each phase has a clear "done" definition.

## Phase 0 — Planning (no code)
**Deliverables**
- [x] Architecture plan + system diagrams (`docs/architecture.md`)
- [x] MVP roadmap (this file)
- [x] Database schema design (`docs/database-design.md`)
- [x] API contracts + WS event schemas (`docs/api-design.md`)
- [x] Docker architecture (`docs/docker-architecture.md`)
- [x] Observability architecture (`docs/observability.md`)
- [x] STT provider decision (`docs/stt-decision.md`)
- [x] Agent roles (`agents/`)
- [x] Skills (`skills/`)
- [x] Project memory (`docs/project-memory.md`)

**Done when:** all planning docs reviewed and consistent.

## Phase 1 — Foundation
**Deliverables**
- Repo structure (`frontend/`, `backend/`, `infra/`, `docs/`, `agents/`, `skills/`)
- Docker: `docker-compose.yml`, per-service Dockerfiles
- Go server: clean architecture, config, `/healthz`, `/readyz`, structured logging
- PostgreSQL connection + migrations runner
- SvelteKit app skeleton
- Nginx reverse proxy

**Done when:** `docker compose up` brings up frontend + backend + postgres + nginx,
and `GET /healthz` + the SvelteKit landing page both respond through nginx.

## Phase 2 — Meeting System
**Deliverables**
- Create meeting (returns slug + join URL)
- Get meeting by slug
- Room management (in-memory registry of active rooms/participants)
- Frontend: create meeting + join flow + lobby

**Done when:** a user can create a meeting, share the URL, and another user can open it.

## Phase 3 — Realtime
**Deliverables**
- WebSocket hub with per-room fan-out
- Signaling relay (offer/answer/ICE)
- WebRTC mesh client (camera/mic, render remote tiles)
- Participant join/leave events

**Done when:** two browsers in the same room see and hear each other.

## Phase 4 — AI Features
**Deliverables**
- Audio chunk capture + upload over WS
- STT provider adapter (see `docs/stt-decision.md`)
- Translation provider adapter
- Live captions UI (original + translated)

**Done when:** spoken audio appears as live captions, optionally translated.

## Phase 5 — Hardening
**Deliverables**
- Prometheus metrics + Grafana dashboards + Loki logs
- OpenTelemetry tracing wiring
- Rate limiting / input validation / CORS hardening
- Basic tests (unit + integration where valuable)

**Done when:** metrics/logs visible in Grafana; key paths covered by tests.

## Feature: Realtime Chat (text-only, inside meetings) — DONE
Added after the core phases. Clean separation: PostgreSQL = source of truth, Redis = realtime.

**Deliverables**
- DB: `messages` table (migration 0002), indexed `(meeting_id, created_at)`, paginated.
- Backend: `pubsub.Broker` (Redis + in-memory), `chat` domain + Postgres repo, WS
  `chat.message`/`chat.new`, REST `GET /api/meetings/{slug}/messages` (keyset pagination).
- Realtime: WS → persist → Redis pub/sub (`room:{slug}`) → fan-out to all participants.
- Frontend: shadcn-svelte chat panel in a tabbed sidebar (Chat / Captions / People).
- Docker: `redis` service (private), `redis_data` + `postgres_data` volumes.

**Done when:** participants in a room exchange messages in realtime, history persists and
reloads. ✅ Verified e2e over Redis (Valkey).

## Priority Order (from spec)
1. Working MVP
2. Developer velocity
3. Clean architecture
4. Scalability path
5. Production readiness
