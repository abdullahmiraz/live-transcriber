# Change History

Chronological journal of **meaningful** changes to Live Meet — what changed, why, how to
verify, and how to roll back.

## Why this exists

Conversation and commits alone do not capture **symptoms, workarounds, and rollback steps**
(e.g. phone HTTPS for camera, dev Docker hot reload). This folder is the place to recover
context when something breaks weeks later.

## How to use

| Goal | Start here |
|------|------------|
| See all documented changes | [`INDEX.md`](INDEX.md) |
| Read one change in depth | `entries/YYYY-MM-DD-*.md` |
| Living decisions & ADRs | [`../project-memory.md`](../project-memory.md) |
| How agents should write entries | [`../../skills/change-documentation.md`](../../skills/change-documentation.md) |

## Adding an entry

1. Copy the template from `skills/change-documentation.md`.
2. Save under `entries/YYYY-MM-DD-short-slug.md`.
3. Add a row to `INDEX.md`.
4. One-line note in `project-memory.md` → **Recent changes**.

Entries are **append-only**; do not delete old entries. If a change is reverted, add a new
entry that says what was reverted and why.
