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
// 纯字段映射 + dsn 脱敏，不做任何默认值兜底--缺省默认由 config.ApplyDefaults
// 在 Load 阶段统一补齐（RateLimitPerMinute / MaxUploadSizeMB / AllowedMimeTypes 等），
// 此处直接透传补齐后的值，保证前端展示与实际生效阈值一致。
// database.path 由 database.dsn 剥离首个 ? 之后的查询参数得到，
// 使前端展示纯文件路径而非带 pragma 的连接串。
// auth 段（含密码与 jwt secret）刻意不暴露。
func (s *SystemService) Config() model.SystemConfigResponse {
	cfg := s.cfg
	dbPath := cfg.Database.DSN
	if i := strings.IndexByte(dbPath, '?'); i >= 0 {
		dbPath = dbPath[:i]
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
			RateLimitPerMinute: cfg.APIKey.RateLimitPerMinute,
			HTTPSOnly:          cfg.APIKey.HTTPSOnly,
		},
		Storage: model.StorageConfigView{
			RootDir:          cfg.Storage.RootDir,
			PublicBaseURL:    cfg.Storage.PublicBaseURL,
			MaxUploadSizeMB:  cfg.Storage.MaxUploadSizeMB,
			AllowedMimeTypes: cfg.Storage.AllowedMimeTypes,
		},
	}
}
