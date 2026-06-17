# Monitoring profile & docs sync — 2026-06-17

## Summary
Integrated Grafana/Prometheus into main `docker-compose.yml` via the `monitoring` profile.
Updated all docs to clarify: plain `docker compose up` does not open ports 3001/9090;
Grafana login comes from `GRAFANA_USER` / `GRAFANA_PASSWORD` in `.env`.

## Start monitoring

```bash
docker compose --profile monitoring up -d
```

## Rollback
Remove `include` and profile from compose; revert doc tables to old two-file compose command.
