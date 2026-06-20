// Package entdao 是 dao 接口的 Ent + SQLite 实现。
//
// 使用纯 Go 的 modernc.org/sqlite 驱动（注册名 "sqlite"），
// 因此整个后端可以 `CGO_ENABLED=0 go build` 产出单文件可执行程序。
package entdao

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/Lestine-Yan/irisImg/backend/config"
	"github.com/Lestine-Yan/irisImg/backend/ent"

	// 纯 Go 的 SQLite 驱动，import 时以 "sqlite" 名称注册到 database/sql。
	_ "modernc.org/sqlite"
)

// Open 根据配置打开数据库并返回 Ent 客户端。
//
// 注意：modernc.org/sqlite 注册的驱动名为 "sqlite"，而 Ent 的方言名是 "sqlite3"，
// 因此这里先用 database/sql 以 "sqlite" 打开连接，再通过 entsql.OpenDB 告知 Ent 方言，
// 避免直接 entsql.Open(dialect.SQLite, ...) 找不到驱动。
func Open(cfg config.DatabaseConfig) (*ent.Client, error) {
	if cfg.Driver != "" && cfg.Driver != "sqlite" {
		return nil, fmt.Errorf("entdao: unsupported driver %q (only sqlite is supported)", cfg.Driver)
	}
	if cfg.DSN == "" {
		return nil, fmt.Errorf("entdao: empty database dsn")
	}

	// SQLite 的 dsn 形如 "data/irisImg.db?_pragma=...", 取 "?" 之前的文件路径，
	// 确保其所在目录存在，否则首次创建数据库文件会失败。
	if err := ensureDir(cfg.DSN); err != nil {
		return nil, err
	}

	// Ent 的自动迁移要求开启 foreign_keys，否则 Schema.Create 会报错。
	// modernc 驱动通过 "_pragma=foreign_keys(on)" 开启，这里在缺省时自动补上。
	dsn := ensureForeignKeys(cfg.DSN)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("entdao: open sqlite: %w", err)
	}

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := ent.NewClient(ent.Driver(drv))
	return client, nil
}

// Migrate 在 AutoMigrate 开启时创建 / 升级表结构。
func Migrate(ctx context.Context, client *ent.Client, cfg config.DatabaseConfig) error {
	if !cfg.AutoMigrate {
		return nil
	}
	if err := client.Schema.Create(ctx); err != nil {
		return fmt.Errorf("entdao: auto migrate: %w", err)
	}
	return nil
}

// ensureForeignKeys 确保 dsn 中开启了 foreign_keys 外键约束。
// 这是 Ent 自动迁移的前置要求；若调用方已显式配置则原样返回。
func ensureForeignKeys(dsn string) string {
	if strings.Contains(dsn, "foreign_keys") {
		return dsn
	}
	const fk = "_pragma=foreign_keys(on)"
	if strings.Contains(dsn, "?") {
		return dsn + "&" + fk
	}
	return dsn + "?" + fk
}

// ensureDir 确保 dsn 指向的数据库文件所在目录存在。
func ensureDir(dsn string) error {
	path := dsn
	if i := strings.IndexByte(path, '?'); i >= 0 {
		path = path[:i]
	}
	// 内存数据库等特殊 dsn 无需建目录。
	if path == "" || strings.HasPrefix(path, ":memory:") || strings.HasPrefix(path, "file::memory:") {
		return nil
	}
	dir := filepath.Dir(path)
	if dir == "" || dir == "." {
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("entdao: create data dir %q: %w", dir, err)
	}
	return nil
}
