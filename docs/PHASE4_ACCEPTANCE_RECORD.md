# Phase 4 Acceptance Record

## Change

- Change name: `phase4-core-oj-closure`
- Schema: `spec-driven`
- Date: 2026-04-20

## Release Checklist

### 1. JudgeCore End-to-End Lifecycle

- [ ] 1.1 Judge payload/writeback contract is versioned and documented
- [ ] 1.2 Writeback authentication and validation enforced
- [ ] 1.3 State machine verified: `pending -> judging -> terminal`
- [ ] 1.4 Idempotent replay and invalid-transition conflict checks verified
- [ ] 1.5 Structured logs/metrics observed in runtime
- [ ] 1.6 Integration tests pass for enqueue/consume/writeback/auth failure

### 2. Testcase Authoring and Judge Binding

- [ ] 2.1 Domain DTO/ports and authorization rules implemented
- [ ] 2.2 Testcase repository CRUD and deterministic ordering implemented
- [ ] 2.3 Testcase HTTP handlers/routes implemented
- [ ] 2.4 Payload validation (required fields/size/count limits) verified
- [ ] 2.5 Judge-facing testcase read endpoint verified
- [ ] 2.6 Unit/integration tests pass for permission/validation/ordering

### 3. Contest ACM/OI Parity

- [x] 3.1 Scoreboard branches by `rule_type`
- [x] 3.2 ACM ranking semantics and freeze behavior implemented
- [x] 3.3 OI per-problem best-score aggregation and tie-break implemented
- [x] 3.4 Contest visibility/join policy remains mode-agnostic
- [x] 3.5 Golden-case tests include ACM freeze and OI scoring

Evidence:
- Backend tests: `go test ./app/backend-rebuild/...`
- Scoreboard tests: `internal/usecase/competitions/scoreboard_test.go`

### 4. Frontend-Backend Contract Alignment

- [ ] 4.1 Legacy field shims removed from critical API adapters
- [ ] 4.2 Status/profile/admin pages consume canonical rebuild DTO
- [x] 4.3 Admin contest scoreboard parsing handles canonical payload and pagination
- [ ] 4.4 Fabricated local fallbacks removed for contest/class critical data
- [x] 4.5 Error semantics align with backend status contracts
- [ ] 4.6 Frontend smoke checklist completed

### 5. Class Management Frontend Workflows

- [ ] 5.1 Class routes and navigation entry implemented
- [ ] 5.2 Teacher class create/edit/archive flows implemented
- [ ] 5.3 Student class join and status pages implemented
- [ ] 5.4 Assistant/teacher membership review UI implemented
- [ ] 5.5 E2E UI checks for class lifecycle and review completed

### 6. Rollout, Migration, Safety

- [x] 6.1 Feature flag strategy defined and available in config
- [ ] 6.2 Additive DB migrations prepared for testcase/submission metadata
- [x] 6.3 Staged rollout runbook created
- [x] 6.4 Rollback steps defined (flag-off/fallback/integrity checks)
- [x] 6.5 Release checklist and acceptance record produced

## Frontend Smoke Test Checklist (Critical Paths)

### Submit Flow

- [ ] Login as student
- [ ] Open problem and submit code
- [ ] Verify submission appears in status list
- [ ] Verify submission transitions to terminal state after judge writeback

### Contest Flow

- [ ] Join encrypted contest with correct password
- [ ] Verify wrong password returns 403 semantic feedback
- [ ] Verify ACM scoreboard ordering and freeze metadata display
- [ ] Verify OI scoreboard total score ordering

### Class Flow

- [ ] Student joins class by code and sees pending state
- [ ] Teacher/assistant reviews membership request
- [ ] Approve path updates member to active
- [ ] Reject path shows deterministic rejected status

### Admin Flow

- [ ] Create/update/delete problem from admin page
- [ ] Create/update/delete contest from admin page
- [ ] Open contest scoreboard and verify pagination + score fields
- [ ] Confirm canonical error feedback for 400/403/404/409/429

## Result Summary

- Overall status: In progress
- Blockers: Testcase and class frontend workflow implementation pending
- Next milestone: complete section 2 and section 5, then rerun full checklist
