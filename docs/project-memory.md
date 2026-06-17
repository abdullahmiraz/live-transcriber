# Project Memory

> Living record of decisions, status, and knowledge. Update after important changes.

## Product
Real-time AI meeting platform: rooms + shareable URLs, browser WebRTC A/V, live STT,
live translated captions. MVP-first, scalable foundation.

## Technology Decisions
| Area | Decision | Rationale |
|---|---|---|
| Frontend | SvelteKit + TypeScript, adapter-node | spec; SSR-capable, containerable |
| Backend | Go, clean architecture | spec; performance + clear boundaries |
| DB | PostgreSQL 16 | spec; relational, production-ready |
| Proxy | nginx | spec; single entry, future TLS/LB |
| Media | WebRTC mesh (no SFU) for MVP | simplest to ship; SFU path kept open |
| Signaling/events | WebSockets, JSON envelope | spec; one channel for signaling + events |
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

## Repository Structure
```
frontend/   SvelteKit app
backend/    Go (cmd/, internal/, migrations/)
infra/      nginx/, monitoring/
docs/       architecture, roadmap, db, api, docker, observability, stt, this file
agents/     architect, backend, frontend, devops, database, testing
skills/     architecture-review, docker-setup, api-design, database-design, testing, debugging
```

## Current Status
- Phase 0 (Planning): **complete** — all design docs, agents, skills written.
- Phase 1 (Foundation): in progress.
- Phases 2–5: pending.

## Known Issues / Watch List
- Mesh WebRTC won't scale past small rooms (by design) — track room sizes.
- `mock` providers are placeholders; real STT/translation needed for production.
- TURN server not yet provided — some restrictive NATs will fail P2P until added.

## Future Improvements
- SFU migration (Pion/LiveKit) when needed.
- Redis-backed WS fan-out for horizontal scaling.
- Auth (the structure is auth-ready) + per-meeting access control.
- Partition `transcript_segments`; add read replicas.
- TURN server for NAT traversal.
