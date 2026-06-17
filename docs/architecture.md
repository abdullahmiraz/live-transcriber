# Architecture — Real-time AI Meeting Platform

> Status: Phase 0 (Planning). This is the source-of-truth architecture document.
> Update `docs/project-memory.md` whenever a decision here changes.

## 1. Product Summary

A Google-Meet-inspired platform focused on:

- Creating meeting rooms with shareable URLs
- Browser-based video/audio calls (WebRTC)
- Live speech-to-text (STT)
- Live translated captions

The MVP must be runnable with a single command (`docker compose up`) and built on a
foundation that can scale toward a production-grade communication platform.

> **After startup:** open [http://localhost/](http://localhost/) for the app; see
> [`docs/local-urls.md`](local-urls.md) for health, API, WebSocket, and monitoring URLs.

## 2. Design Principles

1. **Working MVP first**, but no architectural dead-ends.
2. **Clean architecture** on the backend — clear package boundaries, no giant files.
3. **Provider independence** — STT and translation are swappable behind interfaces.
4. **Realtime-first** — WebSockets for signaling/events, WebRTC for media.
5. **Observability from day one** — structured logs, metrics, tracing hooks.
6. **Scale path, not scale now** — start with WebRTC mesh; design for SFU migration.

## 3. High-Level System Diagram

```mermaid
flowchart LR
    subgraph Browser["Browser (SvelteKit client)"]
        UI[Meeting UI]
        RTC[WebRTC peer connections]
        WS[WebSocket client]
        CAP[Captions renderer]
    end

    subgraph Edge["Nginx (reverse proxy)"]
        NX[TLS termination / routing]
    end

    subgraph Backend["Go backend (clean architecture)"]
        API[HTTP API server]
        SIG[WS signaling + events hub]
        STT[STT service interface]
        TR[Translation service interface]
    end

    DB[(PostgreSQL)]
    OBS[[Observability:\nPrometheus / Loki / Grafana / OTel]]

    UI --> NX
    WS --> NX
    NX --> API
    NX --> SIG
    API --> DB
    SIG --> DB
    RTC <-->|"media (P2P mesh)"| RTC
    SIG -. signaling .- RTC
    SIG --> STT --> TR --> CAP
    API --> OBS
    SIG --> OBS
```

## 4. Component Responsibilities

### Frontend (SvelteKit + TypeScript + Tailwind + shadcn-svelte)
- UI built with **Tailwind CSS v4** and **shadcn-svelte**; design tokens and patterns in
  `docs/design-system.md` (Plus Jakarta Sans, light/dark via `mode-watcher`).
- Shared layout: `BrandLogo`, `AppHeader`, `ThemeToggle`, `PageStage` in `lib/components/layout/`.
- Meeting lobby + room UI, device selection (camera/mic).
- WebRTC client: getUserMedia, peer connection lifecycle, mesh negotiation.
- WebSocket client: signaling + realtime events (participants, captions, chat).
- Captions display: original transcript + translated subtitles.
- **Realtime text chat** panel (history via REST, live via WebSocket).
- Captures audio and streams short chunks for transcription.

### Backend (Go)
- **HTTP API**: meetings CRUD, chat history (paginated), health, readiness, metrics.
- **WS hub**: per-room fan-out, signaling relay, chat + AI events.
- **Chat service**: validate → persist (PostgreSQL) → publish (Redis pub/sub).
- **STT service**: interface + pluggable provider adapters.
- **Translation service**: interface + pluggable provider adapters.
- **Auth-ready**: structure supports adding auth later (middleware, user context) without rewrites.

### Database (PostgreSQL) — source of truth
- Meetings, participants, transcript/translation segments, **chat messages** (permanent).
- Migrations, indexing, relational integrity.

### Redis — realtime layer (not permanent storage)
- Pub/sub fan-out for chat across backend instances; ephemeral connection/session state.
- Channel convention `room:{slug}`; the hub pattern-subscribes `room:*`.
- Locally (no Redis), an in-memory broker with the same interface is used.

### Nginx
- Reverse proxy for frontend + backend, WebSocket upgrade pass-through.
- SSL termination later, load-balancing preparation.

## 5. Backend Clean Architecture

```
backend/
  cmd/server/            # entrypoint (main.go)
  internal/
    config/              # env-driven config
    httpapi/             # router, handlers, middleware (transport layer)
    ws/                  # websocket hub, client, signaling, chat dispatch
    meeting/             # domain: service + repository interface
    chat/                # domain: chat service + repository interface
    transcription/       # STT provider interface + adapters
    translation/         # translation provider interface + adapters
    pubsub/              # broker interface + Redis + in-memory implementations
    storage/postgres/    # repository implementations + db pool
    observability/       # logger, metrics, tracing
    platform/            # shared utils (ids, httputil)
  migrations/            # SQL migrations
```

Dependency rule: `httpapi`/`ws` (transport) → domain services (`meeting`, `chat`,
`transcription`, `translation`) → repository/broker interfaces. Implementations
(`storage/postgres`, `pubsub`, provider adapters) are injected at the composition root
(`cmd/server`). Domain code never imports transport or concrete infrastructure.

## 6. Realtime Media (WebRTC) — MVP Decision

**Decision: WebRTC mesh (full P2P), no media server for MVP.**

- Each participant connects directly to every other participant.
- Signaling (SDP offer/answer, ICE candidates) relayed through the Go WS hub.
- STUN via public `stun:stun.l.google.com:19302`; TURN is a later add-on for NAT traversal.

Rationale: zero media-server infra, simplest to ship. Mesh is fine for small rooms
(~3–5 participants). Connection count grows O(n²), so we **prepare for SFU migration**
(e.g., Pion/LiveKit/mediasoup) by keeping signaling abstract and not assuming P2P in the
event schema.

```mermaid
sequenceDiagram
    participant A as Peer A
    participant S as Go WS Hub
    participant B as Peer B
    A->>S: join(meeting)
    S-->>B: participant.joined(A)
    B->>S: signal.offer(to=A, sdp)
    S-->>A: signal.offer(from=B, sdp)
    A->>S: signal.answer(to=B, sdp)
    S-->>B: signal.answer(from=A, sdp)
    A->>S: signal.ice(to=B, candidate)
    S-->>B: signal.ice(from=A, candidate)
    Note over A,B: Direct P2P media established
```

## 7. AI Pipeline (STT → Translation → Captions)

```mermaid
flowchart LR
    MIC[Mic audio chunk] --> WSU[WS: speech.received]
    WSU --> STTSVC[STT provider adapter]
    STTSVC --> TUP[transcript.updated]
    TUP --> TRSVC[Translation provider adapter]
    TRSVC --> TRUP[translation.updated]
    TRUP --> CAPUI[Captions UI]
```

- Audio chunks captured client-side, sent over WS (binary or base64) to backend.
- Backend forwards to STT adapter, emits `transcript.updated`.
- Transcript forwarded to translation adapter, emits `translation.updated`.
- Both providers sit behind interfaces (`transcription.Provider`, `translation.Provider`)
  so we can swap implementations without touching the hub.

## 7.5 Realtime Chat (text-only)

Clean separation: **PostgreSQL is the source of truth; Redis is the realtime transport.**

```mermaid
flowchart LR
    U[User types message] --> FE[SvelteKit chat panel]
    FE -->|WS chat.message| HUB[Go WS hub]
    HUB --> SVC[chat.Service]
    SVC -->|persist| PG[(PostgreSQL messages)]
    SVC -->|publish room:slug| RED[(Redis Pub/Sub)]
    RED --> SUB[Hub subscriber room:*]
    SUB -->|WS chat.new| FE2[All participants]
```

- A client sends `chat.message` over WS; the server stamps the sender identity.
- `chat.Service` validates (non-empty, ≤4000 chars), **persists to PostgreSQL**, then
  **publishes** the message to Redis channel `room:{slug}`.
- The hub pattern-subscribes `room:*`; on receipt it fans the `chat.new` envelope out to
  all locally-connected clients — including the sender's instance. No instance broadcasts
  locally on send, so delivery is consistent across multiple backend instances and there is
  no double-delivery.
- History loads via REST (`GET /api/meetings/{slug}/messages`, keyset paginated); live
  messages arrive via WebSocket. No attachments/files — text only.
- Locally (no Redis) an in-memory broker with the same `pubsub.Broker` interface is used,
  so `go run` works without Redis while containers use Redis.

## 8. Scale Considerations ("100k users")

| Concern | MVP approach | Scale path |
|---|---|---|
| Media | WebRTC mesh | SFU (Pion/LiveKit) |
| Chat fan-out | **Redis pub/sub** (`room:*`) | already multi-instance ready; NATS if needed |
| Signaling/presence fan-out | in-process room hubs | move onto Redis/NATS like chat |
| Sessions/state | in-memory room registry | external store (Redis) |
| DB | single Postgres | read replicas, partition transcripts + messages |
| STT cost/latency | single provider | provider routing + batching/streaming |
| Stateless API | yes | horizontal scaling behind nginx/LB |

## 9. Security & Auth Readiness

- Request context carries `request_id`, and (later) `user_id`.
- Middleware chain ready for auth injection.
- Meetings identified by unguessable slugs (not sequential IDs) in URLs.
- Input validation at transport boundary; CORS locked to known origins.

See `docs/database-design.md`, `docs/api-design.md`, `docs/docker-architecture.md`,
`docs/observability.md`, and `docs/stt-decision.md` for detailed sub-designs.
