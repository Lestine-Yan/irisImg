package entdao

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/Lestine-Yan/irisImg/backend/config"
	"github.com/Lestine-Yan/irisImg/backend/internal/dao"
	"github.com/Lestine-Yan/irisImg/backend/internal/model"
)

// newTestAPIKeyDAO 在临时目录打开真实 SQLite 并完成迁移，返回 APIKeyDAO。
func newTestAPIKeyDAO(t *testing.T) dao.APIKeyDAO {
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
	return NewAPIKeyDAO(client)
}

func sampleAPIKey() *model.APIKey {
	return &model.APIKey{
		Name:      "ci",
		KeyHash:   "hash-abc",
		Prefix:    "abcdefgh",
		Scope:     model.ScopeReadOnly,
		RateLimit: 0,
	}
}

func TestAPIKeyDAO_CreateAndGet(t *testing.T) {
	d := newTestAPIKeyDAO(t)
	ctx := context.Background()

	created, err := d.Create(ctx, sampleAPIKey())
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.ID == 0 || created.CreatedAt.IsZero() {
		t.Fatalf("expected id and created_at filled: %+v", created)
	}
	if created.Revoked {
		t.Fatalf("new key should not be revoked")
	}

	got, err := d.GetByHash(ctx, "hash-abc")
	if err != nil {
		t.Fatalf("get by hash: %v", err)
	}
	if got.ID != created.ID || got.Scope != model.ScopeReadOnly {
		t.Fatalf("unexpected row: %+v", got)
	}

	byID, err := d.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if byID.Prefix != "abcdefgh" {
		t.Fatalf("prefix mismatch: %s", byID.Prefix)
	}
}

func TestAPIKeyDAO_NotFound(t *testing.T) {
	d := newTestAPIKeyDAO(t)
	ctx := context.Background()

	if _, err := d.GetByHash(ctx, "missing"); !errors.Is(err, dao.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if _, err := d.GetByID(ctx, 9999); !errors.Is(err, dao.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := d.Revoke(ctx, 9999); !errors.Is(err, dao.ErrNotFound) {
		t.Fatalf("expected ErrNotFound on revoke, got %v", err)
	}
}

func TestAPIKeyDAO_RevokeListAndTouch(t *testing.T) {
	d := newTestAPIKeyDAO(t)
	ctx := context.Background()

	created, err := d.Create(ctx, sampleAPIKey())
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// Revoke 后应被标记为已吊销。
	if err := d.Revoke(ctx, created.ID); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	got, err := d.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("get after revoke: %v", err)
	}
	if !got.Revoked {
		t.Fatalf("expected revoked=true")
	}

	// TouchLastUsed 后 last_used_at 应被写入。
	now := time.Now()
	if err := d.TouchLastUsed(ctx, created.ID, now); err != nil {
		t.Fatalf("touch: %v", err)
	}
	got, err = d.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("get after touch: %v", err)
	}
	if got.LastUsedAt == nil {
		t.Fatalf("expected last_used_at to be set")
	}

	list, err := d.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 key, got %d", len(list))
	}
}
