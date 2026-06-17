# Local URLs & Access Guide

> Where to go after you start the project — app, health checks, API, WebSocket, metrics,
> and optional monitoring. **Start here** if you just ran `docker compose up`.

## 1. Start the stack

```bash
cp .env.example .env
docker compose up --build
```

Wait until all services are healthy, then use the links below. **Only nginx is public**
(port 80). Postgres and Redis are internal to Docker and have no browser URL.

---

## 2. Docker Compose (recommended) — base: `http://localhost`

All traffic goes through **nginx on port 80**. Use these URLs in your browser or with
`curl`.

### Application (browser)

| What | URL | Notes |
|---|---|---|
| **Home / landing page** | [http://localhost/](http://localhost/) | Create a meeting or join with a code |
| **Meeting room** | [http://localhost/m/{slug}](http://localhost/m/example-slug) | Replace `{slug}` with your meeting code (e.g. `abc-defg-hij`) |
| **Example after create** | `http://localhost/m/abc-defg-hij` | Shown in the API response as `join_url` |

**Typical flow**

1. Open [http://localhost/](http://localhost/)
2. Enter your name → **Create meeting** (or paste a code → **Join meeting**)
3. On the lobby screen → **Join with camera & microphone** (browser permission prompt)
4. Share the meeting URL from the in-call **Copy link** button

### Health & readiness (ops)

| What | URL | Expected |
|---|---|---|
| **Liveness** | [http://localhost/healthz](http://localhost/healthz) | `200` → `{"status":"ok"}` |
| **Readiness** | [http://localhost/readyz](http://localhost/readyz) | `200` → `{"status":"ready"}` when DB is up; `503` if not |
| **Prometheus metrics** | [http://localhost/metrics](http://localhost/metrics) | Plain-text Prometheus exposition format |

Quick checks:

```bash
curl -s http://localhost/healthz
curl -s http://localhost/readyz
curl -s http://localhost/metrics | head
```

### REST API

Base path: **`http://localhost/api`**

| What | Method & URL | Notes |
|---|---|---|
| Create meeting | `POST http://localhost/api/meetings` | Body: `{"title":"…","host_name":"…"}` |
| Get meeting | `GET http://localhost/api/meetings/{slug}` | `404` if missing |
| End meeting | `POST http://localhost/api/meetings/{slug}/end` | Sets status to `ended` |
| Chat history | `GET http://localhost/api/meetings/{slug}/messages?limit=50` | Paginated; see `docs/api-design.md` |

Example:

```bash
curl -s -X POST http://localhost/api/meetings \
  -H 'Content-Type: application/json' \
  -d '{"title":"Demo","host_name":"Alex"}'
```

Full request/response shapes: **`docs/api-design.md`**.

### WebSocket (realtime)

| What | URL | Notes |
|---|---|---|
| **Signaling + events** | `ws://localhost/ws?meeting={slug}&name={displayName}` | Used by the frontend automatically when you join a room |

Events include presence, WebRTC signaling, chat (`chat.new`), captions
(`transcript.updated`, `translation.updated`). Contract: **`docs/api-design.md`**.

---

## 3. Optional: monitoring overlay (Grafana stack)

The core app does **not** require this. Enable when you want dashboards and log search.

```bash
docker compose -f docker-compose.yml -f infra/monitoring/docker-compose.monitoring.yml up -d --build
```

| What | URL | Default login |
|---|---|---|
| **Grafana** (dashboards) | [http://localhost:3001](http://localhost:3001) | `admin` / `admin` |
| **Prometheus** (raw metrics UI) | [http://localhost:9090](http://localhost:9090) | — |
| **Loki** (log API) | [http://localhost:3100](http://localhost:3100) | Query via Grafana, not meant for direct browsing |

Grafana loads the **Meeting Platform — Overview** dashboard automatically. Backend metrics
are still scraped from `backend:8080/metrics`; the app also exposes them at
[http://localhost/metrics](http://localhost/metrics) through nginx.

Details: **`infra/monitoring/README.md`** and **`docs/observability.md`**.

---

## 4. Local development (without Docker)

Use this when editing frontend/backend code with hot reload. You need **Postgres** (and
optionally **Redis**) running separately.

### Backend (Go)

```bash
cd backend
# Set DATABASE_URL (and optionally REDIS_URL) — see .env.example
go run ./cmd/server
```

| What | URL |
|---|---|
| API | `http://localhost:8080/api/...` |
| Health | `http://localhost:8080/healthz` |
| Readiness | `http://localhost:8080/readyz` |
| Metrics | `http://localhost:8080/metrics` |
| WebSocket | `ws://localhost:8080/ws?meeting={slug}&name={name}` |

If `REDIS_URL` is unset, chat uses an in-memory broker (single instance only).

### Frontend (Vite dev server)

```bash
cd frontend
npm install
npm run dev
```

| What | URL | Notes |
|---|---|---|
| **App** | [http://localhost:3000/](http://localhost:3000/) | Vite proxies `/api`, `/healthz`, `/ws` → `:8080` |
| **Meeting room** | `http://localhost:3000/m/{slug}` | Same as production path |
| API (via proxy) | `http://localhost:3000/api/...` | Requires backend on `:8080` |

`/readyz` and `/metrics` are **not** proxied by Vite — hit the backend directly on `:8080`
or use the Docker stack for ops URLs.

---

## 5. Internal services (not for browsers)

These run inside Docker and are **not** published to your host:

| Service | Internal address | Purpose |
|---|---|---|
| Postgres | `postgres:5432` | Database (source of truth) |
| Redis | `redis:6379` | Chat pub/sub fan-out |
| Frontend (SSR) | `frontend:3000` | SvelteKit app (via nginx only) |
| Backend | `backend:8080` | Go API + WS (via nginx only) |

Do not map these ports in `docker-compose.yml` for production-like local runs unless you
have a specific debugging need.

---

## 6. Port summary

| Port | Service (when running) | Public? |
|---|---|---|
| **80** | nginx → app + API + WS + health + metrics | **Yes** — main entry |
| **3000** | Vite dev **or** frontend container (internal in Docker) | Dev only (direct) |
| **8080** | Go backend (direct in local dev) | Dev only (direct) |
| **3001** | Grafana (monitoring overlay) | Optional overlay |
| **9090** | Prometheus (monitoring overlay) | Optional overlay |
| **3100** | Loki (monitoring overlay) | Optional overlay |
| 5432 | Postgres | **No** (Docker internal) |
| 6379 | Redis | **No** (Docker internal) |

---

## 7. Troubleshooting

| Symptom | Check |
|---|---|
| Blank page / connection refused | Is `docker compose up` running? Try [http://localhost/healthz](http://localhost/healthz) |
| `readyz` returns 503 | Postgres not ready — wait for healthcheck or inspect `docker compose logs backend` |
| API 404 on `/api/...` | Use `/api/meetings`, not `/meetings` (nginx routes `/api/` to backend) |
| WebSocket fails | Ensure URL is `ws://localhost/ws` (through nginx), not a stale port |
| Camera/mic prompt missing | Click **Join with camera & microphone** on the lobby screen (user gesture required) |
| Meeting page blank after Create | Hard-refresh after `docker compose up --build`; slug must come from URL (`page.params`), not load data |
| Camera blocked on IP address | Use **http://localhost** not `http://192.168.x.x` — browsers require HTTPS or localhost for getUserMedia |
| Grafana empty | Start the monitoring overlay; confirm backend metrics at [http://localhost/metrics](http://localhost/metrics) |

---

## 8. Verify after changes

Run the automated smoke test (requires `docker compose up` and Node.js):

```bash
node scripts/smoke-test.mjs http://localhost
```

Covers: `/healthz`, `/readyz`, `/metrics`, home page, `POST/GET /api/meetings`, chat history,
meeting room HTML + JS bundle (lobby UI), WebSocket `room.welcome`, chat send/receive + REST
persistence, and 404 handling.

Camera/microphone (`getUserMedia`) cannot be tested from the CLI — join a room in the browser
at `http://localhost/m/{slug}` and click **Join with camera & microphone**.

## 9. Related documentation

| Topic | File |
|---|---|
| API + WebSocket contracts | `docs/api-design.md` |
| Docker topology | `docs/docker-architecture.md` |
| Metrics & dashboards design | `docs/observability.md` |
| UI / design tokens | `docs/design-system.md` |
| Architecture overview | `docs/architecture.md` |
| Environment variables | `.env.example` |
