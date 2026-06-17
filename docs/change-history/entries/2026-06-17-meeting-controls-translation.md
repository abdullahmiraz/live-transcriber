# Meeting control bar & caption translation target — 2026-06-17

## Summary
Polished the in-call control bar (contrast, spacing, mic speaking wave) and fixed caption
translation: the language dropdown previously only set speech-recognition source; translation
target was server-only (`DEFAULT_TARGET_LANG`). Users now pick **speak** and **translate to**
languages; the client sends `targetLang` with each caption chunk.

## Files
- `frontend/src/app.css` — control bar tokens and mic wave overlay
- `frontend/src/lib/components/meeting/*` — shared control buttons
- `frontend/src/lib/media/mic-level.ts` — Web Audio mic level monitor
- `frontend/src/routes/m/[slug]/+page.svelte` — control bar, dual lang selects, caption handlers
- `backend/internal/ws/client.go`, `message.go` — `targetLang` on `speech.received`
- `frontend/src/lib/realtime/types.ts`, `docs/api-design.md` — WS contract

## Verify
1. Join a meeting, turn on captions, speak in English.
2. Set **Translate to** Russian (or Spanish) — italic line shows `[ru] …` (mock provider).
3. Mic button stays fixed width when speaking wave animates.

## Rollback
Revert commit on `feature/design-system-and-agents`; no migrations.
