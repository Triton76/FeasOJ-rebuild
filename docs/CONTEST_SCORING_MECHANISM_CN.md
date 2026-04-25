# 当前项目竞赛计分机制说明

本文基于当前仓库中 `backend-rebuild` 的实际实现整理，重点说明项目现在如何进行竞赛计分，以及不同 `rule_type` 的实际行为。

适用范围：

- 新版后端 `services/cmd/app/backend-rebuild`
- 当前日期对应仓库状态下的实现

## 1. 结论先看

当前项目的竞赛 `rule_type` 一共支持三种：

- `acm`
- `oi`
- `assignment`

但从实际计分实现看：

- `acm`：已经有完整榜单聚合逻辑
- `oi`：已经有完整榜单聚合逻辑
- `assignment`：模型和类型已预留，但没有独立计分分支；当前 scoreboard 行为会落到 ACM 逻辑

换句话说，当前“赛制类型支持”和“独立计分规则支持”并不完全等价。

## 2. 通用榜单机制

### 2.1 榜单接口

竞赛榜单通过以下接口提供：

```text
GET /api/v1/contests/:contest_id/scoreboard
```

该接口只会在后端开启 `EnableScoreboard` 时暴露。

### 2.2 榜单生成方式

当前榜单不是预计算快照，也不是单独缓存表，而是查询 `submissions` 表后在线聚合生成。

实现特点：

- 只读取当前竞赛的提交记录
- 按 `submitted_at ASC` 顺序处理
- 根据 `contest.rule_type` 选择榜单聚合规则

### 2.3 可见字段

当前榜单项统一返回以下字段：

- `rank`
- `user_id`
- `username`
- `solved`
- `total_score`
- `penalty_minutes`
- `reached_at`

不同赛制实际会使用不同字段：

- ACM 主要使用 `solved`、`penalty_minutes`、`reached_at`
- OI 主要使用 `total_score`、`reached_at`
- `assignment` 当前没有独立字段语义，实际跟随 ACM 分支

### 2.4 封榜机制

如果竞赛存在 `end_at`，则榜单有统一的最后 60 分钟封榜逻辑：

- 比赛未进入最后 60 分钟：统计到“当前时间”
- 进入最后 60 分钟但比赛未结束：只统计到“封榜开始时刻”
- 比赛结束后：只统计到“比赛结束时刻”

其中封榜开始时间计算为：

```text
freeze_start = end_at - 60 minutes
```

当前实现还存在一个需要注意的边界：

- 查询条件使用的是 `submitted_at < cutoff`
- 因此“恰好等于封榜时刻”或“恰好等于结束时刻”的提交，不会进入可见榜单

## 3. ACM 赛制

## 3.1 核心规则

ACM 榜单按“过题数 + 罚时”进行排名。

系统会按“用户 + 题目”聚合提交：

- 某题第一次 `accepted` 后，该题后续提交不再影响榜单
- `pending`、`judging` 不参与榜单计算
- `compile_error` 单独记为一次 CE 尝试
- 除 `accepted`、`compile_error`、`pending`、`judging` 外，其它结果都记为一次错误尝试

### 3.2 罚时规则

只有“最终做出来的题目”才会累计罚时。

对某道已通过题目，罚时由两部分组成：

1. 错误尝试次数和 CE 次数之和，每次加 20 分钟
2. 从比赛开始到该题首次通过的分钟数

可以写成：

```text
penalty_for_problem = (wrong_attempts + ce_attempts) * 20 + accepted_minutes
```

但有两个边界要注意：

- 如果题目一直没过，则这题不会给总罚时增加任何值
- 当前实现里，`compile_error` 也会像普通错误一样产生 20 分钟罚时，但前提仍然是该题最终被做出

### 3.3 排名规则

ACM 排名顺序如下：

1. `solved` 降序
2. `penalty_minutes` 升序
3. `reached_at` 升序
4. `user_id` 升序

这里的 `reached_at` 不是第一次过题时间，而是：

- 该用户所有已通过题目中，最后一个 AC 的时间

因此它表示“达到当前过题数与罚时状态的时间”。

### 3.4 ACM 示例理解

假设两个人都通过 2 题：

- A：总罚时 80
- B：总罚时 100

则 A 排名更高。

如果两人都过 2 题且总罚时同为 80：

- A 在 10:40 完成第 2 题
- B 在 10:50 完成第 2 题

则 A 排名更高，因为 `reached_at` 更早。

## 4. OI 赛制

### 4.1 核心规则

OI 榜单不是按过题数排名，而是按总分排名。

系统同样按“用户 + 题目”聚合，但统计方式不同：

