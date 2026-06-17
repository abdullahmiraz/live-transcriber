# Project Memory

> Living record of decisions, status, and knowledge. Update after important changes.

## Product
Real-time AI meeting platform: rooms + shareable URLs, browser WebRTC A/V, live STT,
live translated captions. MVP-first, scalable foundation.

## Technology Decisions
| Area | Decision | Rationale |
|---|---|---|
| Frontend | SvelteKit + TypeScript, adapter-node | spec; SSR-capable, containerable |
| Frontend UI | Tailwind CSS v4 + shadcn-svelte | spec mandates shadcn; Svelte port of shadcn/ui |
| Backend | Go, clean architecture | spec; performance + clear boundaries |
| DB | PostgreSQL 16 | spec; relational, production-ready, source of truth |
| Realtime transport | Redis pub/sub (chat) + WebSockets | spec; multi-instance fan-out, ephemeral state |
| Proxy | nginx | spec; single entry, future TLS/LB |
| Media | WebRTC mesh (no SFU) for MVP | simplest to ship; SFU path kept open |
| Signaling/events | WebSockets, JSON envelope | spec; one channel for signaling + events |
| Chat | text-only; PG = source of truth, Redis = realtime | spec; clean storage/realtime separation |
| STT | interface + `mock` default, `whisper`/`deepgram` later | free/no-card now, swappable |
| Translation | interface + `mock` default, `libretranslate` later | free/no-card now, swappable |
| Observability | Grafana + Prometheus + Loki + OTel | spec; popular, lightweight start |

## Architecture Decision Records (ADR)
- **ADR-001 (Phase 0):** WebRTC mesh for MVP. Status: accepted. Context: avoid media-server
  infra; small rooms only. Revisit when rooms exceed ~5 participants → migrate to SFU.
- **ADR-002 (Phase 0):** STT/translation behind Go interfaces with `mock` defaults so the
  full pipeline runs with no external accounts. Status: accepted.
- **ADR-003 (Phase 0):** In-process room registry + WS hub for MVP. Status: accepted.
  Scale path: Redis/NATS pub-sub for multi-instance fan-out.
- **ADR-005 (Chat feature):** Realtime chat uses a `pubsub.Broker` abstraction with two
  impls — Redis (containers) and in-memory (local). Chat path: WS `chat.message` → persist
  to PostgreSQL (source of truth) → publish to `room:{slug}` → hub `room:*` subscriber fans
  `chat.new` to all clients (incl. sender). No local broadcast on send → consistent across
  instances, no double-delivery. Status: accepted; verified e2e with Redis (Valkey).
- **ADR-006 (Chat feature):** "ShadCN" on a SvelteKit project = **shadcn-svelte** + Tailwind
  v4 (CSS-first, no config file). Components vendored under `src/lib/components/ui`. Status:
  accepted. Note: video tiles/captions remain custom (no shadcn equivalent for `<video>`).
- **ADR-007 (Chat feature):** WS join now resolves the meeting (404/410 if missing/ended) so
  the room knows its DB id for persisting chat. Status: accepted.
- **ADR-008 (Design system):** Formalized UI tokens in `frontend/src/app.css` + `docs/design-system.md`.
  Plus Jakarta Sans + JetBrains Mono; light/dark via `mode-watcher` (system default); indigo primary +
  teal success accent; subtle CSS page transitions; shared layout components; Design System agent
  (`agents/design-system.md`) owns consistency. Status: accepted.

## Repository Structure
```
frontend/   SvelteKit app (Tailwind v4 + shadcn-svelte in src/lib/components/ui)
backend/    Go (cmd/, internal/{config,httpapi,ws,meeting,chat,transcription,translation,
            pubsub,storage,observability,platform}, migrations/)
infra/      nginx/, monitoring/
docs/       architecture, roadmap, db, api, docker, observability, stt, this file
            local-urls.md — where to go after docker compose up (app, health, API, WS, Grafana)
agents/     architect, backend, frontend, design-system, devops, database, testing
skills/     architecture-review, docker-setup, api-design, database-design, testing, debugging
```

