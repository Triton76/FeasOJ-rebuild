# Schema Migration 范例（FeasOJ）

> 说明：这是“结构版本管理”范例，不涉及旧数据迁移。
> 
> 目标：演示如何同时维护“可执行迁移文件”和“人类可读变更记录”。

---

## 1. 目录建议

- `services/cmd/app/backend-rebuild/migrations/0001_init_schema.sql`
- `services/cmd/app/backend-rebuild/migrations/0002_add_contest_status_scan_index.sql`
- `docs/SCHEMA_MIGRATION_SAMPLE_CN.md`（本文件，作为记录样例）

---

## 2. 命名规范

- 使用递增编号前缀：`0001_`, `0002_`, `0003_` ...
- 文件名描述动作与对象：`add_xxx`, `drop_xxx`, `alter_xxx`
- 已执行版本不修改，只追加新版本

---

## 3. 变更记录模板

每个迁移版本在文档中记录一条：

- 版本号：
- 文件：
- 目的：
- 变更内容：
- 兼容性：向后兼容 / 不兼容
- 回滚方案：
- 风险备注：

---

## 4. 示例记录

### 0001
- 版本号：`0001`
- 文件：`services/cmd/app/backend-rebuild/migrations/0001_init_schema.sql`
- 目的：初始化 schema 与核心示例表
- 变更内容：创建 `schema_migrations`、`contests`
- 兼容性：首次建库
- 回滚方案：删除相关表（仅开发环境）
- 风险备注：示例仅演示最小字段，不代表全量生产模型

### 0002
- 版本号：`0002`
- 文件：`services/cmd/app/backend-rebuild/migrations/0002_add_contest_status_scan_index.sql`
- 目的：优化竞赛状态定时扫描查询
- 变更内容：新增 `contests(status, start_at, end_at)` 复合索引
- 兼容性：向后兼容
- 回滚方案：删除新增索引
- 风险备注：索引命名需与实际 SQL 保持一致

### 0003
- 版本号：`0003`
- 文件：`services/cmd/app/backend-rebuild/migrations/0003_create_phase1_core_tables.sql`
- 目的：按重写草案落地第一阶段核心实体
- 变更内容：创建 `users`、`classes`、`class_memberships`、`problems`、`test_cases`、`contest_problems`、`discussions`、`comments`、`contest_participants`、`submissions`
- 兼容性：向后兼容（新增表）
- 回滚方案：开发环境可按依赖顺序删除新增表
- 风险备注：采用 ENUM + 轻约束策略，复杂业务规则不在数据库层实现

### 0004
- 版本号：`0004`
- 文件：`services/cmd/app/backend-rebuild/migrations/0004_align_contests_phase1_schema.sql`
- 目的：将示例 `contests` 结构升级到第一阶段正式模型
- 变更内容：补充 `owner_user_id`、`class_id`、`visibility`、`rule_type`、`is_encrypted`、`password_hash`、`auto_score` 等字段并增加索引
- 兼容性：向后兼容（扩展字段）
- 回滚方案：删除新增字段与索引（仅开发环境建议）
- 风险备注：旧示例数据可能不满足新字段默认语义，需在接口层做状态收敛与校验

---

## 5. 执行流程建议

1. 新增一个迁移文件（只写本次变更）
2. 在本文件追加一条记录
3. 本地执行并验证
4. 代码评审后合并

这样即使是课程设计、允许破坏性重构，也能保持结构演进可追踪。
