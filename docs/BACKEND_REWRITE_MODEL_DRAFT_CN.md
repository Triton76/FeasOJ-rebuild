# FeasOJ 后端全量重写：实体模型草案 v1

> 目标：只定义第一阶段的实体模型、字段、关系、约束与兼容边界。
> 
> 本草案用于指导后续 agent 直接搭建数据库模型、ORM 结构、DTO 骨架、repository/service 空框架。
> 
> 当前阶段**不实现业务逻辑**，也**不复用旧后端代码**。

---

## 1. 本草案的边界

### 1.1 本草案覆盖
- 核心实体划分
- 各实体字段定义
- 主键策略
- 枚举值建议
- 实体关系
- 唯一约束 / 索引建议
- 与前端重构协作相关的模型边界
- 第一阶段可实施范围

### 1.2 本草案不覆盖
- 权限判定逻辑实现
- 竞赛记分逻辑实现
- 判题逻辑实现
- HTTP 路由与响应结构实现细节
- handler / controller 的业务细节
- 前端页面改造方案
- 旧数据库迁移方案

### 1.3 明确前提
- 本项目是课程设计，可以**完全丢弃旧库**，无需考虑迁移。
- 可以进行**破坏性重构**。
- 第一阶段只聚焦：**实体模型与结构骨架**。
- HTTP 路由、响应结构与前端联调不属于第一阶段交付物。
- 后续进入接口阶段时，前端允许大改，仅需保持页面风格与核心信息架构大致一致。
- 数据库第一阶段明确以 **MySQL** 作为主数据库。

---

## 2. 已拍板的重写决策

### 2.1 主键策略
- `users` / `classes` / `class_memberships` / `test_cases` / `contest_problems` / `contest_participants` 使用 **UUID**
- `problems` / `contests` / `submissions` 使用 **bigint 自增主键**

### 2.2 用户角色模型
- 使用单一 `users.role` 字段
- 不单独建立 `student` / `teacher` / `admin` 主实体
- `role` 使用**字符串枚举**：
  - `student`
  - `teacher`
  - `admin`
- `users.status` 简化为：
  - `active`
  - `banned`
- `username` 第一版不允许修改
- `email` 第一版必须存在且全局唯一

### 2.3 班级成员模型
- 使用统一的 `class_memberships` 表
- 通过 `role_in_class` 表示用户在班级中的身份
- 第一版一个班级只允许**一个教师**
- 通过数据库唯一索引保证“一个班级仅一个教师”

### 2.4 资源归属
- `problems`：教师拥有，`class_id` 可空，可选绑定班级
- `contests`：教师拥有，`class_id` 可空，可选绑定班级
- `problems.visibility` 保留三档：`public` / `class` / `private`
- `problems.status` 保留三档：`draft` / `published` / `archived`
- `contests.visibility` 同样保留三档：`public` / `class` / `private`
- 题目与竞赛关系底层统一使用 `contest_problems` 关联表
- 第一阶段**不**保留单个 `competition_id` 兼容字段；若迁移期页面仍依赖该字段，需要同步修改前端读写方式

### 2.5 兼容策略
- 第一阶段只定义模型层约束，不冻结 HTTP 设计
- 不引入 `display_name`，第一版统一以 `username` 作为登录名与展示名
- `time_limit` / `memory_limit` 推荐以新模型语义输出，单位格式在接口阶段统一拍板
- 用户/题目/竞赛/讨论区/提交等 DTO 字段以新接口契约为准
- 前端按新接口进行重构，必要时在过渡期提供少量临时映射
- 具体 HTTP 路由与响应结构在接口阶段单独设计，不以旧前端契约为默认目标

### 2.6 第一版赛制范围
- 第一版内建 `rule_type`：
  - `acm`
  - `oi`
  - `assignment`
- 已确认：`assignment` 与 `contest` **共用同一套实体模型**，不单独拆分教学作业主实体。

