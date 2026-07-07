package entdao

import (
	"context"

	"github.com/Lestine-Yan/irisImg/backend/ent"
	"github.com/Lestine-Yan/irisImg/backend/ent/image"
	"github.com/Lestine-Yan/irisImg/backend/internal/dao"
	"github.com/Lestine-Yan/irisImg/backend/internal/model"
)

// imageDAO 是 dao.ImageDAO 的 Ent 实现。
type imageDAO struct {
	client *ent.Client
}

// NewImageDAO 基于 Ent 客户端构造 dao.ImageDAO。
func NewImageDAO(client *ent.Client) dao.ImageDAO {
	return &imageDAO{client: client}
}

// 编译期断言：imageDAO 必须实现 dao.ImageDAO。
var _ dao.ImageDAO = (*imageDAO)(nil)

func (d *imageDAO) Create(ctx context.Context, img *model.Image) (*model.Image, error) {
	row, err := d.client.Image.Create().
		SetFilename(img.Filename).
		SetStoredPath(img.StoredPath).
		SetURL(img.URL).
		SetSize(img.Size).
		SetMimeType(img.MimeType).
		SetWidth(img.Width).
		SetHeight(img.Height).
		SetHash(img.Hash).
		SetNillableKeyID(img.KeyID).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return toModel(row), nil
}

func (d *imageDAO) GetByID(ctx context.Context, id int) (*model.Image, error) {
	row, err := d.client.Image.Get(ctx, id)
	if err != nil {
		return nil, wrapErr(err)
	}
	return toModel(row), nil
}

func (d *imageDAO) GetByHash(ctx context.Context, hash string) (*model.Image, error) {
	row, err := d.client.Image.Query().
		Where(image.HashEQ(hash)).
		First(ctx)
	if err != nil {
		return nil, wrapErr(err)
	}
	return toModel(row), nil
}

func (d *imageDAO) List(ctx context.Context, q model.ImageListQuery) ([]*model.Image, int, error) {
	total, err := d.countImages(ctx, q)
	if err != nil {
		return nil, 0, err
	}

	query := d.client.Image.Query()
	if q.KeyID != nil {
		query = query.Where(image.KeyIDEQ(*q.KeyID))
	}

	// 默认升序；仅当显式指定 "desc" 时才倒序。空字符串按升序处理，契合内容中心需求。
	if q.Order == "desc" {
		query = query.Order(ent.Desc(image.FieldCreatedAt))
	} else {
		query = query.Order(ent.Asc(image.FieldCreatedAt))
	}
	if q.Offset > 0 {
		query = query.Offset(q.Offset)
	}
	if q.Limit > 0 {
		query = query.Limit(q.Limit)
	}
	rows, err := query.All(ctx)
	if err != nil {
		return nil, 0, err
	}

	items := make([]*model.Image, 0, len(rows))
	for _, row := range rows {
		items = append(items, toModel(row))
	}
	return items, total, nil
}

// countImages 统计符合过滤条件的图片总数，过滤条件与 List 保持一致。
func (d *imageDAO) countImages(ctx context.Context, q model.ImageListQuery) (int, error) {
	query := d.client.Image.Query()
	if q.KeyID != nil {
		query = query.Where(image.KeyIDEQ(*q.KeyID))
	}
	return query.Count(ctx)
}

func (d *imageDAO) Delete(ctx context.Context, id int) error {
	if err := d.client.Image.DeleteOneID(id).Exec(ctx); err != nil {
		return wrapErr(err)
	}
	return nil
}

// ListByKeyID 返回指定密钥关联的全部图片（不分页），供删除密钥时级联清理使用。
func (d *imageDAO) ListByKeyID(ctx context.Context, keyID int) ([]*model.Image, error) {
	rows, err := d.client.Image.Query().
		Where(image.KeyIDEQ(keyID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]*model.Image, 0, len(rows))
	for _, row := range rows {
		items = append(items, toModel(row))
	}
	return items, nil
}

// DeleteByKeyID 批量删除指定密钥关联的全部图片记录，返回实际删除条数。
func (d *imageDAO) DeleteByKeyID(ctx context.Context, keyID int) (int, error) {
	n, err := d.client.Image.Delete().
		Where(image.KeyIDEQ(keyID)).
		Exec(ctx)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// wrapErr 将 Ent 的「记录不存在」错误统一转换为 dao.ErrNotFound。
func wrapErr(err error) error {
	if ent.IsNotFound(err) {
		return dao.ErrNotFound
	}
	return err
}

// toModel 将 Ent 实体转换为跨层的 model.Image。
func toModel(e *ent.Image) *model.Image {
	if e == nil {
		return nil
	}
	return &model.Image{
		ID:         e.ID,
		Filename:   e.Filename,
		StoredPath: e.StoredPath,
		URL:        e.URL,
		Size:       e.Size,
		MimeType:   e.MimeType,
		Width:      e.Width,
		Height:     e.Height,
		Hash:       e.Hash,
		CreatedAt:  e.CreatedAt,
		KeyID:      e.KeyID,
	}
}
