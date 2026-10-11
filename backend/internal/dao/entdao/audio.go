package entdao

import (
	"context"

	"github.com/Lestine-Yan/irisImg/backend/ent"
	"github.com/Lestine-Yan/irisImg/backend/ent/audio"
	"github.com/Lestine-Yan/irisImg/backend/internal/dao"
	"github.com/Lestine-Yan/irisImg/backend/internal/model"
)

// audioDAO 是 dao.AudioDAO 的 Ent 实现。
type audioDAO struct {
	client *ent.Client
}

// NewAudioDAO 基于 Ent 客户端构造 dao.AudioDAO。
func NewAudioDAO(client *ent.Client) dao.AudioDAO {
	return &audioDAO{client: client}
}

// 编译期断言：audioDAO 必须实现 dao.AudioDAO。
var _ dao.AudioDAO = (*audioDAO)(nil)

func (d *audioDAO) Create(ctx context.Context, a *model.Audio) (*model.Audio, error) {
	row, err := d.client.Audio.Create().
		SetFilename(a.Filename).
		SetStoredPath(a.StoredPath).
		SetURL(a.URL).
		SetSize(a.Size).
		SetMimeType(a.MimeType).
		SetHash(a.Hash).
		SetFormat(a.Format).
		SetTitle(a.Title).
		SetArtist(a.Artist).
		SetAlbum(a.Album).
		SetAlbumArtist(a.AlbumArtist).
		SetComposer(a.Composer).
		SetGenre(a.Genre).
		SetYear(a.Year).
		SetTrackNumber(a.TrackNumber).
		SetDiscNumber(a.DiscNumber).
		SetDurationSeconds(a.DurationSeconds).
		SetBitrateKbps(a.BitrateKbps).
		SetSampleRate(a.SampleRate).
		SetChannels(a.Channels).
		SetBitDepth(a.BitDepth).
		SetEmbeddedCoverURL(a.EmbeddedCoverURL).
		SetNillableKeyID(a.KeyID).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return audioToModel(row), nil
}

func (d *audioDAO) GetByID(ctx context.Context, id int) (*model.Audio, error) {
	row, err := d.client.Audio.Get(ctx, id)
	if err != nil {
		return nil, wrapErr(err)
	}
	return audioToModel(row), nil
}

func (d *audioDAO) GetByHash(ctx context.Context, hash string) (*model.Audio, error) {
	row, err := d.client.Audio.Query().
		Where(audio.HashEQ(hash)).
		First(ctx)
	if err != nil {
		return nil, wrapErr(err)
	}
	return audioToModel(row), nil
}

func (d *audioDAO) List(ctx context.Context, q model.AudioListQuery) ([]*model.Audio, int, error) {
	total, err := d.countAudios(ctx, q)
	if err != nil {
		return nil, 0, err
	}

	query := d.client.Audio.Query()
	if q.KeyID != nil {
		query = query.Where(audio.KeyIDEQ(*q.KeyID))
	}

	// 默认升序；仅当显式指定 "desc" 时才倒序。空字符串按升序处理，与图片列表一致。
	if q.Order == "desc" {
		query = query.Order(ent.Desc(audio.FieldCreatedAt))
	} else {
		query = query.Order(ent.Asc(audio.FieldCreatedAt))
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

	items := make([]*model.Audio, 0, len(rows))
	for _, row := range rows {
		items = append(items, audioToModel(row))
	}
	return items, total, nil
}

// countAudios 统计符合过滤条件的音频总数，过滤条件与 List 保持一致。
func (d *audioDAO) countAudios(ctx context.Context, q model.AudioListQuery) (int, error) {
	query := d.client.Audio.Query()
	if q.KeyID != nil {
		query = query.Where(audio.KeyIDEQ(*q.KeyID))
	}
	return query.Count(ctx)
}

// ListByIDs 按主键集合批量查询现存的音频记录（不存在的 ID 静默跳过），
// 供批量删除前取 StoredPath 做物理文件清理。
// 空切片防御：直接返回空 slice 不发起查询（与 imageDAO 同款防御）。
func (d *audioDAO) ListByIDs(ctx context.Context, ids []int) ([]*model.Audio, error) {
	if len(ids) == 0 {
		return []*model.Audio{}, nil
	}
	rows, err := d.client.Audio.Query().
		Where(audio.IDIn(ids...)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]*model.Audio, 0, len(rows))
	for _, row := range rows {
		items = append(items, audioToModel(row))
	}
	return items, nil
}

// DeleteByIDs 按主键集合批量删除音频记录，返回实际删除条数（不存在的 ID 静默跳过）。
func (d *audioDAO) DeleteByIDs(ctx context.Context, ids []int) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	n, err := d.client.Audio.Delete().
		Where(audio.IDIn(ids...)).
		Exec(ctx)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// ListByKeyID 返回指定密钥关联的全部音频（不分页），供删除密钥时级联清理使用。
func (d *audioDAO) ListByKeyID(ctx context.Context, keyID int) ([]*model.Audio, error) {
	rows, err := d.client.Audio.Query().
		Where(audio.KeyIDEQ(keyID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]*model.Audio, 0, len(rows))
	for _, row := range rows {
		items = append(items, audioToModel(row))
	}
	return items, nil
}

// DeleteByKeyID 批量删除指定密钥关联的全部音频记录，返回实际删除条数。
func (d *audioDAO) DeleteByKeyID(ctx context.Context, keyID int) (int, error) {
	n, err := d.client.Audio.Delete().
		Where(audio.KeyIDEQ(keyID)).
		Exec(ctx)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// UpdateCoverURL 更新用户覆盖的封面 URL：coverURL 为 nil 时清除覆盖（回到内嵌封面）。
// ent 的 SetNillableCoverURL(nil) 是「不操作」而非「清空」，故 nil 分支必须用 ClearCoverURL。
func (d *audioDAO) UpdateCoverURL(ctx context.Context, id int, coverURL *string) (*model.Audio, error) {
	mutation := d.client.Audio.UpdateOneID(id)
	var row *ent.Audio
	var err error
	if coverURL == nil {
		row, err = mutation.ClearCoverURL().Save(ctx)
	} else {
		row, err = mutation.SetCoverURL(*coverURL).Save(ctx)
	}
	if err != nil {
		return nil, wrapErr(err)
	}
	return audioToModel(row), nil
}

// audioToModel 将 Ent 实体转换为跨层的 model.Audio。
func audioToModel(e *ent.Audio) *model.Audio {
	if e == nil {
		return nil
	}
	return &model.Audio{
		ID:               e.ID,
		Filename:         e.Filename,
		StoredPath:       e.StoredPath,
		URL:              e.URL,
		Size:             e.Size,
		MimeType:         e.MimeType,
		Hash:             e.Hash,
		Format:           e.Format,
		Title:            e.Title,
		Artist:           e.Artist,
		Album:            e.Album,
		AlbumArtist:      e.AlbumArtist,
		Composer:         e.Composer,
		Genre:            e.Genre,
		Year:             e.Year,
		TrackNumber:      e.TrackNumber,
		DiscNumber:       e.DiscNumber,
		DurationSeconds:  e.DurationSeconds,
		BitrateKbps:      e.BitrateKbps,
		SampleRate:       e.SampleRate,
		Channels:         e.Channels,
		BitDepth:         e.BitDepth,
		EmbeddedCoverURL: e.EmbeddedCoverURL,
		CoverURL:         e.CoverURL,
		CreatedAt:        e.CreatedAt,
		KeyID:            e.KeyID,
	}
}
