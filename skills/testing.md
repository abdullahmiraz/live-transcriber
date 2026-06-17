# Skill: Testing

## Purpose
Ship business logic with fast, deterministic, contract-focused tests.

## When to use
- Whenever adding domain logic, repositories, endpoints, or WS event handling.

## Process
1. Unit-test domain services with table-driven cases; mock interfaces.
2. Integration-test repositories/migrations against a disposable Postgres.
3. Test the public contract (`docs/api-design.md`), not internals.
4. Use `mock` STT/translation providers for deterministic AI-pipeline tests.
5. Keep tests independent and parallel-safe.

## Output format
- `_test.go` (backend) / `*.test.ts` (frontend) alongside code.
- Brief "how to run" note in the relevant README.
