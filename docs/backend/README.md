# Backend 代码文档

本目录按照 `backend/` 的包结构组织，每个 `.md` 文件对应同名 `.go` 源文件，用于解释其职责、核心类型、关键方法与调用关系。修改源码时请同步更新本目录对应文档。

## 目录结构对照

```
backend/                                  docs/backend/
├── cmd/server/main.go                    └── cmd/server.md
├── config/                               └── config/
│   ├── config.go                            └── config.md
│   └── config.yaml                          (字段说明合并到 config.md)
├── ent/                                  └── ent/
│   ├── schema/image.go                      └── schema/image.md
│   ├── schema/apikey.go                     └── schema/apikey.md
│   ├── schema/log.go                        └── schema/log.md
│   ├── generate.go                          (go:generate 入口，见 DATABASE.md)
│   └── *.go (生成产物)                       (生成代码，不单独建文档)
├── internal/
│   ├── api/                              └── internal/api/
│   │   ├── ping.go                          ├── ping.md
│   │   ├── auth.go                          ├── auth.md
│   │   ├── apikey.go                        ├── apikey.md
│   │   ├── image.go                         ├── image.md
│   │   ├── log.go                           ├── log.md
│   │   └── system.go                        └── system.md
│   ├── dao/                              └── internal/dao/
│   │   ├── dao.go                            ├── dao.md
│   │   ├── errors.go                         ├── errors.md
│   │   └── entdao/                           └── entdao/
│   │       ├── db.go                            ├── db.md
│   │       ├── image.go                         ├── image.md
│   │       ├── apikey.go                        ├── apikey.md
│   │       └── log.go                           └── log.md
│   ├── middleware/                       └── internal/middleware/
│   │   ├── auth.go                          ├── auth.md
│   │   ├── apikey.go                        ├── apikey.md
│   │   ├── https.go                         ├── https.md
│   │   ├── cors.go                          ├── cors.md
│   │   ├── logger.go                        ├── logger.md
│   │   ├── requestid.go                     ├── requestid.md
│   │   └── recovery.go                      └── recovery.md
│   ├── model/                            └── internal/model/
│   │   ├── auth.go                          ├── auth.md
│   │   ├── image.go                         ├── image.md
│   │   ├── apikey.go                        ├── apikey.md
│   │   ├── log.go                           ├── log.md
│   │   └── system.go                        └── system.md
│   ├── pkg/                              └── internal/pkg/
│   │   ├── jwt/jwt.go                       ├── jwt.md
│   │   ├── apikey/apikey.go                 ├── apikey.md
│   │   ├── ratelimit/ratelimit.go           ├── ratelimit.md
│   │   ├── response/response.go             ├── response.md
│   │   ├── storage/storage.go               ├── storage.md
│   │   └── logger/logger.go                 └── logger.md
│   ├── router/                           └── internal/router/
│   │   └── router.go                        └── router.md
│   └── service/                          └── internal/service/
│       ├── auth.go                          ├── auth.md
│       ├── apikey.go                        ├── apikey.md
│       ├── image.go                         ├── image.md
│       ├── log.go                           ├── log.md
│       └── system.go                        └── system.md
```

> 特性级说明（跨多文件）：持久化方案见 [`DATABASE.md`](./DATABASE.md)，登录链路见 [`AUTH.md`](./AUTH.md)，API 密钥鉴权见 [`APIKEY.md`](./APIKEY.md)，图片上传 / 静态反代约定见 [`IMAGE.md`](./IMAGE.md)，操作日志与请求追踪见 [`LOG.md`](./LOG.md)。

## 分层说明

后端遵循经典分层架构 `api → service → dao → model`，但当前业务非常轻量（单用户登录），所以：

- **`api/`**：Gin 控制器，只负责参数解析、调用 service、组装响应
- **`service/`**：业务逻辑，例如登录校验、签发 token
- **`dao/`**：持久化抽象层；`dao.go` 定义 `ImageDAO` 等接口，`entdao/` 是基于 Ent + SQLite 的实现。service 只依赖接口，便于替换存储后端。详见 [`DATABASE.md`](./DATABASE.md)
- **`model/`**：实体与 DTO，跨层数据载体
- **`ent/`**：Ent 代码生成；只手写 `schema/`，其余为 `go generate` 产物
- **`middleware/`**：Gin 中间件（CORS、日志、JWT 鉴权、API 密钥鉴权、HTTPS 强制）
- **`pkg/`**：项目内部可复用的小工具包（JWT 管理器、API 密钥生成/哈希、按密钥限流令牌桶、统一响应体）
- **`router/`**：依赖装配与路由注册的唯一入口
- **`config/`**：YAML 配置的结构体定义与加载
- **`cmd/server/`**：可执行程序入口，组合 config、数据库与 router 启动 HTTP 服务

## 入口阅读顺序

如果你是第一次阅读这份代码，建议按下面的顺序看：

1. [`cmd/server.md`](./cmd/server.md) — 启动流程
2. [`config/config.md`](./config/config.md) — 配置结构
3. [`internal/router/router.md`](./internal/router/router.md) — 路由与依赖装配
4. [`internal/api/auth.md`](./internal/api/auth.md) → [`internal/service/auth.md`](./internal/service/auth.md) → [`internal/pkg/jwt.md`](./internal/pkg/jwt.md) — 登录与 JWT 主链路
5. [`internal/middleware/auth.md`](./internal/middleware/auth.md) — token 校验中间件

登录流程的端到端图示与排错说明见 [`AUTH.md`](./AUTH.md)。

API 密钥鉴权（独立于 JWT 的另一条链路：`api/apikey` → `service/apikey` → `pkg/apikey` → `dao`，校验侧 `middleware/apikey` → `service/apikey`）的端到端说明见 [`APIKEY.md`](./APIKEY.md)。
