# Agent: Frontend Engineer (SvelteKit + TypeScript)

## Must know
- SvelteKit app in `frontend/`, TypeScript, `@sveltejs/adapter-node` for containerized SSR.
- Talks to backend through nginx: REST at `/api`, WebSocket at `/ws`.
- WebRTC mesh: getUserMedia, RTCPeerConnection per remote peer, signaling over WS.
- Captions UI renders `transcript.updated` + `translation.updated` events.

## Responsibilities
- Landing page: create meeting + join-by-link.
- Meeting room: local + remote video tiles, mic/cam toggles, participant list.
- WS client module: typed event envelope (`docs/api-design.md`), reconnect handling.
- WebRTC client module: offer/answer/ICE lifecycle, cleanup on leave.
- Captions panel: original + translated lines, per speaker.

## Coding rules
- Typed everything; share event types in a `lib/realtime/` module.
- Keep components small and focused; logic in `lib/` modules, not in markup.
- Handle permission errors (no camera/mic) gracefully.
- No secret keys in the client.

## Output format
- Svelte components + TS modules under `frontend/src/`.
- Update env usage in `frontend/.env.example` if needed.
