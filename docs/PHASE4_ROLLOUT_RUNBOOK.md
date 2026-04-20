# Phase 4 Staged Rollout Runbook

## Scope

This runbook covers staged enablement for:
- Judge writeback strict mode
- Testcase APIs
- Rule-type specific scoreboard behavior (ACM/OI)
- Frontend class workflow pages and canonical DTO handling

## Feature Flags

Backend feature flags (env or config):
- `BACKEND_REBUILD_ENABLE_JUDGE_WRITEBACK`
- `BACKEND_REBUILD_ENABLE_SCOREBOARD`
- `BACKEND_REBUILD_ENABLE_CLASS_WORKFLOW_V2`

Recommended addition for testcase API enablement:
- `BACKEND_REBUILD_ENABLE_TESTCASE_APIS` (if not yet wired, treat as deployment gate)

## Stage 0: Preflight (staging)

1. Apply additive database migrations.
2. Build and deploy backend with all new code paths disabled by flag.
3. Verify health endpoints and baseline auth flow:
- `GET /api/v1/health`
- `POST /api/v1/auth/login`
4. Confirm queue readiness (memory/rabbit mode) and writeback endpoint reachability.

Exit criteria:
- No migration errors.
- No elevated 5xx in baseline traffic.

## Stage 1: Enable Judge Writeback (staging)

1. Enable `BACKEND_REBUILD_ENABLE_JUDGE_WRITEBACK=true`.
2. Submit a controlled sample submission.
3. Verify state flow:
- `pending -> judging -> terminal`
4. Replay identical terminal writeback and verify idempotent success.
5. Send unauthenticated or invalid writeback payload and verify rejection.

Observability checks:
- Submission transition logs present.
- Writeback success/failure counters increment correctly.
- Latency remains within SLO target.

Exit criteria:
- 100% expected transitions succeed.
- No invalid transition accepted.

## Stage 2: Enable Testcase APIs (staging)

1. Enable testcase API gate (or deploy to environment with route exposure enabled).
2. Validate teacher/admin CRUD:
- Create testcase
- List testcase ordering
- Update testcase
- Delete testcase
- Reorder testcase set
3. Validate permission boundaries:
- Student mutation attempt returns 403.
4. Validate judge-facing testcase read returns full active set.

Exit criteria:
- CRUD and ordering deterministic.
- Validation and permission semantics match contract.

## Stage 3: Enable ACM/OI Scoreboard Semantics (staging)

1. Enable `BACKEND_REBUILD_ENABLE_SCOREBOARD=true`.
2. Run ACM fixtures:
- solved desc, penalty asc, tie-break timestamp
- freeze behavior in final 60 minutes
3. Run OI fixtures:
- per-problem best score aggregation
- deterministic tie-break behavior
4. Verify admin scoreboard UI reads canonical payload keys.

Exit criteria:
- ACM/OI results match expected golden fixtures.
- Frontend shows correct ranking values.

## Stage 4: Class Workflow UI Validation (staging)

1. Enable `BACKEND_REBUILD_ENABLE_CLASS_WORKFLOW_V2=true`.
2. Validate role-based flows:
- Teacher create/edit/archive class
- Student join by class code
- Assistant/teacher review join applications
3. Verify action visibility by role.

Exit criteria:
- All role workflows pass without local fabricated fallback data.

## Stage 5: Canary (production)

1. Roll out backend to canary slice (5%-10% traffic).
2. Enable only judge writeback first.
3. Observe 30-60 minutes:
- writeback error rate
- queue lag
- scoreboard latency
4. If stable, enable testcase APIs and scoreboard mode behavior.
5. Enable class workflow UI last.

Exit criteria:
- Metrics stable within SLO.
- No P1/P2 regression reported.

## Stage 6: Full Production Switch

1. Scale from canary to 100%.
2. Keep enhanced monitoring alerts active for 24 hours.
3. Record release evidence in acceptance document.

## Rollback Procedure

### Trigger Conditions

Rollback if any of:
- Writeback transition conflict spike or terminal-state corruption risk
- Testcase API authorization or data-integrity fault
- Scoreboard ranking regression for ACM/OI
- Class workflow blocks critical user operations

### Flag-off Sequence

1. Disable class workflow UI gate:
- `BACKEND_REBUILD_ENABLE_CLASS_WORKFLOW_V2=false`
2. Disable scoreboard gate:
- `BACKEND_REBUILD_ENABLE_SCOREBOARD=false`
3. Disable testcase API gate:
- `BACKEND_REBUILD_ENABLE_TESTCASE_APIS=false` (or route-level disable)
4. Disable judge writeback gate:
- `BACKEND_REBUILD_ENABLE_JUDGE_WRITEBACK=false`

### Service Fallback

1. Keep submission creation active; pause external writeback producers.
2. Route scoreboard reads to prior stable behavior if available.
3. Hide frontend entry points for affected capabilities.

### Data Integrity Checks

1. Verify no invalid state transitions in submissions table.
2. Verify testcase order uniqueness and bounds per problem.
3. Verify contest ranking snapshots are reproducible for known fixtures.
4. Verify class membership states remain valid (`pending|active|rejected|archived`).

### Recovery Path

1. Root-cause the issue with logs/metrics and failing request samples.
2. Patch in staging and rerun acceptance checklist.
3. Re-enter canary from Stage 5.
