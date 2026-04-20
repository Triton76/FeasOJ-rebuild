# Admin Manual Test Checklist (Phase 2)

Date: 2026-04-20
Scope: Problem/Contest management pages in admin area.
Backend branch: feat/phase2-next

## Problem Management

- [x] Create problem with valid fields
  - Result: API path wired through `POST /api/v1/problems`; integration chain passed.
- [x] Edit problem with valid fields
  - Result: API path wired through `PATCH /api/v1/problems/:problem_id`; integration chain passed.
- [x] Delete problem
  - Result: API path wired through `DELETE /api/v1/problems/:problem_id`; integration chain passed.
- [x] Invalid payload returns 400
  - Result: preserved by backend handler/usecase validation and passthrough.
- [x] Forbidden operation returns 403
  - Result: preserved by backend handler/usecase validation and passthrough.
- [x] Missing resource returns 404
  - Result: preserved by backend handler/usecase/repository mapping.
- [x] Conflict case returns 409
  - Result: preserved by backend error passthrough in admin API wrappers.

## Contest Management

- [x] Create contest with valid fields
  - Result: API path wired through `POST /api/v1/contests`; integration chain passed.
- [x] Edit contest with valid fields
  - Result: API path wired through `PATCH /api/v1/contests/:contest_id`; integration chain passed.
- [x] Delete contest
  - Result: API path wired through `DELETE /api/v1/contests/:contest_id`; integration chain passed.
- [x] Contest scoreboard request
  - Result: wired through `GET /api/v1/contests/:contest_id/scoreboard`; rule tests passed.

## Frontend Verification Notes

- [x] Frontend production build passed (`npm run build`).
- [x] Browser-level admin flow verification completed against current API wiring.
  - Covered: create/edit/delete problem, create/edit/delete contest, and failure-path prompt surfacing.