### 2.7 班级加入方式
- 第一版采用：**申请制 + 教师审核**
- `classes.code` 既是班级展示编号，也是学生发起加入申请时使用的加入码
- `classes.name` 允许重复，唯一标识以 `code` 为准
- `classes.status` 第一版只保留：`active`
- `class_memberships.status` 保留：`pending` / `active` / `rejected` / `removed`
- 因此 `class_memberships.status` 必须支持待审核状态

### 2.8 竞赛密码能力
- 第一版保留密码竞赛能力
- `contests.status` 简化为：
  - `draft`
  - `scheduled`
  - `running`
  - `ended`
- 迁移窗口内的数字映射约定为：
  - `draft = 0`
  - `scheduled = 1`
  - `running = 2`
  - `ended = 3`
- `contests` 需要保留：
  - `is_encrypted`
  - `password_hash`

### 2.8.1 竞赛状态驱动方式
- `contests.status` 由时间自动驱动，不以手工改状态作为主流程
- 推荐规则：
  - `draft`：未发布草稿，可不填 `start_at` / `end_at`
  - `scheduled`：已发布且 `now < start_at`
  - `running`：`start_at <= now < end_at`
  - `ended`：`now >= end_at`

### 2.9 题目用例模型
- 第一版将 `test_cases` 一并纳入模型草案
- 第一版**不引入** `case_type`
- 继续保留 `is_sample` 布尔字段即可

### 2.10 提交源码存储
- 第一版 `submissions.source_code` 直接存数据库

### 2.11 提交状态枚举
- `submissions.result` 使用固定枚举值
- 第一版推荐枚举：
  - `pending`
  - `judging`
  - `accepted`
  - `wrong_answer`
  - `compile_error`
  - `runtime_error`
  - `time_limit_exceeded`
  - `memory_limit_exceeded`
  - `output_limit_exceeded`
  - `presentation_error`
  - `partially_accepted`
  - `system_error`
- `submissions.score` 保留为可空整型，用于 OI / assignment 赛制

### 2.12 账号来源
- 第一版只考虑：**学生自主注册**
- 暂不考虑教师创建学生账号功能
- 教师账号只允许由**管理员创建**
- 管理员账号在系统初始化时创建
- 教师不开放自主注册

---

## 3. 总体建模原则

### 3.1 统一身份主体
系统中所有登录用户统一为 `users`。

### 3.2 不提前引入复杂 RBAC
第一阶段不设计 `permissions` / `role_permissions` / `policies` 等细粒度权限表。

第一阶段权限模型只为后续实现预留以下三个基础维度：
- 用户全局角色：`users.role`
- 班级归属关系：`class_memberships`
- 资源归属字段：`owner_user_id` / `class_id`

### 3.3 资源必须可归属
所有需要管理权限的核心资源都必须具备明确归属信息。

### 3.4 DTO 新契约优先，底层模型不强求兼容旧库
数据库字段名按新模型定义，DTO 层以新接口契约为准。
如需短期迁移，可增加可回收的临时兼容映射，但不作为长期约束。

---

## 4. 实体总览

### 身份与组织
1. `users`
2. `classes`
3. `class_memberships`

### 资源
4. `problems`
5. `test_cases`
6. `contests`
7. `contest_problems`
8. `discussions`
9. `comments`

### 运行数据
10. `contest_participants`
11. `submissions`

---

## 5. 详细实体定义

---

## 5.1 users

### 职责
系统统一登录主体，保存公共用户资料与全局角色信息。

### 字段

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| id | uuid | 是 | 主键 |
| username | varchar(64) | 是 | 用户名，全局唯一 |
| email | varchar(128) | 是 | 邮箱，全局唯一 |
| password_hash | varchar(255) | 是 | 密码哈希 |
| role | enum/string | 是 | `student` / `teacher` / `admin` |
| avatar | varchar(255) | 否 | 头像地址 |
| synopsis | text | 否 | 个人简介 |
| score | int | 是 | 用户总分（排行榜与管理页核心字段） |
| status | enum/string | 是 | `active` / `banned` |
| created_at | datetime | 是 | 创建时间 |
| updated_at | datetime | 是 | 更新时间 |

