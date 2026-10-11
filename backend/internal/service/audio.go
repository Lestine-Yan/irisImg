package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"path"
	"strings"
	"time"

	"github.com/Lestine-Yan/irisImg/backend/config"
	"github.com/Lestine-Yan/irisImg/backend/internal/dao"
	"github.com/Lestine-Yan/irisImg/backend/internal/model"
	"github.com/Lestine-Yan/irisImg/backend/internal/pkg/audiometa"
	"github.com/Lestine-Yan/irisImg/backend/internal/pkg/storage"
)

// 音频业务专属的 sentinel 错误，供 api 层用 errors.Is 区分映射 HTTP 状态码。
// 大小/空文件/MIME 三个校验哨兵与图片共用（ErrEmptyFile/ErrFileTooLarge/ErrUnsupportedMime）。
var (
	// ErrAudioNotFound 表示目标音频不存在（改封面等按 ID 操作时）。
	ErrAudioNotFound = errors.New("audio not found")
)

// AudioService 负责音频上传的业务编排：嗅探 → 去重 → 元数据解析 → 落盘 → 落库。
//
// 与 ImageService 平行，差异点：
//   - 嗅探用 audiometa.Sniff（魔数），不用 http.DetectContentType（不覆盖 m4a/flac）；
//   - 多一步元数据解析（标签 + 技术参数 + 内嵌封面），解析失败零值降级、绝不阻断上传；
//   - 内嵌封面抽取后经 SaveFlat 平铺落盘到 covers/，内容寻址（同封面多音频共享，
//     因此删除音频时刻意不清理封面文件——见 storage.Saver.Delete 的注释）。
type AudioService struct {
	dao dao.AudioDAO
	// saver 是音频存储器（storage.NewAudioSaver 构造，URL 前缀 /audios，根目录 audio_root_dir）。
	saver *storage.Saver
	cfg   config.StorageConfig
	// maxBytes 是单次上传字节上限，由 cfg.AudioMaxUploadSizeMB 推导。
	maxBytes int64
	// allowedMime 是白名单的快速查表（小写）。
	allowedMime map[string]struct{}
}

// NewAudioService 通过依赖注入构造 AudioService。
//
// 入参里同时收 dao / saver / cfg：cfg 中的部分字段（白名单、大小上限）
// 在构造期就预处理成快速查表，避免每次上传都重新解析。
func NewAudioService(d dao.AudioDAO, s *storage.Saver, cfg config.StorageConfig) *AudioService {
	max := int64(cfg.AudioMaxUploadSizeMB)
	if max <= 0 {
		max = 100 // 与配置默认值保持一致
	}
	maxBytes := max * 1024 * 1024

	allow := make(map[string]struct{}, len(cfg.AudioAllowedMimeTypes))
	for _, m := range cfg.AudioAllowedMimeTypes {
		m = strings.ToLower(strings.TrimSpace(m))
		if m != "" {
			allow[m] = struct{}{}
		}
	}

	return &AudioService{
		dao:         d,
		saver:       s,
		cfg:         cfg,
		maxBytes:    maxBytes,
		allowedMime: allow,
	}
}

// MaxBytes 返回单次上传字节上限，供 api 层为 http.MaxBytesReader 设置阈值。
func (s *AudioService) MaxBytes() int64 {
	return s.maxBytes
}

// Upload 是上传音频的主流程：
//  1. 大小 & 空文件校验
//  2. 魔数嗅探格式与规范化 MIME，对照白名单；MP4 家族额外做音轨结构校验（拒收视频）
//  3. 算 sha256，命中已存在记录直接秒传返回
//  4. 解析元数据（标签 + 技术参数 + 内嵌封面；失败零值降级，不阻断）
//  5. 写盘 → 生成对外 URL → 内嵌封面抽取落盘 → 落库
func (s *AudioService) Upload(ctx context.Context, in *model.UploadAudioInput) (*model.Audio, error) {
	if in == nil || len(in.Content) == 0 {
		return nil, ErrEmptyFile
	}
	if int64(len(in.Content)) > s.maxBytes {
		return nil, ErrFileTooLarge
	}

	// 真实格式嗅探：只认内容魔数，客户端伪造 Content-Type / 扩展名没有用。
	// 返回的 mime 是规范化值（mp3→audio/mpeg、m4a→audio/mp4、flac→audio/flac）。
	format, mime := audiometa.Sniff(in.Content)
	if format == audiometa.FormatUnknown {
		return nil, ErrUnsupportedMime
	}
	if _, ok := s.allowedMime[mime]; !ok {
		return nil, ErrUnsupportedMime
	}
	// MP4 家族（ftyp 盒）还需结构校验排除视频文件：ffmpeg 默认把 m4a 的 major brand
	// 编成 isom，brand 区分不可靠，只有走到 moov.trak.mdia.hdlr 看音轨 handler 才可靠。
	// 视频文件 / 结构损坏均按不支持的类型拒收。
	if format == audiometa.FormatM4A && !audiometa.HasAudioTrack(in.Content) {
		return nil, ErrUnsupportedMime
	}

	// 算 sha256（作为文件名 + 去重键）。
	sum := sha256.Sum256(in.Content)
	hash := hex.EncodeToString(sum[:])

	// 秒传：同 hash 已存在，直接复用记录（含首次上传时解析的元数据与封面），不重复写盘落库。
	if existing, err := s.dao.GetByHash(ctx, hash); err == nil {
		return existing, nil
	} else if !errors.Is(err, dao.ErrNotFound) {
		return nil, err
	}

	// 解析元数据（永不失败，缺失零值降级）。
	info := audiometa.Parse(in.Content, format, int64(len(in.Content)))

	// 标题回退：标签无标题时用文件名去扩展名（音乐库的常见兜底行为）。
	title := info.Title
	if title == "" {
		title = strings.TrimSuffix(in.Filename, path.Ext(in.Filename))
	}

	// 落盘路径：<audio_root>/<YYYY>/<MM>/<sha256>.<format>，扩展名即格式名。
	now := time.Now()
	relPath, err := s.saver.Save(in.Content, hash, string(format), now)
	if err != nil {
		return nil, err
	}
	url := s.saver.PublicURL(relPath)

	// 内嵌封面抽取落盘（best-effort）：内容寻址平铺到 covers/，同一专辑封面被多个
	// 音频共享时天然去重；失败只影响封面展示，不阻断上传。
	embeddedCoverURL := ""
	if len(info.PictureData) > 0 && info.PictureExt != "" {
		coverSum := sha256.Sum256(info.PictureData)
		coverRel := "covers/" + hex.EncodeToString(coverSum[:]) + "." + info.PictureExt
		if err := s.saver.SaveFlat(coverRel, info.PictureData); err == nil {
			embeddedCoverURL = s.saver.PublicURL(coverRel)
		}
	}

	a := &model.Audio{
		Filename:         in.Filename,
		StoredPath:       relPath,
		URL:              url,
		Size:             int64(len(in.Content)),
		MimeType:         mime,
		Hash:             hash,
		Format:           string(format),
		Title:            title,
		Artist:           info.Artist,
		Album:            info.Album,
		AlbumArtist:      info.AlbumArtist,
		Composer:         info.Composer,
		Genre:            info.Genre,
		Year:             info.Year,
		TrackNumber:      info.Track,
		DiscNumber:       info.Disc,
		DurationSeconds:  info.DurationSeconds,
		BitrateKbps:      info.BitrateKBPS,
		SampleRate:       info.SampleRate,
		Channels:         info.Channels,
		BitDepth:         info.BitDepth,
		EmbeddedCoverURL: embeddedCoverURL,
		KeyID:            in.KeyID,
	}
	return s.dao.Create(ctx, a)
}

