# Docker Architecture

Everything runs through containers. `docker compose up` is the single entrypoint.

## Services (initial)

```mermaid
flowchart TB
    client[Browser] -->|"only public port :80"| nginx
    subgraph compose["private docker network"]
        nginx[nginx :80] --> frontend[frontend SvelteKit :3000]
        nginx --> backend[backend Go :8080]
        backend --> postgres[("postgres :5432 (private)")]
        backend --> redis[("redis :6379 (private)")]
    end
```

| Service | Image / build | Port | Public? | Purpose |
|---|---|---|---|---|
| `nginx` | `nginx:alpine` + config | 80 | **yes** | reverse proxy, WS upgrade, single entry |
| `frontend` | build `frontend/` | 3000 | no | SvelteKit (Node adapter) |
| `backend` | build `backend/` | 8080 | no | Go API + WS hub |
| `postgres` | `postgres:16-alpine` | 5432 | **no** | database (source of truth) |
| `redis` | `redis:7-alpine` | 6379 | **no** | realtime pub/sub + ephemeral state |

Security rule: **only nginx publishes a port.** `postgres` and `redis` have no `ports:`
mapping, so they are reachable only on the private compose network — never exposed publicly.

Monitoring services (Phase 5, separate `infra/monitoring/docker-compose.yml` overlay):
`prometheus`, `grafana`, `loki`, `promtail` / OTel collector.

## Routing (nginx)
- `/` → `frontend:3000`
- `/api/` → `backend:8080`
- `/ws` → `backend:8080` (with `Upgrade`/`Connection` headers for WebSocket)
- `/healthz`, `/readyz`, `/metrics` → `backend:8080`

## Build Strategy
- **Backend**: multi-stage build — `golang:1.26` builder → distroless/alpine runtime.
- **Frontend**: multi-stage — `node:26` builder → `node:26-alpine` runtime (adapter-node).
- Dev: bind mounts + hot reload optional later; MVP uses built images.

## Config & Secrets
- Environment via `.env` (compose interpolation) — sample provided as `.env.example`.
- Backend reads: `DATABASE_URL`, `REDIS_URL`, `PORT`, `LOG_LEVEL`, `CORS_ORIGINS`,
  `STT_PROVIDER`, `TRANSLATION_PROVIDER`, `DEFAULT_TARGET_LANG`, provider API keys.
- Postgres: `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`.

## Health & Ordering
- `postgres` has a healthcheck (`pg_isready`); `redis` has a healthcheck (`redis-cli ping`).
- `backend` `depends_on` both postgres and redis with `condition: service_healthy`.
- `backend` runs migrations on startup (idempotent) before serving.

## Volumes & Networks (data safety)
- Named volume `postgres_data` → `/var/lib/postgresql/data` (database persistence).
- Named volume `redis_data` → `/data` (redis runs with `--appendonly yes` for durability).
- Containers are disposable; **data survives restart, rebuild, and crash** via these volumes.
- Single user-defined bridge network (compose default) so services resolve by name and stay
  private (only nginx is published).

## Keep It Simple
- No Kubernetes, no service mesh for MVP.
- Add services only when a concrete need appears (e.g., TURN server for NAT traversal).