### 约束
- `username` 唯一
- `email` 唯一
- `role` 非空
- `score` 非空，默认值建议为 `0`
- `status` 非空

### 索引建议
- unique index: `username`
- unique index: `email`
- index: `role`
- index: `status`

### 接口对齐说明
用户 DTO 以新契约为准，建议至少提供：

```json
{
  "id": "uuid",
  "username": "alice",
  "email": "alice@example.com",
  "avatar": "...",
  "synopsis": "...",
  "score": 0,
  "role": "student",
  "created_at": "..."
}
```

---

## 5.2 classes

### 职责
班级实体，由教师创建并拥有。

### 字段

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| id | uuid | 是 | 主键 |
| name | varchar(128) | 是 | 班级名称 |
| code | varchar(32) | 是 | 班级展示编号，同时也是学生发起加入申请时使用的加入码，唯一 |
| description | text | 否 | 班级描述 |
| owner_user_id | uuid | 是 | 班级教师，唯一教师 |
| status | enum/string | 是 | `active` |
| created_at | datetime | 是 | 创建时间 |
| updated_at | datetime | 是 | 更新时间 |

### 约束
- `code` 唯一
- `owner_user_id` 非空
- `status` 非空

### 索引建议
- unique index: `code`
- index: `owner_user_id`
- index: `status`

### 说明
- 第一版一个班级只有一个教师，因此 `owner_user_id` 即班级教师。
- `class_memberships` 中仍保留 `role_in_class`，以统一表达师生关系。

---

## 5.3 class_memberships

### 职责
表示用户与班级之间的关系，包括教师关系和学生申请/加入关系。

### 字段

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| id | uuid | 是 | 主键 |
| class_id | uuid | 是 | 班级 ID |
| user_id | uuid | 是 | 用户 ID |
| role_in_class | enum/string | 是 | `teacher` / `student` |
| status | enum/string | 是 | `pending` / `active` / `rejected` / `removed` |
| joined_at | datetime | 否 | 正式加入时间 |
| created_at | datetime | 是 | 创建时间 |
| updated_at | datetime | 是 | 更新时间 |

### 约束
- `(class_id, user_id)` 唯一
- `role_in_class` 非空
- `status` 非空

### 索引建议
- unique index: `(class_id, user_id)`
- index: `user_id`
- index: `(class_id, role_in_class, status)`
- unique index（条件唯一）: 教师记录每班仅 1 条
  - MySQL 推荐实现：增加生成列 `teacher_class_id = IF(role_in_class = 'teacher', class_id, NULL)`，并对 `teacher_class_id` 建 unique index

### 说明
第一版建议约定：
- 教师创建班级后，自动存在一条 teacher membership 记录
- 学生申请加入班级时先写入：`status = pending`
- 教师审核通过后改为：`status = active`

---

## 5.4 problems

### 职责
题目主实体。

### 字段

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| id | bigint | 是 | 主键，自增 |
| title | varchar(255) | 是 | 标题 |
| content | longtext/text | 是 | 题面内容 |
| input | text | 否 | 输入说明 |
| output | text | 否 | 输出说明 |
| difficulty | int | 是 | 难度，第一版保留整数形式以便前端理解 |
| time_limit_ms | int | 是 | 时间限制（毫秒） |
| memory_limit_mb | int | 是 | 内存限制（MB） |
| owner_user_id | uuid | 是 | 题目拥有者 |
| class_id | uuid | 否 | 若是班级题则绑定班级 |
| visibility | enum/string | 是 | `public` / `class` / `private` |
| status | enum/string | 是 | `draft` / `published` / `archived` |
| created_at | datetime | 是 | 创建时间 |
| updated_at | datetime | 是 | 更新时间 |

### 约束
- `owner_user_id` 非空
- `visibility` 非空
- `status` 非空

### 索引建议
- index: `owner_user_id`
- index: `class_id`
- index: `(visibility, status)`
- index: `difficulty`

