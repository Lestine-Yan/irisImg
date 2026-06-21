package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Lestine-Yan/irisImg/backend/internal/dao"
	"github.com/Lestine-Yan/irisImg/backend/internal/model"
	apikeypkg "github.com/Lestine-Yan/irisImg/backend/internal/pkg/apikey"
)

// mockAPIKeyDAO 是 dao.APIKeyDAO 的可控测试替身。
type mockAPIKeyDAO struct {
	byHash  map[string]*model.APIKey
	created *model.APIKey
}

func (m *mockAPIKeyDAO) Create(_ context.Context, key *model.APIKey) (*model.APIKey, error) {
	key.ID = 1
	key.CreatedAt = time.Unix(0, 0)
	m.created = key
	return key, nil
}

func (m *mockAPIKeyDAO) GetByHash(_ context.Context, hash string) (*model.APIKey, error) {
	if k, ok := m.byHash[hash]; ok {
		return k, nil
	}
	return nil, dao.ErrNotFound
}

func (m *mockAPIKeyDAO) GetByID(_ context.Context, _ int) (*model.APIKey, error) {
	return nil, dao.ErrNotFound
}

func (m *mockAPIKeyDAO) List(_ context.Context) ([]*model.APIKey, error) {
	return nil, nil
}

func (m *mockAPIKeyDAO) Revoke(_ context.Context, _ int) error                     { return nil }
func (m *mockAPIKeyDAO) TouchLastUsed(_ context.Context, _ int, _ time.Time) error { return nil }

func TestAPIKeyService_Create(t *testing.T) {
	m := &mockAPIKeyDAO{byHash: map[string]*model.APIKey{}}
	svc := NewAPIKeyService(m)

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
	svc := NewAPIKeyService(&mockAPIKeyDAO{byHash: map[string]*model.APIKey{}})
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

	m := &mockAPIKeyDAO{byHash: map[string]*model.APIKey{
		hash:        valid,
		revokedHash: revoked,
	}}
	svc := NewAPIKeyService(m)
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
