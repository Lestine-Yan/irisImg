package entdao

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/Lestine-Yan/irisImg/backend/config"
	"github.com/Lestine-Yan/irisImg/backend/internal/dao"
	"github.com/Lestine-Yan/irisImg/backend/internal/model"
)

// newTestDAO 在临时目录打开一个真实的 SQLite 数据库并完成迁移。
// modernc.org/sqlite 是纯 Go 驱动，无需 CGO，测试可离线运行。
func newTestDAO(t *testing.T) dao.ImageDAO {
	t.Helper()
	cfg := config.DatabaseConfig{
		Driver:      "sqlite",
		DSN:         filepath.Join(t.TempDir(), "test.db"),
		AutoMigrate: true,
	}
	client, err := Open(cfg)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { client.Close() })
	if err := Migrate(context.Background(), client, cfg); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return NewImageDAO(client)
}

func sampleImage() *model.Image {
	return &model.Image{
		Filename:   "cat.png",
		StoredPath: "2026/06/cat.png",
		URL:        "/i/2026/06/cat.png",
		Size:       2048,
		MimeType:   "image/png",
		Width:      100,
		Height:     80,
		Hash:       "abc123",
	}
}

func TestImageDAO_CreateAndGet(t *testing.T) {
	d := newTestDAO(t)
	ctx := context.Background()

	created, err := d.Create(ctx, sampleImage())
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.ID == 0 {
		t.Fatalf("expected non-zero id")
	}
	if created.CreatedAt.IsZero() {
		t.Fatalf("expected created_at to be filled")
	}

	got, err := d.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if got.Hash != "abc123" || got.Filename != "cat.png" {
		t.Fatalf("unexpected row: %+v", got)
	}

	byHash, err := d.GetByHash(ctx, "abc123")
	if err != nil {
		t.Fatalf("get by hash: %v", err)
	}
	if byHash.ID != created.ID {
		t.Fatalf("get by hash id mismatch: %d != %d", byHash.ID, created.ID)
	}
}

func TestImageDAO_NotFound(t *testing.T) {
	d := newTestDAO(t)
	ctx := context.Background()

	if _, err := d.GetByID(ctx, 9999); !errors.Is(err, dao.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if _, err := d.GetByHash(ctx, "missing"); !errors.Is(err, dao.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := d.Delete(ctx, 9999); !errors.Is(err, dao.ErrNotFound) {
		t.Fatalf("expected ErrNotFound on delete, got %v", err)
	}
}

func TestImageDAO_ListAndDelete(t *testing.T) {
	d := newTestDAO(t)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		img := sampleImage()
		img.Hash = "hash" + string(rune('a'+i))
		img.StoredPath = "p/" + img.Hash
		if _, err := d.Create(ctx, img); err != nil {
			t.Fatalf("create %d: %v", i, err)
		}
	}

	items, total, err := d.List(ctx, 0, 2)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 3 {
		t.Fatalf("expected total 3, got %d", total)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items with limit, got %d", len(items))
	}

	if err := d.Delete(ctx, items[0].ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, _, err := d.List(ctx, 0, 0); err != nil {
		t.Fatalf("list after delete: %v", err)
	}
}
