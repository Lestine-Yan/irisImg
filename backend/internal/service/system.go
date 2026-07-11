package service

import (
	"strings"

	"github.com/Lestine-Yan/irisImg/backend/config"
	"github.com/Lestine-Yan/irisImg/backend/internal/model"
)

// SystemService 负责把运行时 config 转成对外的只读系统配置快照。
//
// 仅做 config -> DTO 的字段映射与脱敏（剥离 dsn 查询参数、排除敏感段），
// 不读不写其它存储。配置变更需修改 config 文件并重启，本服务不提供任何修改能力。
type SystemService struct {
	cfg *config.Config
}

// NewSystemService 构造 SystemService，注入全局 config 快照。
func NewSystemService(cfg *config.Config) *SystemService {
	return &SystemService{cfg: cfg}
}

// Config 返回当前系统配置的只读视图。
//
// database.path 由 database.dsn 剥离首个 ? 之后的查询参数得到，
// 使前端展示纯文件路径而非带 pragma 的连接串。
// AllowedMimeTypes 为空时返回空切片而非 nil，避免 JSON 序列化为 null。
// auth 段（含密码与 jwt secret）刻意不暴露。
func (s *SystemService) Config() model.SystemConfigResponse {
	cfg := s.cfg
	dbPath := cfg.Database.DSN
	if i := strings.IndexByte(dbPath, '?'); i >= 0 {
		dbPath = dbPath[:i]
	}
	mimes := cfg.Storage.AllowedMimeTypes
	if mimes == nil {
		mimes = []string{}
	}
	// 0 / 负数表示未配置：回退到与 ImageService / ratelimit 一致的生效默认值，
	// 使前端展示的是实际生效阈值，而非误导性的 0。
	rateLimit := cfg.APIKey.RateLimitPerMinute
	if rateLimit <= 0 {
		rateLimit = 100 // 与 ratelimit.NewStore 的默认一致
	}
	maxUpload := cfg.Storage.MaxUploadSizeMB
	if maxUpload <= 0 {
		maxUpload = 20 // 与 ImageService 的默认一致
	}
	return model.SystemConfigResponse{
		Server: model.ServerConfigView{
			Host: cfg.Server.Host,
			Port: cfg.Server.Port,
		},
		Database: model.DatabaseConfigView{
			Driver: cfg.Database.Driver,
			Path:   dbPath,
		},
		APIKey: model.APIKeyConfigView{
			RateLimitPerMinute: rateLimit,
			HTTPSOnly:          cfg.APIKey.HTTPSOnly,
		},
		Storage: model.StorageConfigView{
			RootDir:          cfg.Storage.RootDir,
			PublicBaseURL:    cfg.Storage.PublicBaseURL,
			MaxUploadSizeMB:  maxUpload,
			AllowedMimeTypes: mimes,
		},
	}
}
