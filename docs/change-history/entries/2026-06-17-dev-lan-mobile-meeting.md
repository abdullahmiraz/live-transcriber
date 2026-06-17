# Dev stack, LAN/mobile access, meeting UX — 2026-06-17

## Summary

Default `docker compose up` became a **hot-reload dev stack** (Vite + Air). Phones on the
same Wi‑Fi use **HTTPS** on the PC’s LAN IP for camera/mic. The meeting room gained **flexible
join modes**, **10‑minute empty-room deletion**, **mobile chat reliability**, **theme-aware
in-call UI**, and a **non-scrolling video grid** with `object-cover`. Documentation and
smoke tests were updated for dev mode.

## Motivation / symptoms

| Symptom | Root cause |
|---------|------------|
| Camera broken on `/m/{slug}` | SSR `data: [null,null]`; relative assets; `getUserMedia` not on user gesture |
| Rebuild Docker on every edit | Prod images were default compose |
| Phone can’t open camera on LAN IP | Browsers require HTTPS (or localhost) for `getUserMedia` |
| Phone chat send/receive broken until refresh | WS dropped on mobile; chat input clipped off-screen |
| In-call light/dark theme broken | Hardcoded `text-white` / dark chrome on whole call view |
| PC video area scrollable | `aspect-video` tiles + `overflow-auto` in grid |
| Rooms linger after everyone leaves | Only in-memory WS room removed; DB row remained |

## Changes

### Docker & nginx (dev vs prod)

| File | Behavior |
|------|----------|
| `docker-compose.yml` | **Dev default**: bind mounts, `Dockerfile.dev`, Vite + Air, ports 80+443 |
| `docker-compose.prod.yml` | Production built images (no hot reload) |
| `backend/Dockerfile.dev`, `backend/.air.toml` | Go hot reload via Air |
| `frontend/Dockerfile.dev` | Vite dev server, `node_modules` volume |
| `infra/nginx/dev.conf` | Proxy to Vite + backend; `/@vite/` HMR; WS `proxy_buffering off`; TLS :443 |
| `scripts/generate-dev-certs.sh` | Self-signed cert with LAN IP; `MSYS_NO_PATHCONV` + `cd` for Git Bash on Windows |
| `infra/nginx/certs/` | Generated certs (gitignored) |

**Prior behavior:** `docker compose up --build` built prod SvelteKit/Go images; only port 80.

**Rollback:** Use prod stack only:

```bash
docker compose -f docker-compose.prod.yml up --build
```

Restore previous `docker-compose.yml` from git if a single-file default is needed.

### LAN / mobile access

| File | Behavior |
|------|----------|
| `.env.example` | `CORS_ORIGINS=*` for LAN dev |
| `backend/internal/ws/handler.go` | `*` allowed for WS origins |
| `frontend/vite.config.ts` | HMR via nginx (`VITE_HMR_CLIENT_PORT`, `VITE_HMR_PROTOCOL`) |
| `docs/local-urls.md`, `README.md` | HTTPS phone URL, firewall note |

**Phone URL:** `https://<pc-lan-ip>/` — accept cert warning once.

### Media & lobby

| File | Behavior |
|------|----------|
| `frontend/src/lib/media/request-media.ts` | Progressive `getUserMedia` fallbacks; `MediaJoinPreferences`; secure-context error before “not supported” |
| `frontend/src/routes/m/[slug]/+page.svelte` | Join: both / mic only / camera only / chat only; lobby + in-call UX |

### Empty room auto-delete (10 minutes)

| File | Behavior |
|------|----------|
| `backend/internal/ws/hub.go` | When last client leaves, start 10m timer; `DeleteBySlug` if still empty; cancel on rejoin |
| `backend/internal/meeting/repository.go`, `service.go`, `postgres/meeting_repo.go` | `DeleteBySlug` (CASCADE messages/participants) |
| `backend/cmd/server/main.go` | Pass `meetingSvc` into hub |
| `+page.svelte` | Leave confirm + header hint about 10m cleanup |

**Rollback:** Remove timer logic in `hub.go` and `DeleteBySlug` API; meetings stay in DB until manual end.

### Mobile chat reliability

| File | Behavior |
|------|----------|
| `frontend/src/lib/realtime/signaling.ts` | `onOpen`, `reconnectIfNeeded`, `isOpen`/`isConnecting`, outbox flush |
| `frontend/src/lib/components/Chat.svelte` | Sticky input, safe-area, reconnect banner, Enter to send |
| `+page.svelte` | 4s history poll fallback; `visibilitychange` reconnect + sync |

**Prior behavior:** Chat relied only on WS `chat.new`; mobile often missed events.

### In-call theme & video layout

| File | Behavior |
|------|----------|
| `+page.svelte` | Theme tokens for header/footer/sidebar; video stage stays `bg-video-surface` |
| `frontend/src/app.css` | `.video-stage`, `.video-grid`, `.video-tile` — fill area, `object-cover`, no inner scroll |

**Prior behavior:** Full call view forced `text-white`; video grid used `aspect-video` + `overflow-auto`.

### Tests & docs

| File | Behavior |
|------|----------|
| `scripts/smoke-test.mjs` | Detects Vite dev mode (`__sveltekit_dev`); fetches `.svelte` source for lobby checks |
| `docs/docker-architecture.md`, `docs/project-memory.md` | Dev/prod split documented |

## How to verify

- [ ] `bash scripts/generate-dev-certs.sh` then `docker compose up --build`
- [ ] PC: http://localhost — edit `.svelte`/`.go` without rebuild
- [ ] Phone: `https://<lan-ip>/` — join, camera, chat send/receive without refresh
- [ ] `node scripts/smoke-test.mjs http://localhost` → 22/22
- [ ] Leave meeting alone 10+ min → `GET /api/meetings/{slug}` → 404
- [ ] In-call: toggle light/dark; video fills left panel without scrollbar

## Rollback / prior behavior

### Quick switches

| Want | Do |
|------|-----|
| Prod Docker (no hot reload) | `docker compose -f docker-compose.prod.yml up --build` |
| Old video grid (scrollable 16:9 tiles) | Restore `+page.svelte` video section + remove `.video-*` in `app.css` from git |
| No auto-delete | Revert `hub.go` timer + `DeleteBySlug` chain |
| No chat polling | Remove `startChatSync` / `syncChatHistory` interval in `+page.svelte` |

### Git (after commits exist)

```bash
git log --oneline -- docker-compose.yml frontend/src/routes/m/[slug]/+page.svelte
git show <commit>:path/to/file   # inspect prior version
git checkout <commit> -- path    # restore single file (confirm with user first)
```

### File checklist (main touch points)

```
docker-compose.yml
docker-compose.prod.yml
backend/Dockerfile.dev  backend/.air.toml  backend/internal/ws/hub.go
backend/internal/meeting/*  backend/internal/storage/postgres/meeting_repo.go
frontend/Dockerfile.dev  frontend/vite.config.ts
frontend/src/routes/m/[slug]/+page.svelte
frontend/src/lib/media/request-media.ts
frontend/src/lib/realtime/signaling.ts
frontend/src/lib/components/Chat.svelte
frontend/src/app.css
infra/nginx/dev.conf
scripts/generate-dev-certs.sh
scripts/smoke-test.mjs
docs/local-urls.md  docs/docker-architecture.md
```

## Related

- [`docs/local-urls.md`](../local-urls.md) — URLs after `docker compose up`
- [`docs/design-system.md`](../design-system.md) — in-call shell vs video surface
- [`skills/change-documentation.md`](../../skills/change-documentation.md) — how to add future entries
- ADR-008 design system; project-memory camera/LAN notes
