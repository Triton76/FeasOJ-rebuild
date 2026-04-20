# Config Precedence Policy (backend / backend-rebuild / judgecore)

## Scope
- services/cmd/app/backend
- services/cmd/app/backend-rebuild
- services/cmd/app/judgecore

## Deterministic Precedence

### backend-rebuild
1. `BACKEND_REBUILD_*` environment variables
2. `BACKEND_REBUILD_CONFIG` target file (default: `app/backend-rebuild/config.yaml`)
3. code defaults

### backend (legacy)
1. `config.toml` in runtime working directory
2. code defaults in config initialization

### judgecore
1. `config.toml` in runtime working directory
2. code defaults when missing keys or generated default file

## Operational Rules
- ENV overrides file for same key in backend-rebuild.
- Missing mandatory runtime secrets in production MUST be treated as startup failure by deploy scripts.
- Compatibility window allows YAML (backend-rebuild) and TOML (backend/judgecore) to coexist.

## Required Deploy-time Checks
- Verify effective config source path in startup logs.
- Verify JWT secret and DB DSN are set for backend-rebuild production profiles.
- Verify legacy/judgecore `config.toml` exists when those services are enabled.
