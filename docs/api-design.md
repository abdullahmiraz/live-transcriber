# API Design — HTTP + WebSocket Contracts

Base URL (through nginx): `/api`
All HTTP request/response bodies are JSON. Errors use a consistent envelope.

## Error Envelope
```json
{ "error": { "code": "not_found", "message": "meeting not found" } }
```
Codes: `bad_request`, `not_found`, `conflict`, `internal`, `unauthorized` (later).

## HTTP Endpoints

### Health
- `GET /healthz` → `200 {"status":"ok"}` (liveness)
- `GET /readyz` → `200 {"status":"ready"}` / `503` (DB reachable)
- `GET /metrics` → Prometheus exposition (text)

### Meetings
- `POST /api/meetings`
  - Request: `{ "title": "Team sync", "host_name": "Alice" }`
  - Response `201`:
    ```json
    {
      "id": "uuid",
      "slug": "abc-defg-hij",
      "title": "Team sync",
      "host_name": "Alice",
      "status": "active",
      "join_url": "/m/abc-defg-hij",
      "created_at": "2026-06-17T08:00:00Z"
    }
    ```
- `GET /api/meetings/{slug}`
  - Response `200`: meeting object (as above) or `404`.
- `POST /api/meetings/{slug}/end`
  - Response `200`: updated meeting with `status:"ended"`.

## WebSocket

Endpoint (through nginx): `GET /ws?meeting={slug}&name={displayName}`
Upgrade to WebSocket. The server places the client into the room hub for `{slug}`.

### Envelope
Every message is JSON:
```json
{ "type": "event.name", "from": "participantId", "to": "participantId|null", "payload": { } }
```
- `type`: event name (see below).
- `from`: sender participant id (server-stamped on broadcast).
- `to`: target participant id for directed messages (signaling); `null` = broadcast.
- `payload`: event-specific data.

### Event Catalog

Lifecycle / presence (server → clients, broadcast):
| type | payload |
|---|---|
| `room.welcome` | `{ "selfId": "...", "participants": [{ "id","name" }] }` |
| `participant.joined` | `{ "id": "...", "name": "..." }` |
| `participant.left` | `{ "id": "..." }` |
| `meeting.created` | `{ "slug": "...", "title": "..." }` (emitted on REST create) |
| `meeting.ended` | `{ "slug": "..." }` |

WebRTC signaling (client → server → directed client):
| type | payload |
|---|---|
| `signal.offer` | `{ "sdp": "..." }` (requires `to`) |
| `signal.answer` | `{ "sdp": "..." }` (requires `to`) |
| `signal.ice` | `{ "candidate": { ... } }` (requires `to`) |

AI pipeline:
| type | direction | payload |
|---|---|---|
| `speech.received` | client → server | `{ "audio": "<base64 pcm/opus>", "seq": 12, "lang": "en" }` |
| `transcript.updated` | server → clients | `{ "participantId","text","lang","isFinal","seq" }` |
| `translation.updated` | server → clients | `{ "participantId","text","sourceLang","targetLang","seq" }` |

### Schema Rules
- Unknown `type` → server replies `{ "type":"error", "payload":{ "code":"unknown_type" } }`.
- Signaling messages MUST include `to`; otherwise dropped with an error reply.
- The server never trusts client-supplied `from`; it stamps the authenticated/session id.

### Design Notes
- Event names are stable strings; versioning via a `v` field can be added later.
- The schema does not assume P2P, so an SFU can later originate `transcript.updated`
  without client changes.
