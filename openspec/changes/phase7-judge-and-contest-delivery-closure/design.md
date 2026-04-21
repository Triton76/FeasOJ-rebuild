## Context

Three delivery gaps remain after Phase6:

1. Contest product semantics are still inconsistent. Joined students can be tracked as participants, but contest problem listing is effectively owner/admin-only, which breaks real participation.
2. The backend has testcase and contest-binding APIs, but frontend teaching workflows are incomplete, so teachers cannot finish core authoring flows.
3. The rebuild submission pipeline and legacy JudgeCore are not on the same queue contract. Rebuild emits `judge.submission.*` jobs while legacy JudgeCore still consumes the older `judgeTask` queue and writes directly against old repository behavior.

These gaps are coupled. Shipping them piecemeal risks a system that looks complete in API coverage but fails in real teacher/student usage.

## Goals / Non-Goals

**Goals**

- Make contest participation usable for normal students after join, including contest problem visibility.
- Provide teacher/admin UI for contest problem binding and testcase authoring using rebuild APIs.
- Make rebuild submission processing queue-backed with RabbitMQ as a first-class path, not just in-memory fallback.
- Align JudgeCore or an adapter path with rebuild contracts so submissions can reach terminal states through asynchronous judging.
- Produce validation evidence for the full teacher/student workflow.

**Non-Goals**

- Redesigning JudgeCore sandbox execution internals beyond what is needed for contract compatibility.
- Adding advanced judging features such as SPJ, subtasks, or plagiarism detection.
- Reworking unrelated product areas already closed in Phase6 unless integration requires a small compatibility fix.

## Decisions

1. Decision: Treat contest participation, testcase authoring, and judge delivery as one delivery change.
- Rationale: Teachers cannot verify contest or submission behavior unless all three are usable together.

2. Decision: Contest problem listing is visible to contest owner/admin and active joined participants.
- Rationale: This matches actual contest participation requirements while still avoiding unrestricted disclosure.

3. Decision: Teacher-facing contest problem binding will be delivered in frontend admin/management surfaces rather than hidden API-only support.
- Rationale: Existing API coverage without UI has already proven insufficient.

4. Decision: RabbitMQ remains the durable transport for rebuild asynchronous judging when enabled; memory queue stays as explicit local fallback.
- Rationale: Keeps local development lightweight while preserving a production-ready path.

5. Decision: JudgeCore integration will converge on rebuild contracts instead of preserving the legacy queue payload forever.
- Rationale: A permanent dual-contract story would add ongoing maintenance cost and ambiguity.

## Risks / Trade-offs

- [JudgeCore migration risk] Existing legacy queue consumers may not understand rebuild contracts.
  Mitigation: Add contract fixtures, compatibility tests, and staged rollout with RabbitMQ flag-off fallback.

- [Contest visibility regression] Broadening contest problem visibility could accidentally expose problems too early.
  Mitigation: Restrict to active joined participants, owners, and admins, and add regression tests.

- [Frontend scope expansion] Contest/testcase management pages can balloon in UX scope.
  Mitigation: Focus on deterministic CRUD/binding management, not full redesign.

- [Operational ambiguity] RabbitMQ may appear "missing" when disabled by config.
  Mitigation: Make queue mode explicit in config, logs, and runbooks.

## Migration Plan

1. Add Phase7 contract docs and test fixtures for contest participation and judge payloads.
2. Fix contest problem visibility semantics and add frontend contest binding page.
3. Enable testcase frontend workflows against existing rebuild APIs.
4. Align rebuild RabbitMQ queue and JudgeCore consumption/writeback contracts.
5. Run end-to-end validation for teacher and student flows with feature flag matrix.

Rollback:

- Disable RabbitMQ-backed judging and return to in-memory/no-op local mode.
- Disable testcase frontend entry points if backend flags or judge contracts regress.
- Temporarily restrict contest problem listing back to owner/admin only if disclosure bugs are observed.
