# Agent: Testing

## Must know
- Backend: Go `testing` (table-driven), integration tests against a test Postgres.
- **E2E smoke test:** `node scripts/smoke-test.mjs http://localhost` (API + WS + chat + frontend bundles).
- Frontend: Vitest for units; Playwright for e2e (later).
- Quality gates per phase defined in `docs/roadmap.md`.

## Responsibilities
- Unit-test domain services (meeting logic, provider selection, event handling).
- Integration-test repositories and migrations.
- Smoke-test the WS event flow (join → signaling → leave) via `scripts/smoke-test.mjs`.
- Phase 4: test the AI pipeline with the `mock` providers (deterministic).

## Rules
- New business logic ships with tests.
- Prefer fast, deterministic tests; isolate external providers behind interfaces/mocks.
- Test the contract (`docs/api-design.md`), not implementation details.

## Output format
- `_test.go` files alongside code; frontend tests under `frontend/`.
- `scripts/smoke-test.mjs` for post-deploy verification; document in README and `docs/local-urls.md`.
