# Phase3 API Contract Freeze

This document freezes the backend API and error semantics for Phase3 integration.

## Frozen Error Semantics

All Phase3 APIs MUST keep the following status code meanings:

- 400: invalid argument / malformed request
- 403: resource is visible but action is forbidden
- 404: resource not found or not visible to current actor
- 409: state conflict (duplicate apply, duplicate review, illegal transition)
- 429: rate limited

422 is intentionally not used in Phase3.

## Class Domain Endpoints

- POST /api/v1/classes
- PATCH /api/v1/classes/:class_id
- POST /api/v1/classes/:class_id/archive
- POST /api/v1/classes/join
- POST /api/v1/classes/memberships/review
- GET /api/v1/classes/:class_id/memberships
- GET /api/v1/classes/memberships/self

## Queue / Submission Endpoints

- POST /api/v1/submit-records
- GET /api/v1/submit-records
- POST /api/v1/judge/writeback

## Versioning and Compatibility

- Existing response envelope remains unchanged: `{ data: ... }` for success and `{ error: ... }` for failure.
- New backend features are gated by feature flags where applicable.
- Frontend adapters should treat 409 and 429 as actionable states (conflict/retry).
