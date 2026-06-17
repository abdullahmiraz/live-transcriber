# Observability Overlay (Phase 5)

Lightweight Grafana stack: **Prometheus** (metrics), **Loki** (logs), **Promtail**
(log shipping), **Grafana** (dashboards). See `docs/observability.md` for the design.

## Run

```bash
docker compose -f docker-compose.yml -f infra/monitoring/docker-compose.monitoring.yml up -d --build
```

| Service | URL | Notes |
|---|---|---|
| Grafana | http://localhost:3001 | login `admin` / `admin` (override via `GRAFANA_USER`/`GRAFANA_PASSWORD`) |
| Prometheus | http://localhost:9090 | scrapes `backend:8080/metrics` |
| Loki | http://localhost:3100 | log store (queried via Grafana) |

The **Meeting Platform — Overview** dashboard is auto-provisioned (active meetings,
WS connections, request rate/latency, errors, transcription latency, backend logs).

## What's wired
- Backend exposes Prometheus metrics at `/metrics` (already part of the main stack).
- Promtail tails all container stdout/stderr via the Docker socket and labels lines by
  `service` (the backend emits structured JSON logs with `request_id` / `meeting_id`).
- Grafana datasources (`Prometheus`, `Loki`) and the dashboard are provisioned on start.

## Notes
- This overlay is optional and kept separate so the core MVP stays simple.
- For tracing, set `OTEL_ENABLED=true` and add an OTLP collector here when needed
  (the backend already prepares correlation IDs for traces).
