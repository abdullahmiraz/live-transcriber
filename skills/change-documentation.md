# Skill: Change Documentation & History

## Purpose
Keep a **durable, searchable history** of meaningful project changes so anyone (human or agent)
can understand what changed, why, how to verify it, and **how to roll back** if something breaks.

## When to use
- After completing a feature, bug fix, infra change, or UX overhaul (not every typo).
- When the user asks to document work, create a history entry, or explain how to revert.
- **Before** closing a multi-file task — if you changed behavior users rely on, log it.
- When debugging regressions — read `docs/change-history/INDEX.md` first.

## Do not use for
- Trivial one-line edits with no behavioral impact.
- Secrets, tokens, or `.env` values (never write those in history).

## Process

### 1. Decide if an entry is needed

Create an entry when **any** of these apply:
- New user-visible behavior (UI, API, Docker, mobile/LAN access)
- New env vars, ports, or startup steps
- Architecture or data-flow change
- Known workaround (HTTPS on phone, cert script, etc.)

### 2. Create or update a history entry

1. Open `docs/change-history/INDEX.md` and add a row (newest first).
2. Create `docs/change-history/entries/YYYY-MM-DD-short-slug.md` using the template below.
3. Add a **one-line** pointer in `docs/project-memory.md` under "Recent changes" (keep last ~5).
4. Update topical docs if they exist (`docs/local-urls.md`, `docs/design-system.md`, `README.md`).

### 3. Entry template (copy into new file)

```markdown
# [Title] — YYYY-MM-DD

## Summary
One paragraph: what changed and why.

## Motivation / symptoms
What was broken or requested.

## Changes
| Area | Files | Behavior |
|------|-------|----------|
| … | `path/to/file` | … |

## How to verify
- [ ] Step 1
- [ ] Step 2

## Rollback / prior behavior
- **Git:** `git show <commit>` or `git checkout <commit> -- path`
- **Manual:** describe previous behavior and which files to restore
- **Docker:** e.g. use `docker-compose.prod.yml` instead of default compose

## Related
- Links to ADRs, `docs/…`, agents, skills
```

### 4. Rollback guidance rules

Always include **actionable** rollback:
- List **exact file paths** touched.
- If behavior switched (e.g. dev vs prod compose), name the **old default**.
- If a migration or DB delete was added, note data impact.
- Prefer `git log --oneline -- path` for finding the introducing commit after push.

### 5. INDEX format

`docs/change-history/INDEX.md` table columns:

| Date | Slug | Summary | Key files | Entry |
|------|------|---------|-----------|-------|

Link `Entry` to `entries/YYYY-MM-DD-short-slug.md`.

## Output format (after each documented change)

Reply briefly:
1. Entry path created/updated
2. INDEX row added
3. Which user-facing docs were updated (if any)

## Related files
- Journal index: `docs/change-history/INDEX.md`
- Entries: `docs/change-history/entries/`
- Living decisions: `docs/project-memory.md`
- Design tokens: `docs/design-system.md`
- Local URLs: `docs/local-urls.md`
