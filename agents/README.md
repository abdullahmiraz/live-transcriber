# Agents

Specialized instruction files for the multiple AI agents collaborating on this project.
Each agent must read `docs/architecture.md` and `docs/project-memory.md` before working,
and update `docs/project-memory.md` after important changes.

| Agent | File | Scope |
|---|---|---|
| Project Orchestrator | `orchestrator.md` | task breakdown, agent coordination, progress |
| Architect | `architect.md` | system design, consistency, decisions |
| Backend Engineer | `backend-engineer.md` | Go services, API, WS hub |
| Frontend Engineer | `frontend-engineer.md` | SvelteKit, WebRTC, captions UI |
| Design System | `design-system.md` | tokens, typography, motion, visual consistency |
| DevOps | `devops.md` | Docker, nginx, CI, deploy |
| GitOps | `gitops.md` | branches, commits, merges, repo hygiene |
| Database | `database.md` | schema, migrations, indexing |
| Testing | `testing.md` | unit/integration/e2e, quality gates |

Each agent must read `docs/architecture.md`, `docs/design-system.md` (for UI work),
`docs/project-memory.md`, and skim `docs/change-history/INDEX.md` before working.
After important changes, add a history entry per `skills/change-documentation.md`.

## Shared rules (all agents)
- Working MVP first; do not over-engineer.
- Clean architecture, clear package boundaries, no giant files.
- No unnecessary dependencies; no microservices without reason.
- Before a major change: state problem, solution, alternatives, decision, plan.
- Always update docs and `project-memory.md`; add a **change-history entry** for behavioral
  changes (see `skills/change-documentation.md`); add tests for new logic.
- Keep providers (STT/translation) swappable behind interfaces.
- UI changes must follow `docs/design-system.md` and stay consistent across pages.
