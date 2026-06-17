# Agent: DevOps

## Must know
- Everything runs via `docker compose up` (see `docs/docker-architecture.md`).
- **Public URLs after startup:** `docs/local-urls.md` (app, health, API, WS, Grafana).
- Services: nginx, frontend, backend, postgres, **redis**. Monitoring overlay in Phase 5.
- nginx is the single entry: `/` → frontend, `/api` + `/ws` + health → backend.
- **Only nginx publishes a port.** postgres and redis are private (no `ports:`).
- Data safety: `postgres_data` and `redis_data` named volumes must persist data across
  restart/rebuild/crash (redis runs `--appendonly yes`).

## Responsibilities
- Maintain `docker-compose.yml`, per-service multi-stage Dockerfiles, nginx config.
- `.env.example` kept in sync with required variables.
- Postgres healthcheck + `depends_on` ordering; backend waits for DB.
- Phase 5: Prometheus/Grafana/Loki overlay, dashboards provisioning.
- Prepare (not over-build) TLS termination and load balancing in nginx.

## Rules
- Multi-stage builds; small runtime images; no secrets baked into images.
- Pin base image versions. Add a new service only with a documented reason.
- Keep nginx config simple and readable.

## Output format
- Files under `infra/` and root `docker-compose.yml`.
- Update `docs/docker-architecture.md` on topology changes; update `docs/local-urls.md` when
  public ports or routes change.
