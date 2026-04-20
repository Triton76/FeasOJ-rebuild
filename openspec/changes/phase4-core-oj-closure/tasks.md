## 1. JudgeCore End-to-End Lifecycle

- [x] 1.1 Define and document judge job payload and writeback payload versioned contract
- [x] 1.2 Implement authenticated `/api/v1/judge/writeback` verification policy and validation middleware
- [x] 1.3 Complete submission worker/consumer lifecycle to ensure `pending -> judging -> terminal` state flow
- [x] 1.4 Add idempotent terminal writeback handling and invalid-transition conflict checks
- [x] 1.5 Add structured logs and metrics for enqueue/dequeue/writeback success, failure, and latency
- [x] 1.6 Add integration tests for submit enqueue, consume transition, writeback success, writeback replay, and auth failure

## 2. Testcase Authoring and Judge Binding

- [x] 2.1 Add testcase domain contracts (DTOs/ports) and teacher-admin authorization rules
- [x] 2.2 Implement testcase repository CRUD with deterministic ordering and sample flag support
- [x] 2.3 Add testcase HTTP handlers/routes for create, list, update, delete, and reorder operations
- [x] 2.4 Implement testcase payload validation (required fields, size limits, count limits)
- [x] 2.5 Implement judge-facing testcase read endpoint/service for authoritative active testcase sets
- [x] 2.6 Add unit/integration tests for permissions, validation, ordering, and judge-read behavior

## 3. Contest ACM/OI Mode Parity

- [x] 3.1 Refactor scoreboard service to branch deterministic logic by `rule_type`
- [x] 3.2 Implement ACM ranking semantics (solved, penalty, tie-break) and freeze metadata behavior
- [x] 3.3 Implement OI ranking semantics (per-problem best score aggregation and tie-break)
- [x] 3.4 Ensure contest visibility/join policy remains consistent across modes and encryption settings
- [x] 3.5 Add golden-case scoreboard fixtures and tests for ACM freeze edge cases and OI scoring

## 4. Frontend-Backend Contract Alignment

- [x] 4.1 Remove legacy field shims from `web/src/utils/api/*` for submit, contest, class, and admin flows
- [x] 4.2 Update status/profile/admin pages to consume canonical rebuild DTO fields
- [x] 4.3 Fix admin contest scoreboard parsing and pagination against canonical backend payload
- [x] 4.4 Replace fabricated local-only contest/class fallbacks with backend API-driven data
- [x] 4.5 Normalize frontend error handling for 400/401/403/404/409/429 semantics
- [x] 4.6 Add frontend smoke tests/manual checklist for submit, contest, class, and admin critical paths

## 5. Class Management Frontend Workflows

- [x] 5.1 Add class navigation entry and route structure for class management and membership workflows
- [x] 5.2 Implement teacher class create/edit/archive pages and API integration
- [x] 5.3 Implement student class join and membership status pages
- [x] 5.4 Implement assistant/teacher membership review interface with role-aware action visibility
- [x] 5.5 Add end-to-end UI checks for class lifecycle and membership approval/rejection flows

## 6. Rollout, Migration, and Operational Safety

- [x] 6.1 Add feature flags for judge writeback strict mode, testcase APIs, and mode-specific scoreboard behavior
- [x] 6.2 Prepare additive database migrations for testcase and submission metadata needs
- [x] 6.3 Create staged rollout runbook (staging verification, canary enablement, production switch)
- [x] 6.4 Define rollback steps (flag-off sequence, service fallback, data integrity checks)
- [x] 6.5 Produce release checklist and acceptance record covering all new capabilities