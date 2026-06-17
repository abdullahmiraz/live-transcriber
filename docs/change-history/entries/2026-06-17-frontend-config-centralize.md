# Frontend config centralization — 2026-06-17

## Summary
Moved repeatable meeting UI data and TypeScript types out of `+page.svelte` into shared
modules: `config/languages.json`, `meeting/types.ts`, `meeting/constants.ts`, and helpers
for session, routes, join-media, and encoding.

## Rollback
Revert commit; meeting page imports return inline.
