# Observability Overlay (Phase 5)

Lightweight Grafana stack: **Prometheus** (metrics), **Loki** (logs), **Promtail**
(log shipping), **Grafana** (dashboards). See `docs/observability.md` for the design.

> Full URL reference (app + ops + monitoring): [`docs/local-urls.md`](../../docs/local-urls.md)

## Run

The monitoring stack is **included** in the main `docker-compose.yml` under the `monitoring` profile.

```bash
docker compose --profile monitoring up -d
```

Legacy (same result):

```bash
docker compose -f docker-compose.yml -f infra/monitoring/docker-compose.monitoring.yml up -d
```

Always-on monitoring: add `COMPOSE_PROFILES=monitoring` to `.env`.

| Service | URL | Notes |
|---|---|---|
| Grafana | http://localhost:3001 | Login: `GRAFANA_USER` / `GRAFANA_PASSWORD` from `.env` |
| Prometheus | http://localhost:9090 | scrapes `backend:8080/metrics` |
| Loki | http://localhost:3100 | log store (queried via Grafana) |
| Raw app metrics | http://localhost/metrics | plain text — use Grafana for graphs |

The **Meeting Platform — Overview** dashboard is auto-provisioned (active meetings,
WS connections, request rate/latency, errors, transcription latency, backend logs).

## What's wired
- Backend exposes Prometheus metrics at `/metrics` (already part of the main stack).
- Promtail tails all container stdout/stderr via the Docker socket and labels lines by
  `service` (the backend emits structured JSON logs with `request_id` / `meeting_id`).
- Grafana datasources (`Prometheus`, `Loki`) and the dashboard are provisioned on start.

## Notes
- Plain `docker compose up` does **not** start Grafana/Prometheus — ports 3001/9090 stay closed.
- `/metrics` on port 80 is always available but is not a dashboard; open Grafana for charts.
- For tracing, set `OTEL_ENABLED=true` and add an OTLP collector here when needed
  (the backend already prepares correlation IDs for traces).
