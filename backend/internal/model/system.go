package model

// SystemConfigResponse 是系统配置只读接口（GET /system/config）的响应体。
//
// 仅暴露前端展示所需的非敏感字段，刻意排除 auth.password / auth.jwt.secret 等机密。
// database.path 由 service 层从 database.dsn 剥离 ? 之后的查询参数得到，便于展示纯文件路径。
// 该接口为只读快照，配置变更需修改 config 文件并重启，前端不做在线编辑。
type SystemConfigResponse struct {
	Server   ServerConfigView   `json:"server"`
	Database DatabaseConfigView `json:"database"`
	APIKey   APIKeyConfigView   `json:"apikey"`
	Storage  StorageConfigView  `json:"storage"`
}

// ServerConfigView 暴露服务监听信息。
type ServerConfigView struct {
	Host string `json:"host"` // 监听地址，如 "0.0.0.0"
	Port int    `json:"port"` // 监听端口，如 8080
}

// DatabaseConfigView 暴露数据库位置信息。
type DatabaseConfigView struct {
	Driver string `json:"driver"` // 驱动，当前固定 sqlite
	Path   string `json:"path"`   // 数据库文件路径（dsn 剥离 ? 之后）
}

// APIKeyConfigView 暴露 API 密钥相关的全局开关与默认阈值。
type APIKeyConfigView struct {
	RateLimitPerMinute int  `json:"rate_limit_per_minute"` // 全局默认限流阈值（次/分钟），单密钥 rate_limit 为 0 时回退到此值
	HTTPSOnly          bool `json:"https_only"`            // 是否强制密钥敏感接口走 HTTPS，false 时前端应给出安全警告
}

// StorageConfigView 暴露图片存储相关参数。
type StorageConfigView struct {
	RootDir          string   `json:"root_dir"`           // 图片落盘根目录
	PublicBaseURL    string   `json:"public_base_url"`    // 图片对外访问基址，空表示走相对路径 /imgs/
	MaxUploadSizeMB  int      `json:"max_upload_size_mb"`  // 单次上传上限（MiB）
	AllowedMimeTypes []string `json:"allowed_mime_types"` // 允许上传的 MIME 白名单
}
