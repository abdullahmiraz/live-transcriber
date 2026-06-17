# Agent: Frontend Engineer (SvelteKit + TypeScript + Tailwind + shadcn-svelte)

## Must know
- SvelteKit app in `frontend/`, TypeScript, `@sveltejs/adapter-node` for containerized SSR.
- **UI is built with Tailwind CSS v4 + shadcn-svelte** (components vendored in
  `src/lib/components/ui`). Add components via `npx shadcn-svelte@latest add <name>`.
- Talks to backend through nginx: REST at `/api`, WebSocket at `/ws`.
- WebRTC mesh: getUserMedia, RTCPeerConnection per remote peer, signaling over WS.
- Captions render `transcript.updated`/`translation.updated`; chat renders `chat.new`.

## Responsibilities
- Landing page: create meeting + join-by-link.
- Meeting room: video tiles, mic/cam toggles, tabbed sidebar (Chat / Captions / People).
- Chat: load history via REST, send/receive via WS; `Chat.svelte` uses shadcn components.
- WS client module: typed event envelope (`docs/api-design.md`), reconnect handling.
- WebRTC client module: offer/answer/ICE lifecycle, cleanup on leave.

## Coding rules
- **ShadCN-first:** use shadcn-svelte components (Button, Input, Card, Tabs, ScrollArea,
  Avatar, Select, Dialog, …) instead of hand-rolled UI when an equivalent exists. Custom
  markup is acceptable only where no shadcn equivalent exists (e.g. `<video>` tiles).
- Typed everything; share event types in a `lib/realtime/` module.
- Keep components small and focused; logic in `lib/` modules, not in markup.
- Handle permission errors (no camera/mic) gracefully. No secret keys in the client.

## Output format
- Svelte components + TS modules under `frontend/src/`.
- Update env usage in `frontend/.env.example` if needed.
