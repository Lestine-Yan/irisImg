package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Lestine-Yan/irisImg/backend/config"
	"github.com/Lestine-Yan/irisImg/backend/internal/dao"
	"github.com/Lestine-Yan/irisImg/backend/internal/model"
	apikeypkg "github.com/Lestine-Yan/irisImg/backend/internal/pkg/apikey"
	"github.com/Lestine-Yan/irisImg/backend/internal/pkg/storage"
)

// mockAPIKeyDAO 是 dao.APIKeyDAO 的可控测试替身（内存实现）。
type mockAPIKeyDAO struct {
	byHash  map[string]*model.APIKey
	byID    map[int]*model.APIKey
	created *model.APIKey
	nextID  int
}

func newMockAPIKeyDAO() *mockAPIKeyDAO {
	return &mockAPIKeyDAO{byHash: map[string]*model.APIKey{}, byID: map[int]*model.APIKey{}}
}

func (m *mockAPIKeyDAO) Create(_ context.Context, key *model.APIKey) (*model.APIKey, error) {
	m.nextID++
	key.ID = m.nextID
	key.CreatedAt = time.Unix(0, 0)
	cp := *key
	m.created = &cp
	if m.byHash == nil {
		m.byHash = map[string]*model.APIKey{}
	}
	if m.byID == nil {
		m.byID = map[int]*model.APIKey{}
	}
	m.byHash[cp.KeyHash] = &cp
	m.byID[cp.ID] = &cp
	return &cp, nil
}

func (m *mockAPIKeyDAO) GetByHash(_ context.Context, hash string) (*model.APIKey, error) {
	if k, ok := m.byHash[hash]; ok {
		return k, nil
	}
	return nil, dao.ErrNotFound
}

func (m *mockAPIKeyDAO) GetByID(_ context.Context, id int) (*model.APIKey, error) {
	if k, ok := m.byID[id]; ok {
		return k, nil
	}
	return nil, dao.ErrNotFound
}

func (m *mockAPIKeyDAO) List(_ context.Context) ([]*model.APIKey, error) {
	items := make([]*model.APIKey, 0, len(m.byID))
	for _, k := range m.byID {
		items = append(items, k)
	}
	return items, nil
}

func (m *mockAPIKeyDAO) Revoke(_ context.Context, id int) error {
	if k, ok := m.byID[id]; ok {
		k.Revoked = true
		return nil
	}
	return dao.ErrNotFound
}

func (m *mockAPIKeyDAO) UpdateName(_ context.Context, id int, name string) (*model.APIKey, error) {
	if k, ok := m.byID[id]; ok {
		k.Name = name
		cp := *k
		return &cp, nil
	}
	return nil, dao.ErrNotFound
}

// ResetKey 写入新的哈希 / 前缀并清除吊销状态，同步刷新 byHash 索引。
func (m *mockAPIKeyDAO) ResetKey(_ context.Context, id int, keyHash, prefix string) (*model.APIKey, error) {
	if k, ok := m.byID[id]; ok {
		delete(m.byHash, k.KeyHash)
		k.KeyHash = keyHash
		k.Prefix = prefix
		k.Revoked = false
		m.byHash[keyHash] = k
		cp := *k
		return &cp, nil
	}
	return nil, dao.ErrNotFound
}

func (m *mockAPIKeyDAO) Delete(_ context.Context, id int) error {
	if k, ok := m.byID[id]; ok {
		delete(m.byID, id)
		delete(m.byHash, k.KeyHash)
		return nil
	}
	return dao.ErrNotFound
}

func (m *mockAPIKeyDAO) TouchLastUsed(_ context.Context, _ int, _ time.Time) error { return nil }