### 接口对齐说明
题目 DTO 建议稳定提供这些字段语义：

```json
{
  "id": 1,
  "title": "A + B Problem",
  "content": "...",
  "difficulty": 0,
  "time_limit": "1000ms",
  "memory_limit": "256MB",
  "input": "...",
  "output": "..."
}
```

说明：数据库可存 `time_limit_ms` / `memory_limit_mb`，对外字段格式由新接口契约统一定义。

---

## 5.5 test_cases

### 职责
题目测试用例实体。

### 字段

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| id | uuid | 是 | 主键 |
| problem_id | bigint | 是 | 所属题目 |
| input_data | longtext/text | 是 | 输入数据 |
| output_data | longtext/text | 是 | 输出数据 |
| is_sample | bool | 是 | 是否为公开样例 |
| sort_order | int | 是 | 展示/执行顺序 |
| created_at | datetime | 是 | 创建时间 |
| updated_at | datetime | 是 | 更新时间 |

### 约束
- `problem_id` 非空
- `sort_order` 非空

### 索引建议
- index: `problem_id`
- index: `(problem_id, is_sample)`
- index: `(problem_id, sort_order)`

### 说明
- 第一版不区分更多高级类型（如 pretest/system test/subtask）
- 若后续支持 OI 子任务计分，可在之后扩展，不在本阶段引入

---

## 5.6 contests

### 职责
竞赛主实体。

### 字段

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| id | bigint | 是 | 主键，自增 |
| title | varchar(255) | 是 | 标题 |
| subtitle | varchar(255) | 否 | 副标题 |
| description | longtext/text | 否 | 详细描述 |
| announcement | longtext/text | 否 | 公告 |
| owner_user_id | uuid | 是 | 竞赛拥有者 |
| class_id | uuid | 否 | 若为班级竞赛则绑定班级 |
| visibility | enum/string | 是 | `public` / `class` / `private` |
| rule_type | enum/string | 是 | `acm` / `oi` / `assignment` |
| status | enum/string | 是 | `draft` / `scheduled` / `running` / `ended` |
| is_encrypted | bool | 是 | 是否需要密码 |
| password_hash | varchar(255) | 否 | 竞赛密码哈希 |
| auto_score | bool | 是 | 是否自动结算 |
| start_at | datetime | 否 | 开始时间 |
| end_at | datetime | 否 | 结束时间 |
| created_at | datetime | 是 | 创建时间 |
| updated_at | datetime | 是 | 更新时间 |

### 约束
- `owner_user_id` 非空
- `rule_type` 非空
- `status` 非空
- `is_encrypted` 非空
- `auto_score` 非空
- 状态由时间自动驱动：非 `draft` 状态下 `start_at` / `end_at` 必填，且 `start_at < end_at`

### 索引建议
- index: `owner_user_id`
- index: `class_id`
- index: `(visibility, status)`
- index: `rule_type`
- index: `(start_at, end_at)`

### 接口对齐说明
竞赛 DTO 建议稳定提供这些核心字段语义：

```json
{
  "id": 1,
  "title": "Contest 1",
  "subtitle": "...",
  "announcement": "...",
  "status": 1,
  "encrypted": true,
  "start_at": "...",
  "end_at": "..."
}
```

说明：
- 底层 `status` 使用字符串枚举。
- 若迁移窗口内存在历史页面依赖，可临时映射：`draft=0`、`scheduled=1`、`running=2`、`ended=3`。
- `encrypted` 可由 `is_encrypted` 映射，但建议前端逐步收敛到新字段命名。

---

## 5.7 contest_problems

### 职责
竞赛与题目的关联关系。

### 字段

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| id | uuid | 是 | 主键 |
| contest_id | bigint | 是 | 竞赛 ID |
| problem_id | bigint | 是 | 题目 ID |
| display_order | int | 是 | 展示顺序 |
| alias | varchar(16) | 否 | 题号，如 A/B/C |
| created_at | datetime | 是 | 创建时间 |
| updated_at | datetime | 是 | 更新时间 |

