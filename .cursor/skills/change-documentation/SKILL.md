---
name: change-documentation
description: >-
  Documents project changes in docs/change-history with rollback notes and INDEX updates.
  Use after features, bug fixes, Docker/infra, mobile/LAN, or meeting UI changes; when
  the user asks to document work, keep history, or revert behavior; when debugging
  regressions (read INDEX first).
---

# Change Documentation & History

Follow the project skill: **`skills/change-documentation.md`** (full process and template).

## Quick workflow

1. Read `docs/change-history/INDEX.md` for context.
2. Create `docs/change-history/entries/YYYY-MM-DD-short-slug.md` (use template in skill).
3. Add a row to `INDEX.md` (newest first).
4. Add one line under **Recent changes** in `docs/project-memory.md`.
5. Update `README.md` / `docs/local-urls.md` / `docs/design-system.md` if user-facing.

## Rollback section is mandatory

Every entry must state how to restore prior behavior (git paths, compose file, env vars).

## Do not log secrets

Never put `.env` values, certs, or credentials in history files.
