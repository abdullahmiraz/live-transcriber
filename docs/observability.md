# Observability Architecture

Stack: **Grafana ecosystem** (popular and well-supported, incl. in Russia):
Prometheus (metrics) + Loki (logs) + Grafana (dashboards) + OpenTelemetry (tracing).
Keep it lightweight; wire hooks now, expand later.

## Logs — Structured (JSON)
Every backend log line is JSON and carries correlation fields when available:
- `request_id` — generated per HTTP request / WS connection
- `trace_id` — from OpenTelemetry context (when tracing enabled)
- `user_id` — added once auth exists
- `meeting_id` — for meeting/room/WS scoped logs

Implementation: Go `log/slog` JSON handler. A middleware injects `request_id` into the
context and the logger. Collected by Loki (via promtail or the OTel collector) in Phase 5.

Example:
```json
{"time":"...","level":"INFO","msg":"meeting created","request_id":"r-123","meeting_id":"uuid","slug":"abc-defg-hij"}
```

## Metrics — Prometheus
Exposed at `GET /metrics`. Core metrics:
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

## Dashboards (Phase 5)
Grafana provisioned dashboards:
- Platform overview: active meetings, WS connections, request rate/latency, error rate.
- AI pipeline: transcription latency, STT/translation error rate.

## Principle
Instrument the seams (HTTP, WS hub, AI pipeline) from the start, but defer the full
collector/Grafana deployment to Phase 5 so the MVP stays simple.
