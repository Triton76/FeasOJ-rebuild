## 1. Baseline Fact Lock & Scope Freeze

- [x] 1.1 [P0][Depends: none][AC: 形成已确认/未知事实表并经评审确认] Consolidate evidence matrix for A-E scope and attach repository citations.
- [x] 1.2 [P0][Depends: 1.1][AC: unknown 项全部有验证路径与负责人] Freeze unknown list with validation tasks and owners.
- [x] 1.3 [P0][Depends: 1.2][AC: proposal/design/specs reviewed and no speculative statements remain] Complete scope freeze review and sign-off.

## 2. Contest-Problem Binding Closure

- [x] 2.1 [P0][Depends: 1.3][AC: 新增 contest-problem 读写契约通过接口评审] Add contest-problem binding contracts and DTOs.
- [x] 2.2 [P0][Depends: 2.1][AC: 完成绑定查询/全量更新 API 且具备权限校验] Implement binding query and batch replace APIs.
- [x] 2.3 [P0][Depends: 2.2][AC: 发布前无绑定题目时返回冲突错误] Add publish gate validation requiring at least one bound problem.
- [x] 2.4 [P1][Depends: 2.2][AC: 唯一性/非法引用场景测试通过] Add integrity tests for duplicate problem/order/alias and invalid references.

## 3. Class Join Usability & Feedback

- [ ] 3.1 [P0][Depends: 1.3][AC: class_code 入参与 pending 审核态在接口文档中锁定] Freeze class join/review API semantics and response model.
- [ ] 3.2 [P0][Depends: 3.1][AC: 申请、审核、重复申请、权限不足场景全部有稳定错误码] Implement deterministic status/error mapping.
- [ ] 3.3 [P1][Depends: 3.2][AC: 前端页面可展示 pending/active/rejected/revoked 全状态] Wire frontend state rendering for membership lifecycle.
- [ ] 3.4 [P1][Depends: 3.2][AC: 班级加入与审核链路集成测试通过] Add end-to-end tests for join and review feedback loop.

## 4. Account Security Assist Decision & Landing

- [x] 4.1 [P0][Depends: 1.2][AC: 明确“接入或下线”决策并固化为配置策略] Finalize policy for verification-code/password-reset capability.
- [x] 4.2 [P0][Depends: 4.1][AC: enabled/disabled 两种策略均有确定行为和错误码] Implement policy-driven endpoint behavior.
- [x] 4.3 [P0][Depends: 4.2][AC: 验证码过期/错误/限流场景测试通过] Implement verification code constraints and validation paths.
- [ ] 4.4 [P1][Depends: 4.2][AC: 密码重置成功后旧凭证失效策略验证通过] Add credential invalidation handling and tests.

## 5. Admin Role/Status Governance Completion

- [x] 5.1 [P0][Depends: 1.3][AC: 新增角色治理 API 合约并完成权限评审] Add admin role update contract separated from status update.
- [x] 5.2 [P0][Depends: 5.1][AC: 非 admin 调用返回 forbidden，admin 调用成功可追踪] Implement role update handler/usecase/repository path.
- [ ] 5.3 [P1][Depends: 5.2][AC: 角色/状态操作均写入结构化审计记录] Add governance audit logging fields and persistence.
- [ ] 5.4 [P1][Depends: 5.2][AC: 角色变更与状态变更互不污染的测试通过] Add separation tests for role vs status operations.

## 6. Config & Ops Entry Convergence

- [ ] 6.1 [P0][Depends: 1.3][AC: backend/backend-rebuild/judgecore 的配置优先级文档统一] Define deterministic config precedence policy (ENV vs file).
- [ ] 6.2 [P0][Depends: 6.1][AC: 三服务启动日志均输出生效配置源] Add runtime config source reporting and fail-fast checks.
- [ ] 6.3 [P1][Depends: 6.1][AC: 兼容窗口内 YAML/TOML 迁移有告警且可运行] Implement migration-safe compatibility mode and deprecation warnings.
- [ ] 6.4 [P1][Depends: 6.2][AC: 部署手册更新并经演练通过] Update deployment/runbook with rollback toggles and entry commands.

## 7. Integration, Rollout, and Rollback Readiness

- [ ] 7.1 [P0][Depends: 2.4,3.4,4.4,5.4,6.4][AC: 五大闭环能力回归测试通过] Execute integrated regression checklist for A-E scope.
- [ ] 7.2 [P0][Depends: 7.1][AC: 灰度开关、回滚开关和监控指标在发布前确认] Validate rollout/rollback controls in staging.
- [ ] 7.3 [P0][Depends: 7.2][AC: 发布评审结论为可进入实施] Complete implementation readiness review and handoff.
