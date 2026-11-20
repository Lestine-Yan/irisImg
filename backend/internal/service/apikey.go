package service

import (
	"context"
	"errors"
	"sync"
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
// imageDAO/audioDAO 与 saver/audioSaver 仅用于删除密钥时的级联清理
// （删除该密钥关联的图片 / 音频文件与记录），其余密钥操作不依赖它们。
type APIKeyService struct {
	dao        dao.APIKeyDAO
	imageDAO   dao.ImageDAO
	audioDAO   dao.AudioDAO
	saver      *storage.Saver
	audioSaver *storage.Saver

	// 最近使用时间去抖：同一密钥在 touchInterval 内最多写库一次，
	// 避免每个 API 请求都产生一次 SQLite UPDATE（写锁竞争的主要来源）。
	touchMu       sync.Mutex
	lastTouch     map[int]time.Time
	touchInterval time.Duration
}

// NewAPIKeyService 通过依赖注入构造 APIKeyService。
func NewAPIKeyService(d dao.APIKeyDAO, imgDAO dao.ImageDAO, audioDAO dao.AudioDAO, saver, audioSaver *storage.Saver) *APIKeyService {
	return &APIKeyService{
		dao:           d,
		imageDAO:      imgDAO,
		audioDAO:      audioDAO,
		saver:         saver,
		audioSaver:    audioSaver,
		lastTouch:     make(map[int]time.Time),
		touchInterval: time.Minute,
	}
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

// Delete 物理删除指定密钥，并级联删除其关联的图片与音频（物理文件 + 元信息记录）。
// 返回被删除的图片数量与音频数量。密钥不存在返回 ErrKeyNotFound。
//
// 顺序：取关联图片/音频 → best-effort 删物理文件 → 删记录 → 删密钥。
// 先删内容记录是为了解除外键约束，避免删密钥时被 SQLite 拦截。
// 音频的内嵌封面（covers/ 内容寻址共享文件）刻意不在此清理，与批量删除音频一致。
// 本流程非事务（删除是低频管理操作）：若删密钥失败，内容已清理但密钥仍在，
// 此时密钥已无关联内容，管理员重试删除不会再受外键约束。
func (s *APIKeyService) Delete(ctx context.Context, id int) (int, int, error) {
	if _, err := s.dao.GetByID(ctx, id); err != nil {
		if errors.Is(err, dao.ErrNotFound) {
			return 0, 0, ErrKeyNotFound
		}
		return 0, 0, err
	}

	imgs, err := s.imageDAO.ListByKeyID(ctx, id)
	if err != nil {
		return 0, 0, err
	}
	for _, img := range imgs {
		if img.StoredPath != "" {
			_ = s.saver.Delete(img.StoredPath) // best-effort：文件删除失败不阻断主流程
		}
	}

	removedImages, err := s.imageDAO.DeleteByKeyID(ctx, id)
	if err != nil {
		return 0, 0, err
	}

	audios, err := s.audioDAO.ListByKeyID(ctx, id)
	if err != nil {
		return 0, 0, err
	}
	for _, a := range audios {
		if a.StoredPath != "" {
			_ = s.audioSaver.Delete(a.StoredPath) // best-effort：文件删除失败不阻断主流程
		}
	}

	removedAudios, err := s.audioDAO.DeleteByKeyID(ctx, id)
	if err != nil {
		return 0, 0, err
	}

	if err := s.dao.Delete(ctx, id); err != nil {
		if errors.Is(err, dao.ErrNotFound) {
			return removedImages, removedAudios, ErrKeyNotFound
		}
		return removedImages, removedAudios, err
	}
	return removedImages, removedAudios, nil
}

// Touch 更新密钥的最近使用时间为当前时刻，供鉴权通过后尽力调用。
func (s *APIKeyService) Touch(ctx context.Context, id int) error {
	return s.dao.TouchLastUsed(ctx, id, time.Now())
}

// TouchDebounced 是 Touch 的去抖版本：同一密钥在 touchInterval（默认 60s）内
// 最多写库一次，供每个 API 请求的鉴权中间件调用。
//
// 为什么需要：鉴权中间件原先每请求同步 UPDATE last_used_at，SQLite 单写者 +
// WAL 下这会成为高并发上传的写锁瓶颈；而 last_used_at 只是展示用的近似值，
// 分钟级精度足够。窗口内直接返回，不产生任何数据库 IO。
//
// 写库失败时删除去抖记录，让下一次请求可以立即重试（避免错误状态被窗口固化）。
func (s *APIKeyService) TouchDebounced(ctx context.Context, id int) error {
	now := time.Now()

	s.touchMu.Lock()
	if last, ok := s.lastTouch[id]; ok && now.Sub(last) < s.touchInterval {
		s.touchMu.Unlock()
		return nil
	}
	s.lastTouch[id] = now
	s.touchMu.Unlock()

	if err := s.dao.TouchLastUsed(ctx, id, now); err != nil {
		s.touchMu.Lock()
		delete(s.lastTouch, id)
		s.touchMu.Unlock()
		return err
	}
	return nil
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
