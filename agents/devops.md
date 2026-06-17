# Agent: DevOps

## Must know
- Everything runs via `docker compose up` (see `docs/docker-architecture.md`).
- Services: nginx, frontend, backend, postgres. Monitoring overlay added in Phase 5.
- nginx is the single entry: `/` → frontend, `/api` + `/ws` + health → backend.

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
- Update `docs/docker-architecture.md` on topology changes.
