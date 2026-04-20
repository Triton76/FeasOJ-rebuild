# Phase3 Monitoring and Rollback Plan

## Feature Flags

- `enable_rabbitmq_queue`
- `enable_embedded_judge_worker`
- `enable_class_workflow_v2`

Flags can be toggled independently for staged rollout.

## Key Metrics

- `queue_enqueue_total`
- `queue_consume_total`
- `queue_retry_total`
- `queue_dlq_total`
- `writeback_success_total`
- `writeback_failure_total`

## Alert Suggestions

- DLQ growth > threshold in 10m window
- writeback failure ratio > threshold
- consume stalls with enqueue growing continuously

## Rollback Trigger Conditions

- sustained writeback failures above SLO
- repeated worker consume failures without recovery
- class workflow conflicts or permission regressions impacting core flow

## Rollback Procedure

1. Disable `enable_embedded_judge_worker`
2. Disable `enable_rabbitmq_queue`
3. Keep API service online and inspect queue backlog
4. Replay from RabbitMQ DLQ/main queue after mitigation

## Post-Rollback Checks

- submit-record create/list still functional
- no silent message loss detected
- error rates back to baseline
