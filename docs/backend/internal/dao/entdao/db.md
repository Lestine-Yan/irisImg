# internal/dao/entdao/db.go

`dao` 接口的 **Ent + SQLite** 实现的连接与迁移入口。使用纯 Go 的 `modernc.org/sqlite` 驱动（注册名 `sqlite`），因此后端可 `CGO_ENABLED=0 go build` 产出单文件可执行程序。

## 关键函数

- `Open(cfg config.DatabaseConfig) (*ent.Client, error)`
  - 校验 driver（仅支持 `sqlite`）与非空 DSN。
  - 确保 DSN 指向的数据库文件所在目录存在（`ensureDir`）。
  - 自动补全 `foreign_keys` 外键开关（`ensureForeignKeys`）——Ent 自动迁移的前置要求。
  - **驱动名差异处理**：modernc 注册名是 `sqlite`，而 Ent 方言名是 `sqlite3`；因此先用 `database/sql` 以 `sqlite` 打开连接，再经 `entsql.OpenDB(dialect.SQLite, db)` 告知 Ent 方言，避免 `entsql.Open` 找不到驱动。
- `Migrate(ctx, *ent.Client, cfg) error`：当 `cfg.AutoMigrate` 为真时执行 `client.Schema.Create(ctx)` 建表 / 升级表结构。

## 内部辅助

- `ensureForeignKeys(dsn) string`：DSN 缺少 `foreign_keys` 时追加 `_pragma=foreign_keys(on)`。
- `ensureDir(dsn) string`：创建 DSN 文件路径所在目录；`:memory:` 等特殊 DSN 跳过。

## 配置

读取 [`config.DatabaseConfig`](../../../config/config.md)（`driver` / `dsn` / `auto_migrate`）。

## 调用关系

由 `cmd/server/main.go` 在启动阶段调用 `Open` + `Migrate`，再用返回的 `*ent.Client` 构造各 DAO（见 [`image.go`](./image.md)）。
