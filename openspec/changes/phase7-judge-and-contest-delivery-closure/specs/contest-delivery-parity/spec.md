## ADDED Requirements

### Requirement: Joined users must be able to participate in contests through problem visibility
The system SHALL allow contest owners, admins, and active joined participants to read bound contest problems.

#### Scenario: Joined participant reads contest problems
- **GIVEN** a user has joined a visible contest and their participant status is active for participation
- **WHEN** the user requests contest problems
- **THEN** the backend MUST return bound contest problems instead of rejecting with forbidden

#### Scenario: Outsider cannot read contest problems
- **GIVEN** a user is neither contest owner/admin nor an active joined participant
- **WHEN** the user requests contest problems
- **THEN** the backend MUST reject the request with forbidden

### Requirement: Teachers must have a frontend contest problem binding workflow
The frontend SHALL provide a management workflow for contest problem binding using rebuild contracts.

#### Scenario: Teacher replaces contest bindings
- **WHEN** a teacher or admin edits the contest problem list and saves
- **THEN** the frontend MUST call the rebuild replace-binding endpoint and render the persisted order/alias result
