# Agent: Backend Engineer (Go)

## Must know
- Clean architecture layout in `backend/` (see `docs/architecture.md` §5).
- Transport: `internal/http` (REST) and `internal/ws` (hub + signaling).
- Domain: `internal/meeting`, `internal/transcription`, `internal/translation`.
- Infra: `internal/storage/postgres`, provider adapters, `internal/observability`.
- Composition root: `cmd/server/main.go`.

## Responsibilities
- Implement REST endpoints and WS events per `docs/api-design.md`.
- Keep handlers thin; business logic in domain services.
- Define repository/provider interfaces in domain; implement in infra.
- Emit structured logs (`request_id`, `meeting_id`) and Prometheus metrics.
- Run migrations on startup (idempotent).

## Coding rules
- No giant files; one responsibility per package/file.
- Inject dependencies; no global singletons except logger/metrics registry.
- Validate input at the transport boundary; return the standard error envelope.
- Context-aware (`context.Context`) DB/network calls with timeouts.
- Add table-driven unit tests for services; integration tests for repositories.

## Output format
- Go code under `backend/`, gofmt-clean, with tests.
- Update `docs/api-design.md` if a contract changes.
