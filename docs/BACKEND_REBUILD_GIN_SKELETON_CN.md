# Gin 重构骨架说明（backend_rebuild）

## 1. 目录
- services/cmd/app/backend_rebuild/main.go
- services/cmd/app/backend_rebuild/internal/http/router/router.go
- services/cmd/app/backend_rebuild/internal/http/handler/*.go
- services/cmd/app/backend_rebuild/internal/http/middleware/auth.go
- services/cmd/app/backend_rebuild/internal/ports/contracts.go
- services/cmd/app/backend_rebuild/internal/usecase/stub/services.go
- services/cmd/app/backend_rebuild/internal/bootstrap/wire.go
- services/cmd/app/backend_rebuild/internal/config/config.go

## 2. 设计目的
- 保持前端接口路径与分组不变（/api/v1 + authGroup + adminGroup）
- 后端内部使用接口隔离（ports）
- 先由 stub 实现占位，再逐域替换为真实 usecase/repository

## 3. 接口分层
- Handler: 参数解析、HTTP 映射、错误码映射
- Service Interface: 业务行为抽象（AuthService/UsersService/...）
- Usecase Stub: 当前全部返回 not implemented，供后续替换

## 4. 如何开始实现
1. 先替换 AuthService（注册、登录、验证码、校验、改密）
2. 再替换 UsersService / ProblemsService（读取链路）
3. 再替换 Discussions / Competitions / SubmitRecords
4. 最后替换 AdminService

## 5. 运行方式
在 services/cmd 目录：
- go run ./app/backend_rebuild

可选环境变量：
- BACKEND_REBUILD_ADDR（默认 127.0.0.1:8082）

## 6. 兼容约束
- 路径、方法、Header/Query 参数名保持不变
- 响应字段保持与前端契约一致
- 逐个接口替换 stub，避免大范围同时改动
