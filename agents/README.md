# Agents

Specialized instruction files for the multiple AI agents collaborating on this project.
Each agent must read `docs/architecture.md` and `docs/project-memory.md` before working,
and update `docs/project-memory.md` after important changes.

| Agent | File | Scope |
|---|---|---|
| Architect | `architect.md` | system design, consistency, decisions |
| Backend Engineer | `backend-engineer.md` | Go services, API, WS hub |
| Frontend Engineer | `frontend-engineer.md` | SvelteKit, WebRTC, captions UI |
| DevOps | `devops.md` | Docker, nginx, CI, deploy |
| Database | `database.md` | schema, migrations, indexing |
| Testing | `testing.md` | unit/integration/e2e, quality gates |

## Shared rules (all agents)
- Working MVP first; do not over-engineer.
- Clean architecture, clear package boundaries, no giant files.
- No unnecessary dependencies; no microservices without reason.
- Before a major change: state problem, solution, alternatives, decision, plan.
- Always update docs and `project-memory.md`; add tests for new logic.
- Keep providers (STT/translation) swappable behind interfaces.
