## ADDED Requirements

### Requirement: Rebuild must support RabbitMQ-backed submission delivery as a first-class judging mode
The rebuild backend SHALL enqueue submission jobs through RabbitMQ when queue mode is enabled and SHALL expose explicit fallback behavior when queue mode is disabled.

#### Scenario: RabbitMQ mode enabled
- **GIVEN** RabbitMQ queue mode is enabled and broker connectivity is healthy
- **WHEN** a submission is created
- **THEN** the backend MUST publish a rebuild judge job payload to the configured RabbitMQ topology

#### Scenario: RabbitMQ mode unavailable
- **GIVEN** RabbitMQ queue mode is disabled or broker bootstrap fails
- **WHEN** the backend starts
- **THEN** the backend MUST log the effective fallback mode and continue using the configured non-RabbitMQ fallback

### Requirement: JudgeCore must complete rebuild terminal writeback lifecycle
JudgeCore SHALL consume rebuild-compatible jobs, fetch testcase data, and write back terminal results to rebuild.

#### Scenario: Submission reaches terminal result
- **WHEN** JudgeCore completes judging for a submission
- **THEN** it MUST call rebuild writeback with the agreed contract and the submission MUST transition from `judging` to a terminal state
