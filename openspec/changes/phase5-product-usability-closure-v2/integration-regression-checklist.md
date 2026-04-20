# Integrated Regression Checklist (Phase5 A-E)

Date: 2026-04-20
Scope: contest binding, class join lifecycle, account security reset, admin role governance, config/runtime convergence.

## Backend Regression (A-E)

Command:
`cd services/cmd && go test ./app/backend-rebuild/internal/usecase/auth ./app/backend-rebuild/internal/usecase/competitions ./app/backend-rebuild/internal/usecase/classes ./app/backend-rebuild/internal/usecase/admin ./app/backend-rebuild/internal/http/router`

Result:
- PASS: `internal/usecase/auth`
- PASS: `internal/usecase/competitions`
- PASS: `internal/usecase/classes`
- PASS: `internal/usecase/admin`
- PASS: `internal/http/router`

## Frontend Regression (Class Membership Lifecycle)

Command:
`cd web && npm run build`

Result:
- PASS: production build completed.
- NOTE: existing bundle-size warnings remain (non-blocking for this phase).

## Capability Mapping

- A Contest-Problem Binding: covered by `internal/usecase/competitions` + router integration path.
- B Class Join Usability: covered by `internal/usecase/classes` + router integration path + frontend build.
- C Account Security Assist: covered by `internal/usecase/auth`.
- D Admin Role Governance: covered by `internal/usecase/admin` + router integration path.
- E Config/Ops Entry Convergence: covered by `internal/config` tests and runtime source reporting paths already validated in prior checkpoints.

## Conclusion

- Task 7.1 acceptance condition is satisfied in local CI-equivalent validation.
- Task 7.2 (staging rollout/rollback control validation) still requires staging environment execution.
- Task 7.3 (implementation readiness handoff) should be finalized after 7.2 staging evidence is attached.