func TestAPIKeyService_Create(t *testing.T) {
	m := newMockAPIKeyDAO()
	svc := NewAPIKeyService(m, nil, nil)

	resp, err := svc.Create(context.Background(), &model.CreateAPIKeyRequest{
		Name:  "ci",
		Scope: model.ScopeReadWrite,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if !apikeypkg.IsValidFormat(resp.Key) {
		t.Fatalf("returned plaintext should be valid format")
	}
	// 落库的应是哈希而非明文。
	if m.created.KeyHash == resp.Key {
		t.Fatalf("stored value must be the hash, not the plaintext")
	}
	if m.created.KeyHash != apikeypkg.Hash(resp.Key) {
		t.Fatalf("stored hash mismatch")
	}
}

func TestAPIKeyService_CreateInvalidScope(t *testing.T) {
	svc := NewAPIKeyService(newMockAPIKeyDAO(), nil, nil)
	_, err := svc.Create(context.Background(), &model.CreateAPIKeyRequest{Name: "x", Scope: "bogus"})
	if !errors.Is(err, ErrInvalidScope) {
		t.Fatalf("expected ErrInvalidScope, got %v", err)
	}
}

func TestAPIKeyService_Authenticate(t *testing.T) {
	// 构造一把合法明文及其在库中的实体。
	plaintext, hash, prefix, err := apikeypkg.Generate()
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	valid := &model.APIKey{ID: 1, Name: "ok", Prefix: prefix, Scope: model.ScopeReadOnly, KeyHash: hash}
	revokedPlain, revokedHash, _, _ := apikeypkg.Generate()
	revoked := &model.APIKey{ID: 2, Name: "revoked", Scope: model.ScopeReadOnly, KeyHash: revokedHash, Revoked: true}

	m := newMockAPIKeyDAO()
	m.byHash[hash] = valid
	m.byHash[revokedHash] = revoked
	m.byID[valid.ID] = valid
	m.byID[revoked.ID] = revoked
	svc := NewAPIKeyService(m, nil, nil)
	ctx := context.Background()

	t.Run("ok", func(t *testing.T) {
		got, err := svc.Authenticate(ctx, plaintext)
		if err != nil {
			t.Fatalf("expected success, got %v", err)
		}
		if got.ID != 1 {
			t.Fatalf("unexpected key: %+v", got)
		}
	})

	t.Run("bad format", func(t *testing.T) {
		if _, err := svc.Authenticate(ctx, "short"); !errors.Is(err, ErrInvalidKeyFormat) {
			t.Fatalf("expected ErrInvalidKeyFormat, got %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		// 合法格式但库中不存在。
		absent, _, _, _ := apikeypkg.Generate()
		if _, err := svc.Authenticate(ctx, absent); !errors.Is(err, ErrKeyNotFound) {
			t.Fatalf("expected ErrKeyNotFound, got %v", err)
		}
	})

	t.Run("revoked", func(t *testing.T) {
		if _, err := svc.Authenticate(ctx, revokedPlain); !errors.Is(err, ErrKeyRevoked) {
			t.Fatalf("expected ErrKeyRevoked, got %v", err)
		}
	})
}

// TestAPIKeyService_Rename 覆盖重命名：返回的展示信息应反映新名字；不存在的密钥返回 ErrKeyNotFound。
func TestAPIKeyService_Rename(t *testing.T) {
	svc := NewAPIKeyService(newMockAPIKeyDAO(), nil, nil)
	ctx := context.Background()

	created, err := svc.Create(ctx, &model.CreateAPIKeyRequest{Name: "old", Scope: model.ScopeReadOnly})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	info, err := svc.Rename(ctx, created.ID, "new-name")
	if err != nil {
		t.Fatalf("rename: %v", err)
	}
	if info.Name != "new-name" {
		t.Fatalf("expected name new-name, got %s", info.Name)
	}

	if _, err := svc.Rename(ctx, 9999, "x"); !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound, got %v", err)
	}
}

// TestAPIKeyService_Reset_Unrevokes 覆盖重置明文：
//   - 返回的新明文格式合法、且与旧明文不同；
//   - 重置会取消吊销：先 Revoke 使其无法鉴权，Reset 后新明文可鉴权、旧明文失效。
func TestAPIKeyService_Reset_Unrevokes(t *testing.T) {
	svc := NewAPIKeyService(newMockAPIKeyDAO(), nil, nil)
	ctx := context.Background()

	created, err := svc.Create(ctx, &model.CreateAPIKeyRequest{Name: "k", Scope: model.ScopeReadWrite})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	oldPlaintext := created.Key

	// 吊销后旧明文应无法鉴权。
	if err := svc.Revoke(ctx, created.ID); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if _, err := svc.Authenticate(ctx, oldPlaintext); !errors.Is(err, ErrKeyRevoked) {
		t.Fatalf("expected ErrKeyRevoked after revoke, got %v", err)
	}

	resp, err := svc.Reset(ctx, created.ID)
	if err != nil {
		t.Fatalf("reset: %v", err)
	}
	if !apikeypkg.IsValidFormat(resp.Key) || resp.Key == oldPlaintext {
		t.Fatalf("expected a fresh valid plaintext, got %q", resp.Key)
	}
	if resp.Revoked {
		t.Fatalf("expected revoked=false after reset")
	}

	// 新明文可鉴权，旧明文已失效（哈希被替换）。
	if _, err := svc.Authenticate(ctx, resp.Key); err != nil {
		t.Fatalf("expected new plaintext to authenticate, got %v", err)
	}
	if _, err := svc.Authenticate(ctx, oldPlaintext); !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected old plaintext to be invalid, got %v", err)
	}

	if _, err := svc.Reset(ctx, 9999); !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound, got %v", err)
	}
}

// TestAPIKeyService_Delete_CascadesImages 覆盖删除密钥的级联清理：
//   - 关联图片的物理文件被删除；
//   - 图片记录被删除（返回数量正确）；
//   - 密钥记录被删除（再次删除返回 ErrKeyNotFound）。
func TestAPIKeyService_Delete_CascadesImages(t *testing.T) {
	saver, err := storage.NewSaver(config.StorageConfig{RootDir: filepath.Join(t.TempDir(), "imgs")})
	if err != nil {
		t.Fatalf("new saver: %v", err)
	}
	imgDAO := newMemDAO()

	// 落盘一张图片文件，并登记一条关联 key=1 的图片记录。
	rel, err := saver.Save([]byte("payload"), "img-hash-1", "bin", time.Unix(0, 0))
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	keyID := 1
	imgDAO.byHash["img-hash-1"] = &model.Image{ID: 1, Hash: "img-hash-1", StoredPath: rel, KeyID: &keyID}
	imgDAO.byID[1] = imgDAO.byHash["img-hash-1"]

	// 预放一把密钥（id=1）。
	keyDAO := newMockAPIKeyDAO()
	keyDAO.byID[1] = &model.APIKey{ID: 1, Name: "k", KeyHash: "kh", Prefix: "p", Scope: model.ScopeReadWrite}
	keyDAO.byHash["kh"] = keyDAO.byID[1]

	svc := NewAPIKeyService(keyDAO, imgDAO, saver)
	ctx := context.Background()

	removed, err := svc.Delete(ctx, 1)
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if removed != 1 {
		t.Fatalf("expected 1 image removed, got %d", removed)
	}
	// 物理文件应已删除。
	abs := filepath.Join(saver.RootDir(), filepath.FromSlash(rel))
	if _, statErr := os.Stat(abs); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("expected physical file to be deleted: %s (statErr=%v)", abs, statErr)
	}
	// 图片记录应已删除。
	if _, ok := imgDAO.byID[1]; ok {
		t.Fatalf("expected image record to be deleted")
	}
	// 密钥记录应已删除：再次删除返回 ErrKeyNotFound。
	if _, err := svc.Delete(ctx, 1); !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound on second delete, got %v", err)
	}
}
