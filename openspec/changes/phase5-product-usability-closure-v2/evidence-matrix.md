# Phase5 Evidence Matrix (A-E)

## A. Contest-Problem Binding Closure
- confirmed: `contest_problems` table exists with uniqueness constraints.
  - evidence: `services/cmd/app/backend-rebuild/migrations/0003_create_phase1_core_tables.sql`
- confirmed: contest contracts have no explicit binding payload field.
  - evidence: `services/cmd/app/backend-rebuild/internal/ports/contracts.go`
- confirmed: contest router has no binding read/write endpoint.
  - evidence: `services/cmd/app/backend-rebuild/internal/http/router/router.go`

## B. Class Join Usability and Feedback
- confirmed: class join request accepts `class_code`.
  - evidence: `services/cmd/app/backend-rebuild/internal/ports/contracts.go`
- confirmed: apply join creates `pending` membership.
  - evidence: `services/cmd/app/backend-rebuild/internal/usecase/classes/service.go`
- confirmed: review updates pending membership to approved/rejected state.
  - evidence: `services/cmd/app/backend-rebuild/internal/usecase/classes/service.go`

## C. Account Security Assist
- confirmed: legacy backend exposes captcha and password update endpoints.
  - evidence: `services/cmd/app/backend/server/router.go`, `services/cmd/app/backend/server/handler/users.go`
- confirmed: rebuild backend lacks password-reset/captcha API in auth routes.
  - evidence: `services/cmd/app/backend-rebuild/internal/http/router/router.go`

## D. Admin Role/Status Governance
- confirmed: rebuild admin routes expose list users + status update only.
  - evidence: `services/cmd/app/backend-rebuild/internal/http/router/router.go`
- confirmed: legacy backend has privilege change route.
  - evidence: `services/cmd/app/backend/server/router.go`

## E. Config/Ops Entry Convergence
- confirmed: rebuild config uses YAML + ENV override.
  - evidence: `services/cmd/app/backend-rebuild/internal/config/config.go`
- confirmed: legacy backend and judgecore use TOML config loading.
  - evidence: `services/cmd/app/backend/internal/config/config.go`, `services/cmd/app/judgecore/internal/config/config.go`