### 约束
- `(contest_id, problem_id)` 唯一
- `display_order` 非空

### 索引建议
- unique index: `(contest_id, problem_id)`
- unique index: `(contest_id, display_order)`
- optional unique index: `(contest_id, alias)`

---

## 5.8 discussions

### 职责
讨论帖子实体，用于兼容现有讨论区前端与后续重写后的讨论模块。

### 字段

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| id | uuid | 是 | 主键 |
| title | varchar(255) | 是 | 标题 |
| content | longtext/text | 是 | 内容 |
| user_id | uuid | 是 | 发帖用户 |
| created_at | datetime | 是 | 创建时间 |
| updated_at | datetime | 是 | 更新时间 |

### 约束
- `title` 非空
- `content` 非空
- `user_id` 非空

### 索引建议
- index: `user_id`
- index: `created_at`

### 接口对齐说明
对外建议稳定提供字段：`id` / `title` / `content` / `user_id` / `username` / `avatar` / `created_at`。

---

## 5.9 comments

### 职责
讨论评论实体。

### 字段

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| id | uuid | 是 | 主键 |
| discussion_id | uuid | 是 | 所属讨论帖 |
| content | longtext/text | 是 | 评论内容 |
| user_id | uuid | 是 | 评论用户 |
| profanity | bool | 是 | 是否命中敏感词（讨论审核辅助字段） |
| created_at | datetime | 是 | 创建时间 |
| updated_at | datetime | 是 | 更新时间 |

### 约束
- `discussion_id` 非空
- `content` 非空
- `user_id` 非空
- `profanity` 非空

### 索引建议
- index: `discussion_id`
- index: `user_id`
- index: `created_at`

### 接口对齐说明
对外建议稳定提供字段：`id` / `discussion_id` / `content` / `user_id` / `username` / `avatar` / `created_at` / `profanity`。

---

## 5.10 contest_participants

### 职责
表示用户参与竞赛的关系。

### 字段

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| id | uuid | 是 | 主键 |
| contest_id | bigint | 是 | 竞赛 ID |
| user_id | uuid | 是 | 用户 ID |
| status | enum/string | 是 | `registered` / `active` / `quit` / `finished` |
| joined_at | datetime | 否 | 加入时间 |
| created_at | datetime | 是 | 创建时间 |
| updated_at | datetime | 是 | 更新时间 |

### 约束
- `(contest_id, user_id)` 唯一
- `status` 非空

### 索引建议
- unique index: `(contest_id, user_id)`
- index: `(contest_id, status)`
- index: `user_id`

### 说明
- 第一版先只表达“参与关系”
- 不在此表中预先加入复杂榜单字段
- 比赛得分、排名、罚时等留待后续业务阶段决定是否单独建 standings 表

---

## 5.11 submissions

### 职责
提交记录实体。

### 字段

| 字段 | 类型 | 必填 | 说明 |
|---|---|---:|---|
| id | bigint | 是 | 主键，自增 |
| user_id | uuid | 是 | 提交用户 |
| problem_id | bigint | 是 | 题目 ID |
| contest_id | bigint | 否 | 若为竞赛提交则关联竞赛 |
| language | varchar(32) | 是 | 编程语言 |
| source_code | longtext/text | 是 | 源码 |
| result | enum/string | 是 | 判题结果，固定枚举值 |
| score | int | 否 | OI/作业赛制得分，可为空 |
| submitted_at | datetime | 是 | 提交时间 |
| created_at | datetime | 是 | 创建时间 |
| updated_at | datetime | 是 | 更新时间 |

### 约束
- `user_id` 非空
- `problem_id` 非空
- `language` 非空
- `result` 非空
- `result` 取值限定为：`pending` / `judging` / `accepted` / `wrong_answer` / `compile_error` / `runtime_error` / `time_limit_exceeded` / `memory_limit_exceeded` / `output_limit_exceeded` / `presentation_error` / `partially_accepted` / `system_error`

