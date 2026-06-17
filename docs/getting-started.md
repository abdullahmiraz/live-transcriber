# Getting Started — Run at Full Power

> **One-page checklist** to turn on every MVP feature: meetings, WebRTC A/V, chat,
> live captions + translation, mobile over Wi‑Fi, hot reload, and optional monitoring.

For URL reference after the stack is up, see **[`local-urls.md`](local-urls.md)**.

---

## Prerequisites

| Tool | Why |
|------|-----|
| **Docker Desktop** | Runs Postgres, Redis, backend, frontend, nginx |
| **Node.js 18+** | Runs `scripts/smoke-test.mjs` (optional but recommended) |
| **Git Bash** (Windows) | Runs `scripts/generate-dev-certs.sh` for phone HTTPS |

---

## One-time setup (≈5 minutes)

```bash
git clone <repo-url> live-transcript
cd live-transcript
cp .env.example .env
bash scripts/generate-dev-certs.sh   # required for phone camera/mic on Wi‑Fi
docker compose up --build            # first time only; later: docker compose up
```

Wait until containers are healthy. Open **[http://localhost/](http://localhost/)**.

**Verify the stack:**

```bash
node scripts/smoke-test.mjs http://localhost
```

All checks should pass (health, API, WebSocket, chat, meeting page bundles).

---

## Full-power feature checklist

Use this table to confirm every capability is working on your machine.

| Feature | How to use it | Full power requirements |
|---------|---------------|-------------------------|
| **Create / join meeting** | Home → name → Create or paste code → Join | `docker compose up` running |
| **Video + audio (this PC)** | Lobby → **Join with camera & microphone** (or mic-only / camera-only) | `http://localhost` — browser permission on click |
| **Video + audio (phone / tablet)** | Same Wi‑Fi → `https://<your-pc-lan-ip>/` | **HTTPS required** — run cert script once; accept cert warning on phone |
| **Text chat** | In-call → **Chat** tab → type and send | Chrome/Edge/Safari; auto-reconnects on mobile |
| **Live captions** | Control bar → captions icon (CC) | **Chrome or Edge** (Web Speech API); mic must be on |
| **Translation line** | Set **speak** language → **translate to** language (control bar) | Mock provider shows `[lang] text` under original; real providers via env (see below) |
| **Copy meeting link** | In-call → copy button | Link uses current origin (HTTPS on phone → HTTPS link for guests) |
| **Leave / empty room** | Leave button; room auto-deletes ~10 min after last person leaves | — |
| **Hot reload (dev)** | Edit `frontend/` or `backend/` files | Use default `docker compose up`, **not** `docker-compose.prod.yml` |
| **Monitoring (optional)** | Grafana dashboards + log search | `docker compose --profile monitoring up -d`; login from `.env` |

### Find your LAN IP (for phones)

- **Windows:** `ipconfig` → IPv4 (e.g. `192.168.1.42`)
- **macOS / Linux:** `ip addr` or System Settings → Network

Example meeting URL on phone: `https://192.168.1.42/m/your-meeting-code`

---

## Typical first session (5 steps)

1. **Start stack** — `docker compose up`
2. **PC browser** — [http://localhost/](http://localhost/) → create a meeting
3. **Join with media** — lobby → choose join mode → allow camera/mic
4. **Turn on captions** — CC button → speak → see transcript; pick **translate to** for second line
5. **Phone (optional)** — `https://<lan-ip>/` → same meeting code → accept cert → join

Open a second browser tab or incognito window to simulate two participants (chat + captions sync).

---

## Environment (`.env`)

Copy from `.env.example`. Defaults are enough for full MVP power — **no API keys required**.

| Variable | Default | Purpose |
|----------|---------|---------|
| `CORS_ORIGINS` | `*` | Allows phones on LAN to call the API |
| `STT_PROVIDER` | `mock` | Server echoes browser speech text |
| `TRANSLATION_PROVIDER` | `mock` | Tags translation with target lang (`[ru] …`) |
| `DEFAULT_TARGET_LANG` | `ru` | Fallback if client omits `targetLang` |
| `REDIS_URL` | `redis://redis:6379/0` | Chat fan-out across WS connections |
| `GRAFANA_USER` / `GRAFANA_PASSWORD` | `admin` / `admin` | Grafana login when monitoring profile is on |
| `COMPOSE_PROFILES` | _(unset)_ | Set to `monitoring` to auto-start Grafana/Prometheus |

**Upgrade to real AI later** (optional): set `STT_PROVIDER=whisper` or `deepgram`, `TRANSLATION_PROVIDER=libretranslate`, plus provider keys. See [`stt-decision.md`](stt-decision.md).

---

## Monitoring overlay (optional)

**Grafana and Prometheus are not started by plain `docker compose up`.** Use the `monitoring` profile:

```bash
docker compose --profile monitoring up -d
```

| Service | URL | Login |
|---------|-----|-------|
| **Grafana** (graphs + logs) | [http://localhost:3001](http://localhost:3001) | `GRAFANA_USER` / `GRAFANA_PASSWORD` from `.env` |
| **Prometheus** (query UI) | [http://localhost:9090](http://localhost:9090) | — |
| Raw metrics (text, not graphs) | [http://localhost/metrics](http://localhost/metrics) | — |

Dashboard: **Meeting Platform — Overview** (auto-loaded in Grafana).

Set credentials in `.env` (copy from `.env.example`):

```
GRAFANA_USER=admin
GRAFANA_PASSWORD=admin
```

To always start monitoring with the app:

```
COMPOSE_PROFILES=monitoring
```

Details: [`infra/monitoring/README.md`](../infra/monitoring/README.md), [`observability.md`](observability.md).

---

## Production-style build (no hot reload)

When you want a built image instead of live reload:

```bash
docker compose -f docker-compose.prod.yml up --build
```

Use for demos or closer-to-prod runs. Day-to-day development should use plain `docker compose up`.

---

## Troubleshooting (quick)

| Problem | Fix |
|---------|-----|
| Phone camera blocked | Use **`https://<lan-ip>/`**, not `http://192.168.x.x` |
| Cert warning on phone | Expected — Advanced → Proceed (once per device) |
| Phone can't reach PC | Allow Windows Firewall inbound **80** and **443** for Docker |
| Captions don't appear | Use **Chrome/Edge**; turn on CC; ensure mic is enabled |
| Translation missing | Pick different **speak** vs **translate to** languages |
| Changes not showing | `docker compose up` (dev), not prod compose file |
| `readyz` 503 | Wait for Postgres healthcheck; `docker compose logs backend` |

Full table: [`local-urls.md` §7](local-urls.md#7-troubleshooting).

---

## Documentation map

| I want to… | Read |
|------------|------|
| **Start here (this page)** | `docs/getting-started.md` |
| All URLs after `docker compose up` | [`local-urls.md`](local-urls.md) |
| API + WebSocket contracts | [`api-design.md`](api-design.md) |
| Docker topology | [`docker-architecture.md`](docker-architecture.md) |
| Change log + rollback | [`change-history/INDEX.md`](change-history/INDEX.md) |
| Architecture overview | [`architecture.md`](architecture.md) |
| STT / translation providers | [`stt-decision.md`](stt-decision.md) |
| UI tokens | [`design-system.md`](design-system.md) |

---

## Local dev without Docker

Possible but not the recommended “full power” path (no bundled HTTPS for phones). See
[`local-urls.md` §4](local-urls.md#4-local-development-without-docker).
