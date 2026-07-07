package service

import (
	"context"
	"errors"
	"time"

	"github.com/Lestine-Yan/irisImg/backend/internal/dao"
	"github.com/Lestine-Yan/irisImg/backend/internal/model"
	apikeypkg "github.com/Lestine-Yan/irisImg/backend/internal/pkg/apikey"
	"github.com/Lestine-Yan/irisImg/backend/internal/pkg/storage"
)

// 密钥相关的 sentinel 错误，供上层（中间件 / api）用 errors.Is 区分处理。
var (
	// ErrInvalidKeyFormat 表示密钥字符串格式非法（长度 / 字符集不符）。
	ErrInvalidKeyFormat = errors.New("invalid api key format")
	// ErrKeyNotFound 表示密钥不存在。
	ErrKeyNotFound = errors.New("api key not found")
	// ErrKeyRevoked 表示密钥已被吊销。
	ErrKeyRevoked = errors.New("api key revoked")
	// ErrInvalidScope 表示请求的权限范围非法。
	ErrInvalidScope = errors.New("invalid api key scope")
)

// APIKeyService 处理 API 密钥的签发、管理与鉴权。
//
// imageDAO 与 saver 仅用于删除密钥时的级联清理（删除该密钥关联的图片文件与记录），
// 其余密钥操作不依赖它们。
type APIKeyService struct {
	dao      dao.APIKeyDAO
	imageDAO dao.ImageDAO
	saver    *storage.Saver
}

// NewAPIKeyService 通过依赖注入构造 APIKeyService。
func NewAPIKeyService(d dao.APIKeyDAO, imgDAO dao.ImageDAO, saver *storage.Saver) *APIKeyService {
	return &APIKeyService{dao: d, imageDAO: imgDAO, saver: saver}
}

// Create 生成并落库一把新密钥，返回包含一次性明文的响应。
func (s *APIKeyService) Create(ctx context.Context, req *model.CreateAPIKeyRequest) (*model.CreateAPIKeyResponse, error) {
	if req.Scope != model.ScopeReadOnly && req.Scope != model.ScopeReadWrite {
		return nil, ErrInvalidScope
	}

	plaintext, hash, prefix, err := apikeypkg.Generate()
	if err != nil {
		return nil, err
	}

	rateLimit := req.RateLimit
	if rateLimit < 0 {
		rateLimit = 0
	}

	created, err := s.dao.Create(ctx, &model.APIKey{
		Name:      req.Name,
		KeyHash:   hash,
		Prefix:    prefix,
		Scope:     req.Scope,
		RateLimit: rateLimit,
	})
	if err != nil {
		return nil, err
	}

	return &model.CreateAPIKeyResponse{
		ID:        created.ID,
		Name:      created.Name,
		Prefix:    created.Prefix,
		Scope:     created.Scope,
		Key:       plaintext,
		RateLimit: created.RateLimit,
		CreatedAt: created.CreatedAt,
	}, nil
}

// List 返回全部密钥的展示信息（不含明文与哈希）。
func (s *APIKeyService) List(ctx context.Context) ([]*model.APIKeyInfo, error) {
	keys, err := s.dao.List(ctx)
	if err != nil {
		return nil, err
	}
	infos := make([]*model.APIKeyInfo, 0, len(keys))
	for _, k := range keys {
		infos = append(infos, toAPIKeyInfo(k))
	}
	return infos, nil
}

// Revoke 吊销指定密钥（软删除：密钥仍展示、仍可操作，但无法通过密钥鉴权）。
// 密钥不存在返回 ErrKeyNotFound。
func (s *APIKeyService) Revoke(ctx context.Context, id int) error {
	if err := s.dao.Revoke(ctx, id); err != nil {
		if errors.Is(err, dao.ErrNotFound) {
			return ErrKeyNotFound
		}
		return err
	}
	return nil
}

// Rename 修改密钥标签，返回更新后的展示信息。密钥不存在返回 ErrKeyNotFound。
func (s *APIKeyService) Rename(ctx context.Context, id int, name string) (*model.APIKeyInfo, error) {
	updated, err := s.dao.UpdateName(ctx, id, name)
	if err != nil {
		if errors.Is(err, dao.ErrNotFound) {
			return nil, ErrKeyNotFound
		}
		return nil, err
	}
	return toAPIKeyInfo(updated), nil
}