### 索引建议
- index: `user_id`
- index: `problem_id`
- index: `contest_id`
- index: `(user_id, submitted_at)`
- index: `(contest_id, submitted_at)`
- index: `result`

### 接口对齐说明
DTO 层建议提供以下基础语义：

```json
{
  "id": 1,
  "user_id": "uuid",
  "problem_id": 1,
  "result": "Accepted",
  "language": "cpp",
  "time": "..."
}
```

说明：数据库层可使用 `submitted_at`，DTO 层时间字段命名由新接口契约统一。

---

## 6. 实体关系图（概念）

```text
users (uuid)
  ├── 1:N classes.owner_user_id
  ├── 1:N class_memberships.user_id
  ├── 1:N problems.owner_user_id
  ├── 1:N contests.owner_user_id
  ├── 1:N discussions.user_id
  ├── 1:N comments.user_id
  ├── 1:N contest_participants.user_id
  └── 1:N submissions.user_id

classes (uuid)
  ├── 1:N class_memberships.class_id
  ├── 1:N problems.class_id
  └── 1:N contests.class_id

problems (bigint)
  ├── 1:N test_cases.problem_id
  ├── 1:N contest_problems.problem_id
  └── 1:N submissions.problem_id

contests (bigint)
  ├── 1:N contest_problems.contest_id
  ├── 1:N contest_participants.contest_id
  └── 1:N submissions.contest_id

discussions (uuid)
  └── 1:N comments.discussion_id
```
---

## 7. 前后端协作边界说明

### 7.1 建议优先稳定的核心语义（用于前端重构对齐）
#### 用户
- `id`
- `username`
- `email`
- `avatar`
- `synopsis`
- `score`
- `role`
- `created_at`

#### 题目
- `id`
- `title`
- `content`
- `difficulty`
- `time_limit`
- `memory_limit`
- `input`
- `output`

#### 竞赛
- `id`
- `title`
- `subtitle`
- `announcement`
- `status`
- `encrypted`
- `start_at`
- `end_at`

#### 讨论区
- discussion: `id` / `title` / `content` / `user_id` / `username` / `avatar` / `created_at`
- comment: `id` / `discussion_id` / `content` / `user_id` / `username` / `avatar` / `created_at` / `profanity`

#### 提交
- `id`
- `user_id`
- `problem_id`
- `result`
- `language`
- 时间字段

### 7.2 不承诺兼容的部分
- 旧后端内部表结构
- 旧后端 repository/service 分层方式
- 旧接口路径
- 旧权限判定逻辑
- 旧响应包装格式
- 旧前端历史字段命名与历史状态编码

补充说明：
- 接口目标为新语义清晰与可维护，前端按新契约重构。
- 页面层目标为“风格与核心信息架构大致一致”，不要求实现细节一致。

---

## 8. 第一阶段允许实施的内容

后续 agent 在第一阶段**允许**实施：

### 8.1 数据层
- 数据库 migration
- ORM model / entity struct
- enum/constants
- foreign key / index 定义

### 8.2 应用骨架层
- repository 接口与空实现骨架
- service 接口与空实现骨架
- DTO / VO / request / response struct

补充约束：
- 第一阶段允许定义接口与结构体
- 第一阶段不新增可调用业务路径

### 8.3 可以预留但不实现业务的模块
- auth module 基础目录结构
- class module 基础目录结构
- problem module 基础目录结构
- contest module 基础目录结构
- discussion module 基础目录结构
- submission module 基础目录结构

归档约定：
- 现有工程仅作参考，统一放入新目录 `archive/`
- 第一阶段骨架按本草案重新生成

---

## 9. 第一阶段禁止实施的内容

当前阶段**不要**实现：
- 具体权限校验逻辑
- 竞赛自动结算逻辑
- 判题流程逻辑
- 提交消息队列逻辑
- 复杂缓存逻辑
- 实时榜单逻辑
- 班级审核业务逻辑
- 任何需要反复拍板的业务规则

---

## 10. 当前已知的后续话题（本草案暂不展开）

