# Observability Architecture

Stack: **Grafana ecosystem** (popular and well-supported, incl. in Russia):
Prometheus (metrics) + Loki (logs) + Grafana (dashboards) + OpenTelemetry (tracing).
Keep it lightweight; wire hooks now, expand later.

## Quick access

| What | URL | Notes |
|------|-----|-------|
| **Grafana** (graphs + logs) | [http://localhost:3001](http://localhost:3001) | Requires `docker compose --profile monitoring up` |
| **Prometheus** (queries) | [http://localhost:9090](http://localhost:9090) | Same profile |
| **Raw metrics** (plain text) | [http://localhost/metrics](http://localhost/metrics) | Always on with main stack — not human-friendly |

**Grafana login:** `GRAFANA_USER` / `GRAFANA_PASSWORD` in `.env` (defaults in `.env.example`).

`/metrics` is Prometheus exposition format for scrapers. **Use Grafana for charts**, not the browser on `/metrics`.

## Logs — Structured (JSON)
Every backend log line is JSON and carries correlation fields when available:
- `request_id` — generated per HTTP request / WS connection
- `trace_id` — from OpenTelemetry context (when tracing enabled)
- `user_id` — added once auth exists
- `meeting_id` — for meeting/room/WS scoped logs

Implementation: Go `log/slog` JSON handler. A middleware injects `request_id` into the
context and the logger. Collected by Loki (via Promtail) when the monitoring profile is running.

Example:
```json
{"time":"...","level":"INFO","msg":"meeting created","request_id":"r-123","meeting_id":"uuid","slug":"abc-defg-hij"}
```

## Metrics — Prometheus
Exposed at **`GET http://localhost/metrics`** (via nginx). See [`docs/local-urls.md`](local-urls.md) for
Grafana/Prometheus URLs. Core metrics:

| metric | type | labels |
|---|---|---|
| `meetings_active` | gauge | — |
| `ws_connections_active` | gauge | — |
| `http_request_duration_seconds` | histogram | `route`, `method`, `status` |
| `http_requests_total` | counter | `route`, `method`, `status` |
| `errors_total` | counter | `component` (incl. `chat`) |
| `transcription_latency_seconds` | histogram | `provider` |
| `chat_messages_total` | counter | — |

Implementation: `prometheus/client_golang` with a middleware that records duration/count
per route, plus gauges updated by the room registry and WS hub.

## Tracing — OpenTelemetry
- Prepare an OTel SDK setup (tracer provider) behind a feature flag (`OTEL_ENABLED`).
- HTTP middleware + WS handlers start spans; `trace_id` flows into logs.
- Exporter (OTLP) configured in Phase 5; no-op by default to avoid early complexity.

## Dashboards
Grafana at [http://localhost:3001](http://localhost:3001) when the **`monitoring` profile** is running
(`docker compose --profile monitoring up`). See `infra/monitoring/README.md` and `docs/local-urls.md`.
Provisioned dashboard **Meeting Platform — Overview**:
- Active meetings, WS connections, request rate/latency, error rate
- Transcription p95 latency by provider
- Backend logs (Loki)

Set `COMPOSE_PROFILES=monitoring` in `.env` to start Grafana/Prometheus with every `docker compose up`.

## Principle
Instrument the seams (HTTP, WS hub, AI pipeline) from the start. The monitoring profile is
optional so the core MVP stays light; metrics at `/metrics` are always available for scrapers.
