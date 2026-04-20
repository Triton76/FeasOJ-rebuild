# Deployment Runbook Update (Phase5 Usability Closure)

## Entry Commands

### backend-rebuild
```bash
cd services/cmd
BACKEND_REBUILD_CONFIG=app/backend-rebuild/config.yaml \
BACKEND_REBUILD_ENABLE_PASSWORD_RESET=false \
go run ./app/backend-rebuild
```

### legacy backend
```bash
cd services/cmd/app/backend
go run .
```

### judgecore
```bash
cd services/cmd/app/judgecore
go run .
```

## Rollout Toggles

### backend-rebuild flags
- `BACKEND_REBUILD_ENABLE_JUDGE_WRITEBACK`
- `BACKEND_REBUILD_ENABLE_SCOREBOARD`
- `BACKEND_REBUILD_ENABLE_TESTCASE_APIS`
- `BACKEND_REBUILD_ENABLE_RABBITMQ_QUEUE`
- `BACKEND_REBUILD_ENABLE_EMBEDDED_JUDGE_WORKER`
- `BACKEND_REBUILD_ENABLE_CLASS_WORKFLOW_V2`
- `BACKEND_REBUILD_ENABLE_PASSWORD_RESET`

## Rollback Plan
1. Disable newly introduced capabilities first:
   - set `BACKEND_REBUILD_ENABLE_PASSWORD_RESET=false`
   - keep other new flags unchanged unless incident scope requires broader rollback.
2. If API-level regressions persist, route traffic back to legacy backend endpoints.
3. Keep judgecore unchanged unless queue/writeback symptoms are observed.
4. Validate health endpoints and smoke tests before restoring traffic.

## Pre-release Checklist
- Confirm change task progress in `openspec/changes/phase5-product-usability-closure-v2/tasks.md`.
- Confirm auth, competitions, and router package tests pass.
- Confirm rollback toggles are available in deployment environment.
