# Skill: Database Design

## Purpose
Evolve the schema safely with migrations, integrity, and good indexing.

## When to use
- Adding/altering tables, columns, indexes, or constraints.

## Process
1. Model the change against query patterns (how will it be read/written?).
2. Add a new forward migration `NNNN_description.sql` (never edit merged migrations).
3. Use UUID PKs, `timestamptz`, FKs with explicit ON DELETE behavior.
4. Add indexes for lookups (e.g., `slug`, `meeting_id`, `(meeting_id, created_at)`).
5. For high-volume tables, plan partitioning before it hurts.
6. Verify migration runs cleanly via the startup runner.

## Output format
- Migration file(s) in `backend/migrations/`.
- Updated `docs/database-design.md` + ER diagram.