// Reset 重置密钥明文：生成新的明文 / 哈希 / 前缀并落库，同时取消吊销状态（重新激活）。
// 返回包含一次性新明文的响应，调用方需妥善保存。密钥不存在返回 ErrKeyNotFound。
func (s *APIKeyService) Reset(ctx context.Context, id int) (*model.ResetAPIKeyResponse, error) {
	plaintext, hash, prefix, err := apikeypkg.Generate()
	if err != nil {
		return nil, err
	}
	updated, err := s.dao.ResetKey(ctx, id, hash, prefix)
	if err != nil {
		if errors.Is(err, dao.ErrNotFound) {
			return nil, ErrKeyNotFound
		}
		return nil, err
	}
	return &model.ResetAPIKeyResponse{
		ID:        updated.ID,
		Name:      updated.Name,
		Prefix:    updated.Prefix,
		Key:       plaintext,
		Revoked:   updated.Revoked,
		CreatedAt: updated.CreatedAt,
	}, nil
}

// Delete 物理删除指定密钥，并级联删除其关联的图片（物理文件 + 元信息记录）。
// 返回被删除的图片数量。密钥不存在返回 ErrKeyNotFound。
//
// 顺序：取关联图片 → best-effort 删物理文件 → 删图片记录 → 删密钥。
// 先删图片记录是为了解除外键约束，避免删密钥时被 SQLite 拦截。
// 本流程非事务（删除是低频管理操作）：若删密钥失败，图片已清理但密钥仍在，
// 此时密钥已无关联图片，管理员重试删除不会再受外键约束。
func (s *APIKeyService) Delete(ctx context.Context, id int) (int, error) {
	if _, err := s.dao.GetByID(ctx, id); err != nil {
		if errors.Is(err, dao.ErrNotFound) {
			return 0, ErrKeyNotFound
		}
		return 0, err
	}

	imgs, err := s.imageDAO.ListByKeyID(ctx, id)
	if err != nil {
		return 0, err
	}
	for _, img := range imgs {
		if img.StoredPath != "" {
			_ = s.saver.Delete(img.StoredPath) // best-effort：文件删除失败不阻断主流程
		}
	}

	removed, err := s.imageDAO.DeleteByKeyID(ctx, id)
	if err != nil {
		return 0, err
	}

	if err := s.dao.Delete(ctx, id); err != nil {
		if errors.Is(err, dao.ErrNotFound) {
			return removed, ErrKeyNotFound
		}
		return removed, err
	}
	return removed, nil
}

// Touch 更新密钥的最近使用时间为当前时刻，供鉴权通过后尽力调用。
func (s *APIKeyService) Touch(ctx context.Context, id int) error {
	return s.dao.TouchLastUsed(ctx, id, time.Now())
}

// Authenticate 校验明文密钥：格式校验 -> 查库 -> 吊销判定。
// 返回对应的密钥实体，失败时返回区分的 sentinel 错误。
func (s *APIKeyService) Authenticate(ctx context.Context, plaintext string) (*model.APIKey, error) {
	if !apikeypkg.IsValidFormat(plaintext) {
		return nil, ErrInvalidKeyFormat
	}

	key, err := s.dao.GetByHash(ctx, apikeypkg.Hash(plaintext))
	if err != nil {
		if errors.Is(err, dao.ErrNotFound) {
			return nil, ErrKeyNotFound
		}
		return nil, err
	}

	if key.Revoked {
		return nil, ErrKeyRevoked
	}

	return key, nil
}

// toAPIKeyInfo 将密钥实体转换为对外展示的 APIKeyInfo（不含明文与哈希）。
func toAPIKeyInfo(k *model.APIKey) *model.APIKeyInfo {
	if k == nil {
		return nil
	}
	return &model.APIKeyInfo{
		ID:         k.ID,
		Name:       k.Name,
		Prefix:     k.Prefix,
		Scope:      k.Scope,
		RateLimit:  k.RateLimit,
		Revoked:    k.Revoked,
		LastUsedAt: k.LastUsedAt,
		CreatedAt:  k.CreatedAt,
	}
}