- 每道题只取该用户的最高得分
- 最终总分等于各题最高得分之和

### 4.2 哪些提交会参与计分

当前实现中：

- `pending`、`judging` 不参与计分
- 其它终态提交都可能参与“每题最高分”比较

这意味着当前 OI 逻辑并不是“只有 `accepted` 才有分”，而是：

- 只要该次提交写回了更高的 `score`
- 即使结果名义上不是 `accepted`
- 也可能成为该题的最佳成绩

这属于当前实现层面的真实行为，和传统 OI 平台是否完全一致，需要按产品预期另行确认。

### 4.3 同题多次提交

同一用户同一题多次提交时，系统会保留：

- `score` 更高的那次提交
- 如果分数相同，则保留提交时间更早的那次

### 4.4 排名规则

OI 排名顺序如下：

1. `total_score` 降序
2. `reached_at` 升序
3. `user_id` 升序

这里的 `reached_at` 含义是：

- 该用户各题“最佳成绩提交时间”中的最大值

可以理解为：

- 用户达到当前总分时刻的时间

### 4.5 OI 示例理解

如果用户 A：

- 题 1 最高分 60
- 题 2 最高分 90

则总分为：

```text
150
```

如果用户 B：

- 题 1 最高分 70
- 题 2 最高分 60

则总分为：

```text
130
```

则 A 排名在 B 前。

## 5. Assignment 赛制

### 5.1 当前状态

`assignment` 已经进入数据模型和枚举：

- 数据库 `contests.rule_type` 允许 `assignment`
- 设计文档也明确第一版内建该类型
- `submissions.score` 也预留说明可用于 `oi / assignment`

### 5.2 当前是否有独立计分实现

没有。

当前 scoreboard 分发逻辑只有一条显式分支：

- 如果 `rule_type == "oi"`，走 OI 聚合
- 否则全部走 ACM 聚合

因此当前 `assignment` 的实际榜单行为是：

- 类型上允许创建
- 参与权限、入场密码、可见性等通用流程
- 但榜单计分没有作业型独立逻辑，实际上会按 ACM 方式生成

### 5.3 这意味着什么

这意味着 `assignment` 目前更像“业务类型预留”，而不是一个已经完整交付的独立赛制。

如果后续要把作业赛制补完整，通常至少还需要明确：

- 是否按总分排名
- 是否按题目最高分聚合
- 是否需要截止时间后封榜或直接显示最终成绩
- 是否需要支持教师手动评分 / 自动评分混合
- 是否需要和 ACM/OI 使用不同的前端榜单展示字段

## 6. 当前实现中的几个关键细节

### 6.1 非终态提交不参与榜单

以下状态当前不会进入有效计分：

- `pending`
- `judging`

### 6.2 榜单是“可见榜单”而非最终历史快照

由于当前榜单读取时会应用封榜时间裁剪，因此在比赛进行中看到的是“当前时刻可见的榜单”，不一定等于最终真实结果。

### 6.3 当前没有看到 assignment 专属测试

当前测试覆盖了：

- ACM 排序规则
- ACM CE 罚时规则
- ACM 封榜行为
- OI 每题最高分聚合
- OI 同分 tie-break

但没有看到 `assignment` 的独立计分测试。

## 7. 一张表看完

| 赛制 | 当前是否可创建 | 当前是否有独立计分逻辑 | 排名主字段 | 备注 |
| --- | --- | --- | --- | --- |
| `acm` | 是 | 是 | `solved`、`penalty_minutes` | 已实现完整榜单 |
| `oi` | 是 | 是 | `total_score` | 已实现每题最高分求和 |
| `assignment` | 是 | 否 | 实际落到 ACM | 仅类型预留，未完成独立作业计分 |

## 8. 主要代码依据

以下文件是本文整理时直接核对过的主要实现来源：

- `services/cmd/app/backend-rebuild/internal/usecase/competitions/service.go`
- `services/cmd/app/backend-rebuild/internal/usecase/competitions/scoreboard_test.go`
- `services/cmd/app/backend-rebuild/internal/repository/competitions/contest_repository_gorm.go`
- `services/cmd/app/backend-rebuild/internal/ports/contracts.go`
- `services/cmd/app/backend-rebuild/migrations/0004_align_contests_phase1_schema.sql`
- `docs/BACKEND_REWRITE_MODEL_DRAFT_CN.md`

如果后续你需要，我可以继续把这份说明再整理成：

- 面向产品的“非代码版说明”
- 面向开发的“规则与边界清单”
- 或补一份“assignment 应该如何设计计分机制”的建议文档
