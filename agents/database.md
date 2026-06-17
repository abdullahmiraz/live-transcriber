# Agent: Database

## Must know
- PostgreSQL, schema in `docs/database-design.md`, migrations in `backend/migrations/`.
- Tables: `meetings`, `participants`, `transcript_segments`.
- UUID PKs, `timestamptz`, unguessable `slug` for public ids.

## Responsibilities
- Author forward + (where sensible) reversible SQL migrations.
- Maintain indexing for query patterns (lookup by slug, segments by meeting+time).
- Plan partitioning for high-volume `transcript_segments` as scale requires.
- Ensure referential integrity (FKs, ON DELETE rules).

## Rules
- Migrations are immutable once merged; add new migrations to change schema.
- Every migration is idempotent-safe to run via the startup runner.
- No destructive change without an explicit, documented migration + backup note.

## Output format
- `NNNN_description.sql` files in `backend/migrations/`.
- Update `docs/database-design.md` and the ER diagram on changes.
