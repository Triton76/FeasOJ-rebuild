# Gin 重构骨架说明（backend-rebuild）

## 1. 目录
- services/cmd/app/backend-rebuild/main.go
- services/cmd/app/backend-rebuild/internal/http/router/router.go
- services/cmd/app/backend-rebuild/internal/http/handler/*.go
- services/cmd/app/backend-rebuild/internal/http/middleware/auth.go
- services/cmd/app/backend-rebuild/internal/ports/contracts.go
- services/cmd/app/backend-rebuild/internal/usecase/stub/services.go
- services/cmd/app/backend-rebuild/internal/bootstrap/wire.go
- services/cmd/app/backend-rebuild/internal/config/config.go

## 2. 设计目的
- 接口设计以新模型语义为主，允许路径、分组与字段重构
- 前端允许大改，仅要求页面风格与核心信息架构大致一致
- 后端内部使用接口隔离（ports）
- 先由 stub 实现占位，再逐域替换为真实 usecase/repository

## 2.1 项目归档与重建约定
- 现有工程仅作参考，后续统一归档到新目录 `archive/`
- 新后端按草案重新生成骨架，不继承旧项目目录约束

## 3. 接口分层
- Handler: 参数解析、HTTP 映射、错误码映射
- Service Interface: 业务行为抽象（AuthService/UsersService/...）
- Usecase Stub: 当前全部返回 not implemented，供后续替换

第一阶段约束：
- 允许定义接口与结构体（DTO / ports / service contracts）
- 不新增可调用业务路径（不开放新的可访问路由）

## 4. 如何开始实现
1. 先替换 AuthService（注册、登录、验证码、校验、改密）
2. 再替换 UsersService / ProblemsService（读取链路）
3. 再替换 Discussions / Competitions / SubmitRecords
4. 最后替换 AdminService

## 5. 运行方式
在 services/cmd 目录：
- go run ./app/backend-rebuild

可选环境变量：
- BACKEND_REBUILD_ADDR（默认 127.0.0.1:8082）

## 5.1 Docker + YAML 本地联调（Auth 实测）
在 `services/cmd/app/backend-rebuild` 目录：

1. 拉取并启动 MySQL：
	- `docker compose -f docker-compose.yaml up -d`
2. 使用 YAML 配置（默认读取 `config.yaml`）：
	- 参考 `config-example.yaml`
	- 可通过 `BACKEND_REBUILD_CONFIG` 指定其他 YAML 路径
3. 启动后端：
	- 回到 `services/cmd` 执行 `go run ./app/backend-rebuild`

优先级说明：
- 环境变量优先于 YAML 配置
- 未配置项回退到默认值

## 6. 兼容约束
- 接口与字段以新模型语义为准，不再以旧前端契约为默认约束
- 前端可按新接口重构，但页面风格与核心信息架构需保持大致一致
- 允许在过渡期提供临时兼容映射，兼容层应可回收且有下线计划
- 逐个接口替换 stub，避免大范围同时改动

追加约束（2026-04-19）：
- 第一阶段不以“长期 DTO 双向兼容映射”为默认目标，优先采用新模型字段语义。
- 前端允许大改，旧接口仅按迁移节奏保留必要兼容窗口。
- 后端仅在必要处提供临时兼容映射，避免长期维护历史负担。
- 数据校验优先放在应用层，数据库保持轻约束（主键/唯一/必要非空）。
