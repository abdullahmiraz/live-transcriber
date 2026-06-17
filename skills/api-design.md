# Skill: API Design

## Purpose
Design consistent, evolvable HTTP and WebSocket contracts.

## When to use
- Adding/changing a REST endpoint or a WS event type.

## Process
1. Define resource/event name, method, request/response, error cases.
2. Reuse the standard error envelope and event envelope (`docs/api-design.md`).
3. Keep handlers thin; validate input at the boundary.
4. For WS: define `type`, whether `to` is required, and the payload schema.
5. Never trust client-supplied identity; server stamps `from`.
6. Consider versioning/back-compat before changing an existing contract.

## Output format
- Updated `docs/api-design.md` table(s).
- Handler/event code + tests asserting the contract.
