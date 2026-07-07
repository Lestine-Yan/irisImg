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

// TestAPIKeyDAO_UpdateResetDelete 覆盖重命名 / 重置 / 物理删除：
//   - UpdateName 修改 name；
//   - ResetKey 替换 hash+prefix 并清除吊销状态，旧哈希失效、新哈希生效；
//   - Delete 物理删除，再次 GetByID 返回 ErrNotFound；
//   - 三者在密钥不存在时均返回 ErrNotFound。
func TestAPIKeyDAO_UpdateResetDelete(t *testing.T) {
	d := newTestAPIKeyDAO(t)
	ctx := context.Background()

	created, err := d.Create(ctx, sampleAPIKey())
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	renamed, err := d.UpdateName(ctx, created.ID, "renamed")
	if err != nil {
		t.Fatalf("update name: %v", err)
	}
	if renamed.Name != "renamed" {
		t.Fatalf("expected name renamed, got %s", renamed.Name)
	}

	// 先吊销，再重置，验证重置清除吊销并替换哈希/前缀。
	if err := d.Revoke(ctx, created.ID); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	reset, err := d.ResetKey(ctx, created.ID, "new-hash", "newpfx")
	if err != nil {
		t.Fatalf("reset key: %v", err)
	}
	if reset.Revoked {
		t.Fatalf("expected revoked=false after reset")
	}
	if reset.KeyHash != "new-hash" || reset.Prefix != "newpfx" {
		t.Fatalf("expected hash/prefix replaced, got %+v", reset)
	}
	// 旧哈希应查不到，新哈希能查到。
	if _, err := d.GetByHash(ctx, "hash-abc"); !errors.Is(err, dao.ErrNotFound) {
		t.Fatalf("expected old hash gone, got %v", err)
	}
	if got, err := d.GetByHash(ctx, "new-hash"); err != nil || got.ID != created.ID {
		t.Fatalf("expected new hash to resolve, got %v", err)
	}

	if err := d.Delete(ctx, created.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := d.GetByID(ctx, created.ID); !errors.Is(err, dao.ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}

	// 不存在的密钥：UpdateName / ResetKey / Delete 均返回 ErrNotFound。
	if _, err := d.UpdateName(ctx, 9999, "x"); !errors.Is(err, dao.ErrNotFound) {
		t.Fatalf("expected ErrNotFound on update missing, got %v", err)
	}
	if _, err := d.ResetKey(ctx, 9999, "h", "p"); !errors.Is(err, dao.ErrNotFound) {
		t.Fatalf("expected ErrNotFound on reset missing, got %v", err)
	}
	if err := d.Delete(ctx, 9999); !errors.Is(err, dao.ErrNotFound) {
		t.Fatalf("expected ErrNotFound on delete missing, got %v", err)
	}
}
