# Live Meet — Interview preparation

> **Repo:** `live-transcript` · **Target:** Middle-track portfolio depth (Yandex-tier and similar)

Project-oriented study guide: pitch, file map, answered Q&A, diagrams index, cram sheet, and study plan — all in one place.

---

## Table of contents

- [Quick reference (cram sheet)](#quick-reference-cram-sheet)
- [Official project docs](#official-project-docs)
- [3-day study plan](#3-day-study-plan)
- [1. 60-second pitch](#1-60-second-pitch)
- [2. Project map & diagrams](#2-project-map--diagrams)
- [3. System design — Q&A](#3-system-design--qa)
- [4. Backend (Go) — Q&A](#4-backend-go--qa)
- [5. Realtime, WebRTC, captions — Q&A](#5-realtime-webrtc-captions--qa)
- [6. Frontend (SvelteKit) — Q&A](#6-frontend-sveltekit--qa)
- [7. Data & API — Q&A](#7-data--api--qa)
- [8. Auth & security — Q&A](#8-auth--security--qa)
- [9. DevOps & quality — Q&A](#9-devops--quality--qa)
- [10. Algorithms / CS — Q&A](#10-algorithms--cs--qa)
- [11. Production readiness](#11-production-readiness)
- [12. Behavioral / ownership](#12-behavioral--ownership)
- [13. Self-assessment](#13-self-assessment)
- [14. Practice drills](#14-practice-drills)

---

## Quick reference (cram sheet)

**Pitch:** Browser meeting app (Meet-like): shareable URL, WebRTC A/V, realtime chat, live captions + translation. Stack: **SvelteKit**, **Go**, **Postgres**, **Redis**, **nginx**. MVP: **WebRTC mesh**, **browser Web Speech API** for STT, **mock** translation (`[lang] text`). `docker compose up` runs everything. **Not production-ready** without auth, TURN/SFU, real AI, multi-instance WS.

| Diagrams in repo | `docs/architecture.md` |
|------------------|------------------------|
| System overview | §3 mermaid |
| WebRTC signaling | §6 sequence |
| STT → translation | §7 flowchart |
| Chat fan-out | §7.5 flowchart |
| Scale path | §8 |

| Key files | Path |
|-----------|------|
| WS hub / speech / chat | `backend/internal/ws/hub.go`, `client.go` |
| Meeting / chat domain | `backend/internal/meeting/`, `chat/` |
| STT / translation | `backend/internal/transcription/`, `translation/` |
| Meeting UI | `frontend/src/routes/m/[slug]/+page.svelte` |
| WebRTC / speech / WS | `frontend/src/lib/realtime/webrtc.ts`, `speech.ts`, `signaling.ts` |

| Feature | Today | Planned (`docs/stt-decision.md`) |
|---------|-------|----------------------------------|
| STT | Browser Web Speech + mock echo | Whisper / Deepgram |
| Translation | Mock `[ru] text` | LibreTranslate |
| Auth | Slug URL only | `docs/architecture.md` §9 |
| Media | Mesh, STUN only | SFU + TURN |

**Top 3 tradeoffs:** mesh (simple, breaks ~5+ peers) · mock AI (zero keys) · slug URLs (easy share, no auth).

**Top 3 before production:** auth · TURN + SFU · LibreTranslate + server STT.

**URLs:** App `http://localhost/` · WS `/ws` · Metrics `/metrics` · Grafana `:3001` (`docker compose --profile monitoring up`).

**Bug stories:** SSR `data: [null,null]` → `page.params.slug` · mobile chat → WS reconnect + 4s poll + `visibilitychange` · phone camera → HTTPS on LAN (`scripts/generate-dev-certs.sh`).

**Trap question:** “Is STT server-side?” → **No today.** Browser Web Speech transcribes; server receives text, mock echo + mock translation, broadcasts.

---

## Official project docs

| Topic | Path |
|-------|------|
| Architecture + **mermaid diagrams** | [`docs/architecture.md`](../docs/architecture.md) |
| API + WebSocket contract | [`docs/api-design.md`](../docs/api-design.md) |
| Database schema | [`docs/database-design.md`](../docs/database-design.md) |
| STT + translation plan | [`docs/stt-decision.md`](../docs/stt-decision.md) |
| Docker / nginx | [`docs/docker-architecture.md`](../docs/docker-architecture.md) |
| Observability | [`docs/observability.md`](../docs/observability.md) |
| Decisions & status | [`docs/project-memory.md`](../docs/project-memory.md) |
| Roadmap | [`docs/roadmap.md`](../docs/roadmap.md) |

## 3-day study plan

**Day 1 (2–3 h):** Memorize §1 pitch. Draw system from memory; check `docs/architecture.md` §3. Read §3 Q&A aloud.

**Day 2 (2–3 h):** Trace chat + captions with cited code files. Answer §4–§6 without looking. Run `docker compose up` + `node scripts/smoke-test.mjs http://localhost`.

**Day 3 (2 h):** §8–§10 (honest gaps). §11 production table. Morning of interview: re-read [Quick reference](#quick-reference-cram-sheet) only.

**What interviewers want:** you own tradeoffs (mesh, mock AI, slugs); you know MVP vs production; you cite real files and `docs/` diagrams.

---

## 1. 60-second pitch

**Problem:** Remote meetings need video, chat, and accessible captions — especially across languages.

**User:** Small teams or friends who want a self-hostable Meet-like room with a shareable link, no account required for MVP.

**Solution:** Browser app + Go backend. Create meeting → share `/m/{slug}` → WebRTC mesh for A/V → WebSocket for signaling, chat, captions. Postgres stores meetings/messages; Redis pub/sub fans chat across backend instances.

**Stack rationale:** Go for concurrent WebSockets and a single deployable binary; SvelteKit for fast reactive UI; Postgres as source of truth; Redis for realtime fan-out; nginx as one front door.

**Honest status:** Demo/MVP-ready locally. STT is **browser Web Speech API** with server **mock** echo; translation is **mock** (`[lang] text`). No real auth, no TURN, mesh only.

**If I started again:** Add TURN stub and auth interfaces on day one; add integration tests for WS flows earlier; consider SFU interface alongside mesh from the start.

**Hardest bugs (have stories ready):**
- SvelteKit SSR returned `data: [null,null]` for slug → fixed by using `page.params.slug` instead of `load()` data.
- Mobile chat stopped when WebSocket suspended → reconnect + 4s polling + `visibilitychange` refresh.
- Phone camera on LAN → `getUserMedia` needs **secure context** (HTTPS); added dev certs + `docs/local-urls.md`.

---

## 2. Project map & diagrams

### Repository layout

```
live-transcript/
├── frontend/src/
│   ├── routes/              # home, lobby, /m/[slug] meeting room
│   └── lib/
│       ├── api.ts           # REST client
│       ├── realtime/        # signaling.ts, webrtc.ts, speech.ts
│       ├── meeting/         # types, constants, session, join-media
│       └── config/          # languages.json, app.ts
├── backend/
│   ├── cmd/server/          # Composition root (main.go)
│   ├── internal/
│   │   ├── httpapi/         # REST transport
│   │   ├── ws/              # Hub, client, signaling, speech, chat
│   │   ├── meeting/         # Domain: create/join/delete meeting
│   │   ├── chat/            # Domain: validate, persist, publish
│   │   ├── transcription/   # STT Provider interface + mock
│   │   ├── translation/     # Translation Provider interface + mock
│   │   ├── pubsub/          # Redis + in-memory broker
│   │   ├── storage/postgres/
│   │   └── observability/
│   └── migrations/
├── infra/nginx/             # Reverse proxy
├── infra/monitoring/        # Prometheus + Grafana (compose profile)
├── docs/                    # Architecture, API, diagrams
└── scripts/smoke-test.mjs   # E2E smoke against nginx
```

### Diagrams already in the repo (cite in interviews)

| Topic | Document | Section |
|-------|----------|---------|
| System overview | [`docs/architecture.md`](../docs/architecture.md) | §3 mermaid |
| WebRTC signaling | same | §6 sequence diagram |
| STT → translation → captions | same | §7 flowchart |
| Chat: Postgres + Redis fan-out | same | §7.5 flowchart |
| Scale path | same | §8 table |
| Auth readiness | same | §9 |
| Docker topology | [`docs/docker-architecture.md`](../docs/docker-architecture.md) | service diagram |
| STT/translation providers | [`docs/stt-decision.md`](../docs/stt-decision.md) | options + pipeline |

### Code paths to trace while studying

| Flow | Start here |
|------|------------|
| Create meeting | `backend/internal/httpapi/handlers.go` → `meeting/service.go` |
| WS join + fan-out | `backend/internal/ws/hub.go`, `room.go` |
| Chat message | `backend/internal/ws/client.go` `handleChat` → `chat/service.go` |
| Speech / captions | `frontend/src/lib/realtime/speech.ts` → `client.go` `handleSpeech` |
| WebRTC mesh | `frontend/src/lib/realtime/webrtc.ts` |
| Meeting UI | `frontend/src/routes/m/[slug]/+page.svelte` |

---

## 3. System design — Q&A

### Q: Draw the architecture from browser to database. What hits nginx?

**Answer:** Browser talks to **nginx** on port 80 (443 with dev certs). Nginx routes:
- `/` → SvelteKit frontend
- `/api/*` → Go HTTP API
- `/ws` → Go WebSocket (upgrade headers)
- `/metrics`, `/healthz`, `/readyz` → backend

Go API writes to **PostgreSQL**. WS hub uses **chat.Service** (persist + Redis publish) and **STT/translation providers**. **WebRTC media is P2P** — does not go through nginx.

**Diagram:** `docs/architecture.md` §3.

---

### Q: Why one WebSocket for signaling + chat + captions?

**Answer:** One TCP connection per client reduces overhead and simplifies mobile reconnect. Message types multiplexed in JSON envelopes (`type` in `backend/internal/ws/message.go`). Tradeoff: handler bug could affect connection — mitigated by per-message errors and separate domain services behind the hub.

---

### Q: Why Postgres + Redis? Why not only one?

**Answer:**
- **Postgres** = durable source of truth (meetings, chat history, transcripts).
- **Redis** = pub/sub for **multi-instance** chat fan-out (`room:{slug}` → `chat.new` on each node).

Only Redis loses history; only Postgres can't push realtime across instances.

**Diagram:** `docs/architecture.md` §7.5.

---

### Q: Multiple backend instances — what breaks today?

**Answer:** **Chat works multi-instance** via Redis — hub does not broadcast locally on send.

**Breaks today:** in-memory **room registry** (signaling/presence per process), **empty-room timer** per instance, need **sticky WS** on load balancer.

**Fix:** Sticky sessions short-term; Redis/NATS for presence long-term (`docs/architecture.md` §8).

---

### Q: WebRTC mesh — when does it fail? Migration path?

**Answer:** **O(n²)** connections/bandwidth. Fails ~**5+ participants** or strict NAT (needs **TURN**; we only have **STUN**).

**Migration:** SFU (Pion, LiveKit, mediasoup). Keep signaling schema abstract; swap `webrtc.ts` mesh for SFU upstream/downstream. `docs/architecture.md` §6, §8.

---

### Q: Scale to 1k meetings? 10k users in one room?

**Answer:** 1k meetings → horizontal API/WS, sticky sessions, PG pooling, TURN/SFU, externalized room registry. 10k in one room → mesh impossible; SFU + sharded fan-out + async captions with backpressure. MVP targets small rooms.

---

### Q: Single point of failure in Docker?

**Answer:** Single Postgres, single nginx, no Redis replica. Production needs PG HA, Redis cluster, multiple backend replicas.

---

### Q: Production deployment (K8s, regions, CDN)?

**Answer:** K8s Deployments + HPA. CDN for static assets only — WebRTC/WS not CDN-cacheable. Regional stacks with TURN near users; region-pinned meetings.

---

## 4. Backend (Go) — Q&A

### Q: Clean architecture in your repo?

**Answer:** **Transport** (`httpapi`, `ws`) → **domain** (`meeting`, `chat`, `transcription`, `translation`) via interfaces. Implementations in `storage/postgres`, `pubsub`, providers. Wired in `cmd/server/main.go`. Domain never imports HTTP/Redis.

**Diagram:** `docs/architecture.md` §5.

---

### Q: How does the WebSocket hub work?

**Answer:** `Hub` owns `map[slug]*Room`. `Register`/`Unregister` manage clients; `Broadcast` fans out JSON. Empty room → **10-minute timer** → `DeleteBySlug`. Timer cancelled on re-join (`hub.go`).

**Files:** `backend/internal/ws/hub.go`, `room.go`, `client.go`.

---

### Q: Why does the server stamp identity on messages?

**Answer:** **Chat:** `handleChat` uses `SenderID: c.id`, `SenderName: c.name` from WS join — payload is only `content`. **Signaling:** `relaySignal` sets `env.From = c.id` (*never trust client-supplied identity*).

---

### Q: Create meeting → join → chat message?

**Answer:**
1. **POST `/api/meetings`** → slug in Postgres.
2. **WS** `/ws?meeting={slug}&name=...` → `participant.joined`.
3. **`chat.message`** → validate → INSERT Postgres → PUBLISH `room:{slug}` → **`chat.new`** to all clients.
4. History: **GET `/api/meetings/{slug}/messages?cursor=...`**.

---

### Q: Empty room auto-delete — races? Timer leaks?

**Answer:** Timer on last `Unregister`; cancelled when someone re-joins. Room mutex prevents duplicate timers. Graceful shutdown should cancel timers.

---

### Q: Graceful shutdown with open WebSockets?

**Answer:** SIGTERM → stop accept → close clients → drain pumps with timeout → close DB. K8s `preStop` + grace period.

---

### Q: Why mock STT/translation? Add LibreTranslate without changing hub?

**Answer:** Zero API keys; full pipeline in CI. New adapter implements `translation.Provider`, register in `factory.go`, set `TRANSLATION_PROVIDER`. Hub only calls interfaces. `docs/stt-decision.md`.

---

### Q: Rate limiting on `chat.message` and `speech.received`?

**Answer:** Token bucket / sliding window per `(room, participantId)` in Redis. Reject with WS error `rate_limited` in `client.handle`.

---

### Q: Metrics — what do they tell you?

**Answer:** `meetings_active`, `ws_connections_active`, latency histograms. Prometheus scrapes `/metrics`; Grafana with `--profile monitoring`. `docs/observability.md`.

---

## 5. Realtime, WebRTC, captions — Q&A

### Q: WebRTC mesh signaling — who initiates?

**Answer:** On `participant.joined`, new peer creates **offer** → `signal.offer` → **answer** → ICE both ways → P2P media.

**Diagram:** `docs/architecture.md` §6. **Code:** `webrtc.ts`, `signaling.ts`.

---

### Q: STUN vs TURN?

**Answer:** **STUN** = discover public address. **TURN** = relay when P2P fails. We have STUN only — some users get one-way video.

---

### Q: Phone camera only on HTTPS on LAN?

**Answer:** `getUserMedia` requires **secure context**. HTTP on LAN IP blocked. Fix: dev certs + `https://<lan-ip>/`. `docs/local-urls.md`.

---

### Q: Captions pipeline — why not raw audio to server today?

**Answer:** Browser **Web Speech API** → text as `speech.received` (base64 in `audio` field) + `targetLang` → server **mock STT** echoes → **mock translation** → `transcript.updated` / `translation.updated` → UI.

**Why:** Zero GPU/keys for demo. **Planned:** Whisper/Deepgram per `docs/stt-decision.md`.

**Files:** `speech.ts`, `client.go` `handleSpeech`, `transcription/mock.go`.

---

### Q: Latency? Streaming STT (Deepgram)?

**Answer:** Today: browser STT + WS RTT. **Deepgram:** streaming partials ~100ms → new event types + backpressure.

---

### Q: Partial vs final transcripts?

**Answer:** Web Speech `interimResults` for UI gray line; server mock marks `IsFinal: true`. Production: `isFinal` + `seq` per participant.

---

### Q: Two users speak at once — caption ordering?

**Answer:** Order by `(timestamp, participantId, seq)` per speaker. Don't interleave two streams in one line.

---

### Q: “Is transcription server-side?”

**Answer (critical):** **No today** — browser-side STT. Server receives text, runs provider interface (mock), broadcasts. Server-side STT **planned** (Whisper/Deepgram).

---

## 6. Frontend (SvelteKit) — Q&A

### Q: Why `page.params.slug` instead of `load()` data?

**Answer:** SSR hydration bug: `load()` returned `data: [null,null]`. Route param is reliable. `frontend/src/routes/m/[slug]/+page.svelte`.

---

### Q: Mobile chat when WebSocket suspends?

**Answer:** `visibilitychange` reconnect + **REST poll** every 4s (`CHAT_POLL_INTERVAL_MS`). Merge by message id. `+page.svelte`, `meeting/constants.ts`.

---

### Q: Lobby join modes — getUserMedia and user gesture?

**Answer:** `join-media.ts`: mic/cam/both/none. `getUserMedia` in button click handler before navigate.

---

### Q: Theme in-call vs home?

**Answer:** `mode-watcher` + Tailwind dark. `docs/design-system.md`.

---

### Q: Why centralize config in `lib/meeting/`, `languages.json`?

**Answer:** Single source for langs, phases, WS types — no UI/API drift.

---

### Q: Test meeting UI without manual clicks?

**Answer:** Playwright with stubbed `getUserMedia`/`RTCPeerConnection`. Smoke: `scripts/smoke-test.mjs` (API/WS/chat, not WebRTC/speech).

---

## 7. Data & API — Q&A

### Q: Persisted vs ephemeral?

| Data | Storage |
|------|---------|
| Meetings, messages, transcripts | **PostgreSQL** |
| WS presence, room hub | **In-memory** per backend |
| Chat fan-out | **Redis pub/sub** |

**Schema:** `docs/database-design.md`, `backend/migrations/`.

---

### Q: Chat — why Postgres AND Redis?

**Answer:** Postgres = history + pagination. Redis = cross-instance realtime without DB polling.

---

### Q: Pagination — keyset vs offset?

**Answer:** **Keyset** cursor — stable under concurrent inserts. `docs/api-design.md`.

---

### Q: Redis down?

**Answer:** Dev: in-memory broker (`pubsub/memory.go`). Prod: messages persist but no cross-instance fan-out; `/readyz` should fail.

---

### Q: Meeting slug security?

**Answer:** Random slug, casual use only. Prod: password, expiry, auth-required join, rate limits.

---

## 8. Auth & security — Q&A

### Q: No auth today — add without rewrite?

**Answer:** `docs/architecture.md` §9: auth middleware → `user_id` in context; token on WS upgrade; optional `users` table. Slug URLs can stay with signed meeting tokens.

---

### Q: JWT vs session cookies vs meeting tokens?

**Answer:** Session cookie for web; JWT for API/mobile; short-lived meeting token scoped to `slug` + role.

---

### Q: Prevent joining if slug leaks?

**Answer:** Password, host approval, org auth, rate limits, audit log.

---

### Q: CORS `*` in dev — production?

**Answer:** Explicit allowlist in middleware / nginx.

---

### Q: WebSocket auth on upgrade?

**Answer:** Validate cookie or `?token=` before `readPump`. Bind `participantId` to token claims.

---

### Q: Can client spoof `from`?

**Answer:** **No.** Signaling: server sets `From`. Chat: identity from join handshake, stored in DB.

---

## 9. DevOps & quality — Q&A

### Q: `docker compose up` vs `--profile monitoring`?

**Answer:** Default: app stack. `--profile monitoring` adds Prometheus + Grafana. `docs/getting-started.md`.

---

### Q: `/metrics` vs Grafana?

**Answer:** `/metrics` = raw Prometheus text from Go. Grafana = dashboards on top.

---

### Q: `/healthz` vs `/readyz`?

**Answer:** Liveness vs readiness (DB/Redis ping).

---

### Q: Hot reload in containers?

**Answer:** Vite HMR via nginx; Air for Go. `docs/docker-architecture.md`.

---

### Q: Smoke test coverage?

**Answer:** health, ready, metrics, create meeting, WS welcome, chat round-trip + REST persistence, 404. **Not:** speech, WebRTC, camera.

---

### Q: Add CI?

**Answer:** `go test ./...`, `npm run check`, compose + smoke, optional Playwright.

---

## 10. Algorithms / CS — Q&A

### Q: Rate limiter for chat per user per room?

**Answer:** Sliding window or token bucket in Redis (`rate:{room}:{user}:{minute}`).

---

### Q: Consistent hashing for room → server?

**Answer:** Hash `slug` to ring; sticky LB or route WS to owner node.

---

### Q: Dedupe chat on client retry?

**Answer:** `clientMessageId` UUID + unique index `(room_id, client_message_id)`.

---

### Q: Caption order if `seq` resets?

**Answer:** Server-assigned seq or `(participantId, clientSeq, timestamp)`.

---

### Q: Bandwidth — mesh, N participants?

**Answer:** Each peer ≈ **(N-1) × bitrate** send + receive. N=5 at 1 Mbps ≈ 8–16 Mbps per client — mesh breaks quickly.

---

## 11. Production readiness

| Question | Weak | Strong |
|----------|------|--------|
| Ready for production? | “Works locally” | “MVP demo; needs auth, TURN, real AI, multi-instance WS, load tests” |
| Biggest risk? | “Bugs” | “Mesh + no TURN + mock AI + in-memory room registry” |
| Before 100 users? | “Deploy” | “TURN, SFU or cap size, real translation, alerts” |
| Why Go? | “I like Go” | “WS concurrency, single binary, low memory per conn” |

---

## 12. Behavioral / ownership

**Why mock providers first?** End-to-end pipeline without external deps; swap via env (`docs/stt-decision.md`).

**How document changes?** `docs/change-history/`, `docs/project-memory.md`, `agents/gitops.md`.

**Cut if 2 weeks left?** SFU, fancy UI — keep auth skeleton, TURN, LibreTranslate, smoke + metrics.

**Validate mobile on Wi‑Fi?** HTTPS certs, `https://<lan-ip>`, Android Chrome + iOS Safari.

---

## 13. Self-assessment

| If you can answer well… | Signal |
|-------------------------|--------|
| §1–§5 + honest gaps | Junior / strong intern |
| + scaling, Redis/PG, WS hub, security plan | **Middle** |
| + SFU, multi-instance, SLOs, load tests, auth design | Senior |

Strong **middle-track** portfolio if you explain mock vs real, mesh limits, production path.

---

## 14. Practice drills

1. **Whiteboard** (10 min): Draw `docs/architecture.md` §3 from memory; add P2P dashed line.
2. **Trace** (15 min): Chat path from keypress to other screen — name every file.
3. **Trap** (2 min): “Is STT server-side?” — answer §5.
4. **Deep dive** (5 min): Chat **or** captions message-by-message.
5. **Mock interview:** 10 random questions from §3–§9; then check [Quick reference](#quick-reference-cram-sheet).

---

*Aligned with repo: monitoring profile, translation target lang, centralized frontend config. Update when `docs/architecture.md` changes.*
