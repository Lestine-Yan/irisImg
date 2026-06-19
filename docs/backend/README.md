# Backend 代码文档

本目录按照 `backend/` 的包结构组织，每个 `.md` 文件对应同名 `.go` 源文件，用于解释其职责、核心类型、关键方法与调用关系。修改源码时请同步更新本目录对应文档。

## 目录结构对照

```
backend/                                  docs/backend/
├── cmd/server/main.go                    └── cmd/server.md
├── config/                               └── config/
│   ├── config.go                            └── config.md
│   └── config.yaml                          (字段说明合并到 config.md)
├── internal/
│   ├── api/                              └── internal/api/
│   │   ├── ping.go                          ├── ping.md
│   │   └── auth.go                          └── auth.md
│   ├── dao/                              (当前为空，预留给图片存储)
│   ├── middleware/                       └── internal/middleware/
│   │   ├── auth.go                          ├── auth.md
│   │   ├── cors.go                          ├── cors.md
│   │   └── logger.go                        └── logger.md
│   ├── model/                            └── internal/model/
│   │   └── auth.go                          └── auth.md
│   ├── pkg/                              └── internal/pkg/
│   │   ├── jwt/jwt.go                       ├── jwt.md
│   │   └── response/response.go             └── response.md
│   ├── router/                           └── internal/router/
│   │   └── router.go                        └── router.md
│   └── service/                          └── internal/service/
│       └── auth.go                          └── auth.md
```

## 分层说明

后端遵循经典分层架构 `api → service → dao → model`，但当前业务非常轻量（单用户登录），所以：

- **`api/`**：Gin 控制器，只负责参数解析、调用 service、组装响应
- **`service/`**：业务逻辑，例如登录校验、签发 token
- **`dao/`**：持久化抽象层；当前没有任何持久化需求，目录为空，未来加入图片元信息存储时在此回填
- **`model/`**：实体与 DTO，跨层数据载体
- **`middleware/`**：Gin 中间件（CORS、日志、JWT 鉴权）
- **`pkg/`**：项目内部可复用的小工具包（JWT 管理器、统一响应体）
- **`router/`**：依赖装配与路由注册的唯一入口
- **`config/`**：YAML 配置的结构体定义与加载
- **`cmd/server/`**：可执行程序入口，组合 config 和 router 启动 HTTP 服务

## 入口阅读顺序

如果你是第一次阅读这份代码，建议按下面的顺序看：

1. [`cmd/server.md`](./cmd/server.md) — 启动流程
2. [`config/config.md`](./config/config.md) — 配置结构
3. [`internal/router/router.md`](./internal/router/router.md) — 路由与依赖装配
4. [`internal/api/auth.md`](./internal/api/auth.md) → [`internal/service/auth.md`](./internal/service/auth.md) → [`internal/pkg/jwt.md`](./internal/pkg/jwt.md) — 登录与 JWT 主链路
5. [`internal/middleware/auth.md`](./internal/middleware/auth.md) — token 校验中间件

登录流程的端到端图示与排错说明见 [`AUTH.md`](./AUTH.md)。
