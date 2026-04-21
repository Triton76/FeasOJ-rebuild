## ADDED Requirements

### Requirement: Teachers must be able to manage testcase sets from frontend
The frontend SHALL expose testcase CRUD and reorder workflows for teacher/admin users against rebuild testcase APIs.

#### Scenario: Teacher edits testcase set
- **WHEN** a teacher opens testcase management for a problem
- **THEN** the frontend MUST list existing testcase items, allow create/update/delete/reorder, and persist changes through rebuild APIs

#### Scenario: Student is blocked from testcase management
- **WHEN** a non-teacher non-admin user accesses testcase management APIs or UI
- **THEN** the system MUST reject backend mutation requests and the frontend MUST not present privileged actions
