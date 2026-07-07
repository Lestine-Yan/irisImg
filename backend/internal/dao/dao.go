// Package dao 定义持久化访问的抽象接口。
//
// 业务层（service）只依赖这里的接口，不感知具体存储后端，
// 因此可以在不改动业务逻辑的前提下替换底层实现（SQLite / 其他数据库 / 内存等）。
// 当前唯一实现位于子包 entdao，基于 Ent + modernc.org/sqlite。
package dao

import (
	"context"
	"time"

	"github.com/Lestine-Yan/irisImg/backend/internal/model"
)

// ImageDAO 抽象图片元信息的持久化操作。
type ImageDAO interface {
	// Create 落库一条图片元信息，成功后回填自增 ID 与创建时间。
	Create(ctx context.Context, img *model.Image) (*model.Image, error)
	// GetByID 按主键查询，未找到返回 ErrNotFound。
	GetByID(ctx context.Context, id int) (*model.Image, error)
	// GetByHash 按内容哈希查询，用于秒传 / 去重；未找到返回 ErrNotFound。
	GetByHash(ctx context.Context, hash string) (*model.Image, error)
	// List 按 ImageListQuery 过滤（可选 key_id）、排序（asc/desc）、分页返回，
	// 同时给出符合过滤条件的总条数（用于前端计算总页数）。
	List(ctx context.Context, q model.ImageListQuery) (items []*model.Image, total int, err error)
	// ListByKeyID 返回指定密钥关联的全部图片（不分页），供删除密钥时级联清理使用。
	ListByKeyID(ctx context.Context, keyID int) ([]*model.Image, error)
	// Delete 按主键删除，未找到返回 ErrNotFound。
	Delete(ctx context.Context, id int) error
	// DeleteByKeyID 批量删除指定密钥关联的全部图片记录，返回实际删除条数。
	DeleteByKeyID(ctx context.Context, keyID int) (int, error)
}

// APIKeyDAO 抽象 API 密钥的持久化操作。
type APIKeyDAO interface {
	// Create 落库一把密钥（存哈希），成功后回填自增 ID 与创建时间。
	Create(ctx context.Context, key *model.APIKey) (*model.APIKey, error)
	// GetByHash 按密钥哈希查询，用于鉴权；未找到返回 ErrNotFound。
	GetByHash(ctx context.Context, hash string) (*model.APIKey, error)
	// GetByID 按主键查询，未找到返回 ErrNotFound。
	GetByID(ctx context.Context, id int) (*model.APIKey, error)
	// List 按创建时间倒序返回全部密钥。
	List(ctx context.Context) ([]*model.APIKey, error)
	// Revoke 将指定密钥标记为已吊销，未找到返回 ErrNotFound。
	Revoke(ctx context.Context, id int) error
	// UpdateName 修改密钥标签，未找到返回 ErrNotFound。
	UpdateName(ctx context.Context, id int, name string) (*model.APIKey, error)
	// ResetKey 重置密钥明文：写入新的哈希与前缀，并清除吊销状态，未找到返回 ErrNotFound。
	ResetKey(ctx context.Context, id int, keyHash, prefix string) (*model.APIKey, error)
	// Delete 物理删除指定密钥，未找到返回 ErrNotFound。
	// 调用方需先清理引用该密钥的图片记录，否则会因外键约束失败。
	Delete(ctx context.Context, id int) error
	// TouchLastUsed 更新指定密钥的最近使用时间。
	TouchLastUsed(ctx context.Context, id int, t time.Time) error
}
