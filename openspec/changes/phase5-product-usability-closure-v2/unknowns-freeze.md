# Unknowns Freeze and Validation Ownership

## Unknown-1
- topic: frontend class membership full-state rendering (`pending/active/rejected/revoked`)
- validation method: inspect class-related frontend API usage and state mapping views
- owner: FE maintainers
- target output: validated state mapping list + gaps

## Unknown-2
- topic: production traffic split and runtime entry between legacy backend and backend-rebuild
- validation method: inspect deployment scripts, gateway routing, runtime logs in staging/production
- owner: SRE/ops
- target output: service routing map and cutover sequence

## Unknown-3
- topic: final product decision for account security assist (enable in rebuild vs deprecate)
- validation method: product/security decision record in release checklist
- owner: PM + Security
- target output: signed policy and rollout flag default
