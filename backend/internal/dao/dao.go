// Package dao 定义持久化访问的抽象接口。
//
// 业务层（service）只依赖这里的接口，不感知具体存储后端，
// 因此可以在不改动业务逻辑的前提下替换底层实现（SQLite / 其他数据库 / 内存等）。
// 当前唯一实现位于子包 entdao，基于 Ent + modernc.org/sqlite。
package dao

import (
	"context"

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
	// List 按创建时间倒序分页返回，同时给出总条数。
	List(ctx context.Context, offset, limit int) (items []*model.Image, total int, err error)
	// Delete 按主键删除，未找到返回 ErrNotFound。
	Delete(ctx context.Context, id int) error
}