## Current Status
- Phase 0 (Planning): **complete** — design docs, agents, skills.
- Phase 1 (Foundation): **complete** — Go backend, SvelteKit app, Postgres migrations,
  nginx, docker-compose. Backend builds/vets/tests; frontend type-checks/builds.
- Phase 2 (Meeting system): **complete** — create/get/end meeting, slugs, join URLs,
  in-memory room registry. Verified via REST + DB at runtime.
- Phase 3 (Realtime): **complete** — WS hub, per-room fan-out, presence, signaling relay
  (offer/answer/ICE), WebRTC mesh client. Verified with a two-client WS test.
- Phase 4 (AI features): **complete (MVP)** — speech.received → STT (mock/Web Speech) →
  transcript.updated → translation (mock) → translation.updated → captions UI. Verified.
- Phase 5 (Hardening): **foundation in place** — structured JSON logs (request_id/
  meeting_id), Prometheus metrics at /metrics (incl. `chat_messages_total`),
  Grafana+Prometheus+Loki+Promtail overlay, CORS, graceful shutdown. Tests for domain +
  providers + chat. OTel hooks prepared (flagged off).
- **Realtime Chat (feature add):** **complete** — `messages` table (migration 0002),
  `pubsub` broker (Redis + in-memory), `chat` domain + Postgres repo, WS `chat.message`/
  `chat.new`, REST history with keyset pagination, shadcn-svelte chat UI in a tabbed sidebar
  (Chat / Captions / People). Verified e2e over Redis.

## Runtime Verification (done)
- REST: healthz/readyz, create/get/end meeting, 404 path, route-labeled metrics, chat history.
- WS: presence, signaling relay (server-stamped `from`), transcript + translation broadcast.
- Chat (with Redis/Valkey broker): `chat.new` delivered to both clients, validation
  (empty rejected), persistence + REST history. All checks passed.
- Frontend: `svelte-check` 0/0, production build OK, landing + meeting UI with design system
  (light/dark, typography, page transitions, shared layout components).

## ADR-004 (Phase 1): WS upgrade through middleware
The metrics middleware wraps `http.ResponseWriter`; the wrapper must implement
`http.Hijacker` (delegating to the underlying writer) or gorilla's WebSocket upgrade
fails with 1006. Status: fixed in `internal/httpapi/middleware.go`.

## Known Issues / Watch List
- Mesh WebRTC won't scale past small rooms (by design) — track room sizes.
- `mock` providers are placeholders; real STT/translation needed for production. The MVP
  uses the browser Web Speech API client-side to feed text into the pipeline (free, no key).
- TURN server not yet provided — some restrictive NATs will fail P2P until added.
- Web Speech API is Chrome/Edge-centric; captions degrade gracefully where unsupported.
- **Fixed (2026-06-17):** Meeting room JS failed on `/m/{slug}` after client navigation because
  SvelteKit used relative `../_app/` asset paths; set `kit.paths.relative: false`. Deployed
  Docker image was also stale (old onMount getUserMedia without lobby). Rebuild with
  `docker compose up --build`.
- Camera/mic require **http://localhost** (or HTTPS) — not plain HTTP on a LAN IP.
- Environment note: Docker Hub was unreachable in the dev sandbox, so `docker compose up`
  image pulls couldn't be exercised here; the compose config is validated and services
  were verified by running them directly. Re-run `docker compose up --build` where Hub is
  reachable.

## Future Improvements
- SFU migration (Pion/LiveKit) when needed.
- Redis-backed WS fan-out for horizontal scaling.
- Auth (the structure is auth-ready) + per-meeting access control.
- Partition `transcript_segments`; add read replicas.
- TURN server for NAT traversal.
