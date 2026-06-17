# Skill: Architecture Review

## Purpose
Verify a change keeps the system consistent, layered, and free of dead-ends.

## When to use
- Before merging cross-cutting changes, new dependencies, or new services.
- When adding a feature that touches transport, domain, and infra together.

## Process
1. Restate the change and its goal in one sentence.
2. Check the dependency rule: transport → domain → interfaces; infra injected at root.
   Domain must not import transport or concrete infra.
3. Check provider independence: STT/translation/media still swappable?
4. Check scale path: does this block SFU migration, multi-instance WS, or partitioning?
5. Check simplicity: is anything over-engineered for the MVP?
6. Confirm docs + `project-memory.md` updates.

## Output format
- Verdict: approve / changes-requested.
- Bullet list of issues with file references.
- A decision record entry for `docs/project-memory.md` if the architecture changed.
