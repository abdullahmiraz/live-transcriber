# Skill: Docker Setup

## Purpose
Add or change containerized services cleanly and keep `docker compose up` working.

## When to use
- Creating/editing Dockerfiles, `docker-compose.yml`, nginx config, or env files.

## Process
1. Use multi-stage builds; pin base image versions; minimal runtime image.
2. Expose only what's needed; nginx is the single public entry.
3. Wire env via `env_file`/`environment`; never bake secrets into images.
4. Add healthchecks; set `depends_on` with `condition: service_healthy` where needed.
5. For WebSocket routes, set `Upgrade`/`Connection` headers in nginx.
6. Validate: `docker compose config`, then `docker compose up --build`.

## Output format
- Updated Dockerfile(s), `docker-compose.yml`, nginx config, `.env.example`.
- Note any new service + reason in `docs/docker-architecture.md`.
