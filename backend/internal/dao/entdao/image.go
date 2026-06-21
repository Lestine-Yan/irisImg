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

func (d *imageDAO) List(ctx context.Context, offset, limit int) ([]*model.Image, int, error) {
	total, err := d.client.Image.Query().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	q := d.client.Image.Query().Order(ent.Desc(image.FieldCreatedAt))
	if offset > 0 {
		q = q.Offset(offset)
	}
	if limit > 0 {
		q = q.Limit(limit)
	}
	rows, err := q.All(ctx)
	if err != nil {
		return nil, 0, err
	}

	items := make([]*model.Image, 0, len(rows))
	for _, row := range rows {
		items = append(items, toModel(row))
	}
	return items, total, nil
}

func (d *imageDAO) Delete(ctx context.Context, id int) error {
	if err := d.client.Image.DeleteOneID(id).Exec(ctx); err != nil {
		return wrapErr(err)
	}
	return nil
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
