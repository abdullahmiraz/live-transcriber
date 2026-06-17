# Skill: Debugging

## Purpose
Diagnose issues systematically using logs, metrics, and reproduction.

## When to use
- A test fails, a service won't start, or realtime/media/captions misbehave.

## Process
1. Reproduce deterministically; capture exact steps and environment.
2. Read structured logs by `request_id` / `meeting_id`; check `/metrics`.
3. Isolate the layer: transport vs domain vs infra vs provider vs browser.
4. For WS/WebRTC: check signaling order (offer→answer→ICE), ICE state, console.
5. Form one hypothesis, change one thing, verify; avoid shotgun fixes.
6. Add a regression test once fixed.

## Output format
- Root-cause summary (1–3 sentences) + the minimal fix.
- Regression test + a `project-memory.md` "known issue → resolved" note if notable.
