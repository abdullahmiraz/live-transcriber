# Speech-to-Text (STT) & Translation — Provider Decision

Priority criteria (from spec): **free tier → no credit card → easy integration →
low latency → good accuracy**. Architecture MUST allow swapping providers.

## Options Evaluated (STT)

| Option | Free / No card | Latency | Accuracy | Integration | Notes |
|---|---|---|---|---|---|
| **faster-whisper (self-hosted)** | ✅ free, no card | medium (CPU) / low (GPU) | high | medium (Python svc) | Whisper models, fully offline |
| whisper.cpp (self-hosted) | ✅ free, no card | medium | high | medium | C++ binary, light |
| Deepgram (hosted) | ✅ free credit (signup) | very low (streaming) | high | easy (API) | best for production streaming |
| Google STT | ❌ needs billing acct | low | high | easy | requires credit card |
| Vosk (self-hosted) | ✅ free, no card | low | medium | easy | lightweight, many langs |
| Web Speech API (browser) | ✅ free, no card | low | high | trivial | Chrome/Edge only, no server control |

## Decision

**Architecture:** STT lives behind a Go interface `transcription.Provider`. Nothing
in the WS hub or API depends on a concrete provider. Selection via `STT_PROVIDER` env.

**MVP default provider: `mock`** — a deterministic local provider so the *entire
pipeline* (audio → transcript → translation → captions) runs with `docker compose up`
and **zero external accounts or keys**. This maximizes developer velocity and keeps the
MVP self-contained.

**Recommended real provider: `whisper` (self-hosted faster-whisper)** — free, no credit
card, good accuracy, runs as an optional container. Added when real transcription is
needed without changing application code (just `STT_PROVIDER=whisper`).

**Low-latency production path: `deepgram`** — streaming, lowest latency; drop-in adapter
when a hosted, scalable option is desired.

### Why this choice
- Satisfies *free + no credit card + easy integration* immediately (mock requires nothing).
- Satisfies *accuracy* via Whisper when you flip one env var.
- Satisfies *low latency at scale* via Deepgram adapter later.
- The interface guarantees provider independence (no lock-in).

### Limitations
- `mock` does not perform real transcription (returns canned/echoed text) — intended for
  wiring and demos.
- `whisper` on CPU has higher latency; use small models or GPU for production.
- Streaming partial results require a streaming-capable provider (Deepgram) or chunked
  approximation with Whisper.

### Replacement strategy
1. Implement a new type satisfying `transcription.Provider`.
2. Register it in the provider factory (`transcription/factory.go`).
3. Set `STT_PROVIDER=<name>` + any provider keys. No other code changes.

## Translation — same pattern

Behind `translation.Provider`, selected via `TRANSLATION_PROVIDER`.

| Option | Free / No card | Notes |
|---|---|---|
| **mock** (MVP default) | ✅ | tags text with target lang, zero deps |
| **LibreTranslate (self-hosted)** | ✅ free, no card | open source, container, many langs |
| DeepL API | ✅ free tier (card sometimes) | high quality |
| Google Translate | ❌ billing | high quality |

**Decision:** default `mock`; recommended self-hosted `libretranslate`; swap via env.

## Pipeline Independence
`audio → STT.Provider → transcript → Translation.Provider → captions`. Each stage is an
interface; providers are injected at the composition root. No tight coupling.
