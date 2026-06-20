# 持久化（SQLite + Ent）

本文档说明图床后端的数据持久化方案与端到端链路，是跨多文件的「特性级」说明。

## 技术选型

- **ORM：[Ent](https://entgo.io/)**——schema 即 Go 代码，`go generate` 产出类型安全的客户端，类似 Prisma 的开发体验。
- **驱动：[`modernc.org/sqlite`](https://pkg.go.dev/modernc.org/sqlite)**——纯 Go 实现，**无需 CGO**。因此整个后端可以：

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
│   ├── schema/image.go           ← 唯一需要手写的 schema
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

## 相关文档

- 配置项：[`config/config.md`](./config/config.md)
- 连接与迁移：[`internal/dao/entdao/db.md`](./internal/dao/entdao/db.md)
- DAO 接口：[`internal/dao/dao.md`](./internal/dao/dao.md)
- 实体 schema：[`ent/schema/image.md`](./ent/schema/image.md)
