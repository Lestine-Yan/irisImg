package service

import (
	"context"
	"errors"
	"time"

	"github.com/Lestine-Yan/irisImg/backend/internal/dao"
	"github.com/Lestine-Yan/irisImg/backend/internal/model"
	apikeypkg "github.com/Lestine-Yan/irisImg/backend/internal/pkg/apikey"
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
type APIKeyService struct {
	dao dao.APIKeyDAO
}

// NewAPIKeyService 通过依赖注入构造 APIKeyService。
func NewAPIKeyService(d dao.APIKeyDAO) *APIKeyService {
	return &APIKeyService{dao: d}
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
		infos = append(infos, &model.APIKeyInfo{
			ID:         k.ID,
			Name:       k.Name,
			Prefix:     k.Prefix,
			Scope:      k.Scope,
			RateLimit:  k.RateLimit,
			Revoked:    k.Revoked,
			LastUsedAt: k.LastUsedAt,
			CreatedAt:  k.CreatedAt,
		})
	}
	return infos, nil
}

// Revoke 吊销指定密钥。密钥不存在时返回 dao.ErrNotFound。
func (s *APIKeyService) Revoke(ctx context.Context, id int) error {
	return s.dao.Revoke(ctx, id)
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