// List 查询音频列表，按 AudioListQuery 过滤 / 排序 / 分页。
// Limit<=0 时兜底为 24（与内容中心默认页大小一致），避免无限制拉取。
// 排序方向、key_id 过滤由 dao 层落实，service 只做参数兜底与结果组装。
func (s *AudioService) List(ctx context.Context, q model.AudioListQuery) (*model.AudioListResult, error) {
	if q.Limit <= 0 {
		q.Limit = 24
	}
	items, total, err := s.dao.List(ctx, q)
	if err != nil {
		return nil, err
	}
	return &model.AudioListResult{Items: items, Total: total}, nil
}

// GetByID 按 ID 查询单条音频详情，未找到返回 ErrAudioNotFound。
// 供改封面等按 ID 操作前取记录 / api 层直接返回详情。
func (s *AudioService) GetByID(ctx context.Context, id int) (*model.Audio, error) {
	a, err := s.dao.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, dao.ErrNotFound) {
			return nil, ErrAudioNotFound
		}
		return nil, err
	}
	return a, nil
}

// BatchDelete 批量删除音频：物理文件 best-effort 删除 + 元信息记录批量删除。
// 返回 (实际删除条数, 实际删除的音频 ID 列表)；不存在的 ID 静默跳过（幂等语义）。
//
// 与图片同款顺序：ListByIDs 取现存记录 → 逐条 best-effort 删物理文件 → DeleteByIDs 批量删记录。
// 非事务（低频管理操作，与 ImageService.BatchDelete 同款权衡）。
//
// 刻意不删内嵌封面（covers/ 下内容寻址文件）：同一封面被多个音频共享，删一个不应
// 影响其余音频的封面展示；孤儿封面留待后续可选的清扫脚本统一处理。
func (s *AudioService) BatchDelete(ctx context.Context, ids []int) (int, []int, error) {
	if len(ids) == 0 {
		return 0, nil, nil
	}

	// 只删现存的记录：混入的不存在 ID 直接跳过，不构成错误。
	items, err := s.dao.ListByIDs(ctx, ids)
	if err != nil {
		return 0, nil, err
	}
	if len(items) == 0 {
		return 0, []int{}, nil
	}

	// 先逐条删物理音频文件（best-effort）：失败不阻断（如个别文件权限问题），
	// Saver.Delete 对已不存在的文件返回 nil，天然幂等。
	for _, a := range items {
		_ = s.saver.Delete(a.StoredPath)
	}

	deleted, err := s.dao.DeleteByIDs(ctx, ids)
	if err != nil {
		return 0, nil, err
	}

	deletedIDs := make([]int, 0, len(items))
	for _, a := range items {
		deletedIDs = append(deletedIDs, a.ID)
	}
	return deleted, deletedIDs, nil
}

// UpdateCover 修改用户覆盖的封面 URL：空串表示清除覆盖（恢复展示内嵌封面）。
// URL 长度上限（≤2048）由 api 层 binding 校验，这里只做 trim 与空串语义转换；
// 未找到返回 ErrAudioNotFound。
func (s *AudioService) UpdateCover(ctx context.Context, id int, coverURL string) (*model.Audio, error) {
	coverURL = strings.TrimSpace(coverURL)
	var override *string
	if coverURL != "" {
		override = &coverURL
	}
	a, err := s.dao.UpdateCoverURL(ctx, id, override)
	if err != nil {
		if errors.Is(err, dao.ErrNotFound) {
			return nil, ErrAudioNotFound
		}
		return nil, err
	}
	return a, nil
}
