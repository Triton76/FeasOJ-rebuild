## Why

Phase6 closed several frontend capability gaps, but the product is still not deliverable as a coherent OJ system. The remaining blockers are no longer isolated placeholder issues: they span contest participation semantics, contest-problem management UI, testcase authoring UX, RabbitMQ-backed submission delivery, and JudgeCore integration with rebuild contracts.

The current rebuild can persist submissions and expose selected APIs, but teachers still cannot complete the full workflow of authoring problems, attaching testcase sets, binding problems to contests, publishing contests, having students join, submit, and receive terminal judging results through a stable asynchronous pipeline.

## What Changes

- Fix contest participation and contest-problem visibility semantics so joined users can actually participate and view assigned problems.
- Deliver frontend management workflows for contest problem binding and testcase authoring against rebuild APIs.
- Re-enable and standardize RabbitMQ-backed submission delivery in rebuild with explicit operational toggles and failure behavior.
- Align JudgeCore with rebuild submission job, testcase fetch, and writeback contracts, or provide an explicit compatibility adapter where direct parity is missing.
- Add end-to-end validation for teacher and student product flows: create problem, manage testcase, bind contest, publish, join, submit, judge, and view result/scoreboard.

## Capabilities

### New Capabilities

- `contest-delivery-parity`: Contest participation, problem visibility, and problem-binding management behave consistently for teachers and joined students.
- `testcase-authoring-frontend-closure`: Teachers can manage testcase CRUD/reorder flows from frontend against rebuild APIs.
- `rebuild-rabbitmq-judge-pipeline`: Rebuild can enqueue submissions through RabbitMQ and complete terminal judging via JudgeCore-compatible contracts.
- `judgecore-rebuild-contract-alignment`: JudgeCore consumes rebuild job payloads, fetches testcase data from rebuild endpoints, and writes back signed terminal results.
- `teacher-student-e2e-oj-flow`: The end-to-end OJ flow is validated as one delivery unit across backend, queue, judge, and frontend.

### Modified Capabilities

- `frontend-rebuild-integration-closure`
- `contest-membership-and-quit-parity`
- `testcase-authoring-and-judge-binding`
- `judgecore-end-to-end-writeback`

## Impact

- Affected backend: `services/cmd/app/backend-rebuild/internal/usecase/{competitions,submitrecords,testcases}`, queue adapters, handlers, router, config, and observability.
- Affected judge service: `services/cmd/app/judgecore/**` queue consumption, testcase read path, and writeback contract.
- Affected frontend: `web/src/pages/{Competition,Admin,Problem}`, `web/src/utils/api/*`, and router/navigation.
- Affected operations: RabbitMQ enablement policy, judge callback secret management, testcase API flags, and end-to-end rollout checklist.