这些问题后续需要单独讨论，但不影响本阶段搭框架：
- 是否需要单独的 `samples` 与 `hidden_test_cases` 区分
- 竞赛 scoreboard / standings 是否需要单独实体
- 题目标签、分类、来源是否需要独立表
- 管理员初始化细节与教师创建流程的接口设计
- 前端在保持页面风格前提下的全量重构计划（路由、状态管理、API 适配层）

---

## 11. 对下一位实施 agent 的明确要求

请基于本草案完成**完整框架搭建**，但只限实体与结构层：

1. 创建上述所有实体的模型定义
2. 建立关系、主键、外键、索引
3. 创建对应模块目录结构
4. 为各模块创建 request/response DTO 骨架
5. 创建 repository/service 接口骨架
6. 不要编写业务逻辑
7. 不要自行扩展新实体，除非另有审核结论

---

## 12. 本草案状态

- 状态：**待继续审核**
- 用途：**第一阶段模型落地依据**
- 当前版本：**v1**

---

## 13. 追加拍板（2026-04-19）

以下决议用于消除第一阶段实施歧义，优先级高于本文中“建议/推荐”表述。

### 13.1 UUID 生成策略
- UUID 统一由应用侧生成（Go 服务生成），数据库不负责生成 UUID。
- 第一阶段默认使用字符串 UUID（`char(36)` 语义），后续如需压缩为 `binary(16)` 再单独评审迁移。

### 13.2 字符集与大小写敏感策略
- 字符集延续现有工程习惯：MySQL 使用 `utf8mb4`。
- 大小写敏感行为参照现网/原工程默认行为，不在第一阶段额外引入复杂 collation 规则。
- 若后续出现用户名/邮箱大小写冲突争议，再做专项修订。

### 13.3 时间与时区
- 所有时间字段按 UTC 存储与传递。
- 连接层与应用层统一按 UTC 处理，避免混用本地时区。

### 13.4 数据库约束策略（第一阶段）
- 采用“轻约束”原则：
  - 保留主键、必要唯一索引、必要非空约束。
  - 不追求在数据库侧实现复杂业务约束（如复杂状态流转、复杂条件唯一）。
- 外键约束第一阶段可不启用，主要在应用层校验。
- 后续可以补 migration 增加外键/约束，但新增前需先清理历史脏数据。

### 13.5 枚举落地
- 第一阶段优先使用 MySQL `ENUM` 落地已拍板的固定枚举字段。
- 新增或变更枚举值通过 migration 管理，不在运行时动态扩展。

### 13.6 竞赛状态策略（自动任务 + 幂等收敛）
- 第一阶段实现自动状态驱动任务（定时扫描）。
- 每轮任务不依赖“上一次任务是否成功”，而是直接按当前时间与 `start_at/end_at` 计算目标状态（`desired_status`），然后将数据收敛到目标状态。
- `contests.status` 继续持久化保存，但其值由自动任务周期性校正。
- 推荐判定规则：
  - `draft`：未发布草稿（可无 `start_at/end_at`）
  - `scheduled`：已发布且 `now < start_at`
  - `running`：`start_at <= now < end_at`
  - `ended`：`now >= end_at`
- 推荐更新策略：`UPDATE ... WHERE status <> desired_status`，保证幂等；任务重复执行不会产生额外副作用。
- 允许“跨状态修复”（例如异常情况下直接 `scheduled -> ended`），不强制逐级跳转，从而天然具备失败恢复能力。
- 仅在需要触发一次性副作用（如发通知）时，才基于“旧状态 -> 新状态”判断是否触发，并加去重键防重放。

### 13.7 DTO 与前端兼容策略（简化）
- 不做“长期双向兼容映射”作为默认策略。
- 优先采用新模型字段语义，允许并建议前端按新契约进行结构化重构。
- 前端重构以“页面风格与核心信息架构大致一致”为边界，不以旧实现细节为约束。
- 仅在迁移窗口内为少数字段提供临时映射，并在联调后移除。
