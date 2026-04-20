# JudgeCore Payload Contract V1

## Versioning

- Queue payload field: `contract_version`
- Writeback payload field: `contract_version`
- Current supported value: `v1`

Compatibility policy:
- `contract_version` omitted in writeback is treated as legacy-compatible request and accepted as V1 behavior.
- Explicit unsupported versions are rejected with `400`.

## Queue Job Payload (Backend -> Judge Worker)

```json
{
  "contract_version": "v1",
  "submission_id": 123,
  "user_id": "u-1",
  "problem_id": 101,
  "contest_id": 0,
  "language": "cpp",
  "source_code": "int main(){return 0;}"
}
```

Semantics:
- `submission_id` is the idempotency key for queue delivery.
- Submission is persisted first, then enqueued.

## Writeback Payload (JudgeCore -> Backend)

Endpoint: `POST /api/v1/judge/writeback`

```json
{
  "contract_version": "v1",
  "submission_id": 123,
  "result": "accepted",
  "score": 100,
  "source": "judgecore"
}
```

Validation:
- `submission_id` must be positive.
- `result` must be terminal (`accepted`, `wrong_answer`, `compile_error`, `runtime_error`, `time_limit_exceeded`, `memory_limit_exceeded`, `output_limit_exceeded`, `presentation_error`, `partially_accepted`, `system_error`).
- `score` must be non-negative when present.
- Unsupported `contract_version` is rejected.

State machine:
- Allowed transitions: `pending -> judging -> terminal`
- Invalid transitions return `409`.
- Replayed equivalent terminal writeback is idempotent and returns success.

## Authentication Policy

Optional strict token check:
- Config key: `judge.writeback_token`
- Env key: `BACKEND_REBUILD_JUDGE_WRITEBACK_TOKEN`
- Header: `X-Judge-Token`

Behavior:
- When token is configured, missing/invalid token returns `401`.
- When token is empty, endpoint remains open for trusted internal network usage.
