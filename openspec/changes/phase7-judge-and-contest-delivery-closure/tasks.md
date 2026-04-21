## 1. Contest Delivery Parity

- [x] 1.1 Fix contest-problem visibility so active joined users can read bound problems.
- [x] 1.2 Add regression tests for owner/admin/joined-user/outsider contest-problem access semantics.
- [x] 1.3 Add frontend teacher/admin management flow for contest problem binding.
- [x] 1.4 Expose deterministic frontend states for empty contest bindings, forbidden access, and save conflicts.

## 2. Testcase Authoring Frontend Closure

- [x] 2.1 Add frontend testcase API wrappers for create/list/update/delete/reorder.
- [x] 2.2 Add teacher/admin testcase management page with CRUD and ordering UX.
- [x] 2.3 Add entry points from problem/admin pages into testcase management.
- [ ] 2.4 Add regression coverage for testcase permissions and frontend save/delete flows.

## 3. RabbitMQ Submission Delivery

- [x] 3.1 Make rebuild queue mode explicit in config/run logs and validate RabbitMQ bootstrap behavior.
- [ ] 3.2 Ensure submission enqueue/dequeue semantics are test-covered for RabbitMQ mode.
- [x] 3.3 Preserve local fallback behavior when RabbitMQ is disabled or unavailable.
- [ ] 3.4 Document queue topology, required exchanges/queues, and operational toggles.

## 4. JudgeCore Contract Alignment

- [x] 4.1 Define rebuild submission job fixture consumed by JudgeCore.
- [x] 4.2 Align JudgeCore testcase fetch path to rebuild judge testcase endpoint and auth token policy.
- [x] 4.3 Align JudgeCore terminal result writeback to rebuild `/api/v1/judge/writeback` contract.
- [ ] 4.4 Add compatibility/integration tests for submit -> judging -> terminal writeback.

## 5. End-to-End Product Validation

- [ ] 5.1 Validate teacher flow: create problem, manage testcase, bind to contest, publish.
- [ ] 5.2 Validate student flow: view contest, join, see problems, submit, receive final result.
- [ ] 5.3 Validate scoreboard/result visibility after judged submissions.
- [ ] 5.4 Record rollout checklist, fallback steps, and remaining known gaps.
