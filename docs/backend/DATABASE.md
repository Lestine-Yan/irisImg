# 持久化（SQLite + Ent）

本文档说明图床后端的数据持久化方案与端到端链路，是跨多文件的「特性级」说明。

## 技术选型

- **ORM：[Ent](https://entgo.io/)**--schema 即 Go 代码，`go generate` 产出类型安全的客户端，类似 Prisma 的开发体验。
- **驱动：[`modernc.org/sqlite`](https://pkg.go.dev/modernc.org/sqlite)**--纯 Go 实现，**无需 CGO**。因此整个后端可以：

  ```bash
  CGO_ENABLED=0 go build -o irisImg ./cmd/server
  ```

  产出零外部依赖的单文件可执行程序，适合个人服务器直接部署。

> 注意：`go mod tidy` 可能把 `github.com/mattn/go-sqlite3`（CGO 驱动）作为 Ent/atlas 的**间接**测试依赖写入 `go.sum`，但项目代码从不 import 它，`CGO_ENABLED=0` 构建也不会编译它。

## 目录与分层

```
backend/
├── ent/                          Ent 代码生成根目录
│   ├── generate.go               //go:generate 指令入口
│   ├── schema/image.go           图片元信息 schema
│   ├── schema/apikey.go          API 密钥 schema
│   ├── schema/log.go             访问日志 schema（表名 logs）
│   └── *.go                      ← 生成产物（勿手改）
└── internal/
    ├── model/image.go            跨层实体 DTO（不依赖 Ent）
    └── dao/
        ├── dao.go                ImageDAO 接口
        ├── errors.go             ErrNotFound
        └── entdao/               接口的 Ent 实现
            ├── db.go             Open / Migrate / DSN 处理
            ├── image.go          imageDAO
            └── image_test.go     真实 SQLite 往返测试
```

分层原则（与 `docs/backend/README.md` 一致）：**service / api 只依赖 `dao` 接口与 `model`，不直接依赖 Ent**。`entdao` 负责 `ent.Image ↔ model.Image` 的转换与错误归一化，因此将来替换存储后端（如换 PostgreSQL 或对象存储）时业务层无感。

## 端到端链路

```
main.go
 ├─ entdao.Open(cfg.Database)         打开 *ent.Client（sqlite 驱动 + dialect.SQLite）
 ├─ entdao.Migrate(...)               cfg.AutoMigrate 时 Schema.Create 建表
 ├─ entdao.NewImageDAO(client)        构造 dao.ImageDAO
 └─ router.New(cfg, imageDAO)         注入（图片路由待接入）
```

## 关键实现细节

1. **驱动名 vs 方言名**：modernc 注册名是 `sqlite`，Ent 方言名是 `sqlite3`。`Open` 先 `sql.Open("sqlite", dsn)`，再 `entsql.OpenDB(dialect.SQLite, db)`，避免直接 `entsql.Open` 找不到驱动。
2. **foreign_keys pragma**：Ent 自动迁移要求开启外键，`Open` 在 DSN 缺省时自动追加 `_pragma=foreign_keys(on)`。
3. **数据目录**：`Open` 会 `MkdirAll` DSN 文件所在目录；`data/` 已加入 `.gitignore`，数据库文件不入库。

## 代码生成

修改 `ent/schema/*.go` 后需重新生成：

```bash
cd backend && go generate ./ent
```

其中日志 schema 见 [`ent/schema/log.md`](./ent/schema/log.md)，日志链路的异步落库、查询过滤、直方图聚合、清理策略等特性级说明见 [`LOG.md`](./LOG.md)。

## 数据表登记

数据库表由 `ent/schema/*.go` 定义，经 `entdao.Migrate` 自动建表。图片表（`images`）与密钥表（`api_keys`）的字段细节见 [`ent/schema/image.md`](./ent/schema/image.md) 与 [`ent/schema/apikey.md`](./ent/schema/apikey.md)；本节登记日志中心的 `logs` 表。

### logs（AccessLog）

日志中心的统一日志表，表名经 `entsql.Annotation{Table: "logs"}` 固定（Go 侧 schema 类型为 `AccessLog`，规避 ent 预声明的 `Log` 标识符，避免默认的 `access_logs`）。该表既记录每条 HTTP 请求（`event=http.request`），也记录关键业务事件（`image.upload` / `apikey.*` / `auth.*` / `log.clear`）与 panic（`event=panic`），供日志中心查询 / 直方图 / 清理。访问日志由中间件异步批量落库。

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| `timestamp` | time | 不可变，默认 `time.Now` | 日志发生时间，用于排序 / 直方图 / 时间范围过滤 |
| `level` | enum | `debug` / `info` / `warn` / `error`，默认 `info` | 日志级别，按 HTTP 状态推导（≥500 error / ≥400 warn / 否则 info）或业务事件显式指定 |
| `event` | string | 非空（`NotEmpty`） | 事件类型：`http.request` / `image.upload` / `apikey.*` / `auth.*` / `log.clear` / `panic` |
| `method` | string | 可空（Optional） | HTTP 方法，仅访问日志有值 |
| `path` | string | 可空（Optional） | 请求路径，用于关键字模糊匹配 |
| `status` | int | 可空（Optional + Nillable） | HTTP 状态码，非请求类事件为空（NULL） |
| `duration_ms` | int | 可空（Optional + Nillable） | 请求耗时（毫秒），仅访问日志有值 |
| `client_ip` | string | 可空（Optional） | 客户端 IP |
| `request_id` | string | 可空（Optional） | 请求追踪 ID，关联同一请求的访问日志与业务事件 |
| `api_key_id` | int | 可空（Optional + Nillable） | 来源 API 密钥 ID，无关联时为空；不建 Edge |
| `username` | string | 可空（Optional） | 操作者用户名（JWT 登录用户） |
| `message` | string | 默认 `""` | 可读描述 / 错误信息 |
| `created_at` | time | 默认 `time.Now` | 记录落库时间 |

**索引**：`timestamp` / `level` / `event` / `request_id` / `api_key_id`（均为普通索引，分别覆盖时间排序与范围过滤、级别筛选、事件筛选、追踪 ID 串联、来源密钥筛选）。

**关联关系**：无外键。日志为高频写入实体，为避免外键约束带来的插入开销，`api_key_id` 仅以普通字段记录来源密钥 ID，不建立 `edge.To` / `edge.From`。

**调用关系**：schema 仅供 Ent 代码生成使用；运行时由 [`internal/dao/entdao/log.go`](./internal/dao/entdao/log.md) 通过生成的 `ent.Client` 读写。异步落库、查询过滤、直方图聚合、清理策略等特性级说明见 [`LOG.md`](./LOG.md)。

## 相关文档

- 配置项：[`config/config.md`](./config/config.md)
- 连接与迁移：[`internal/dao/entdao/db.md`](./internal/dao/entdao/db.md)
- DAO 接口：[`internal/dao/dao.md`](./internal/dao/dao.md)
- 实体 schema：[`ent/schema/image.md`](./ent/schema/image.md)
- 日志 schema：[`ent/schema/log.md`](./ent/schema/log.md)
- 日志链路：[`LOG.md`](./LOG.md)
