package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 是整个应用的配置根结构。
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	App      AppConfig      `yaml:"app"`
	Auth     AuthConfig     `yaml:"auth"`
	Database DatabaseConfig `yaml:"database"`
	APIKey   APIKeyConfig   `yaml:"apikey"`
	Storage  StorageConfig  `yaml:"storage"`
	Logger   LoggerConfig   `yaml:"logger"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

type AppConfig struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

// AuthConfig 描述唯一管理员账号以及 JWT 签发参数。
// 个人图床仅服务于本人，因此用户名/密码直接来源于配置文件。
type AuthConfig struct {
	Username string    `yaml:"username"`
	Password string    `yaml:"password"`
	JWT      JWTConfig `yaml:"jwt"`
}

// JWTConfig 是 JWT 签发与校验所需的参数。
type JWTConfig struct {
	Secret      string `yaml:"secret"`
	Issuer      string `yaml:"issuer"`
	ExpireHours int    `yaml:"expire_hours"`
}

// DatabaseConfig 描述持久化所用的数据库。
// 当前仅支持 SQLite，使用纯 Go 的 modernc.org/sqlite 驱动（无需 CGO）。
type DatabaseConfig struct {
	// Driver 预留多后端切换能力，目前固定为 "sqlite"。
	Driver string `yaml:"driver"`
	// DSN 是数据库连接串。对 SQLite 而言即数据库文件路径，
	// 可附带查询参数，例如 "data/irisImg.db?_pragma=busy_timeout(5000)"。
	DSN string `yaml:"dsn"`
	// AutoMigrate 为 true 时启动阶段自动建表 / 升级表结构。
	AutoMigrate bool `yaml:"auto_migrate"`
}

// APIKeyConfig 描述 API 密钥鉴权相关的参数。
// API 密钥用于外部程序「申请图片 / 添加图片」，独立于后台 JWT 登录。
type APIKeyConfig struct {
	// RateLimitPerMinute 是单个密钥默认的限流阈值（次/分钟），默认 100。
	// 密钥自身的 rate_limit 字段为 0 时沿用此全局默认。
	RateLimitPerMinute int `yaml:"rate_limit_per_minute"`
	// HTTPSOnly 为 true 时，密钥创建等敏感接口要求请求经由 HTTPS
	// （后端通过 X-Forwarded-Proto 二次校验 Nginx 反代）。本地开发可置 false。
	HTTPSOnly bool `yaml:"https_only"`
}

// StorageConfig 描述图片落盘存储相关的参数。
//
// 物理目录由 RootDir 决定（相对路径相对于后端进程工作目录，部署时建议改为绝对路径），
// 真实文件按 <RootDir>/<YYYY>/<MM>/<sha256>.<ext> 的形式排布。
// 对外访问 URL 由 PublicBaseURL + 相对路径拼接：
//   - PublicBaseURL 为空 → 返回 "/imgs/<rel>"，前端/Nginx 同域代理；
//   - 非空（如 "https://img.example.com"，结尾不带斜杠）→ 拼成绝对地址。
type StorageConfig struct {
	// RootDir 是图片落盘根目录，默认 "data/imgs"。
	RootDir string `yaml:"root_dir"`
	// PublicBaseURL 是对外访问 URL 前缀，空表示走相对路径 "/imgs/..."。
	PublicBaseURL string `yaml:"public_base_url"`
	// MaxUploadSizeMB 限制单次上传字节数（MiB），0 表示走默认 20。
	MaxUploadSizeMB int `yaml:"max_upload_size_mb"`
	// AllowedMimeTypes 是真实 MIME 白名单。
	// 后端用 http.DetectContentType 嗅探内容头部得到真实类型，
	// 不信任客户端提交的 Content-Type，未命中白名单直接拒收。
	AllowedMimeTypes []string `yaml:"allowed_mime_types"`
}

// Global 是加载后的全局配置，便于其他包直接引用。
var Global *Config

// LoggerConfig 描述结构化日志（zap）的参数。
//
// 访问日志经中间件异步落库供日志中心查询；此处控制的是 zap 输出到 stdout/文件的部分，
// 用于运维采集。所有字段都有合理默认值，缺省时按 info/json/stdout/iso8601 处理。
type LoggerConfig struct {
	// Level 是日志级别：debug|info|warn|error，默认 info。
	Level string `yaml:"level"`
	// Encoding 是输出编码：json|console，默认 json。
	Encoding string `yaml:"encoding"`
	// Output 是输出目标：stdout|stderr|<文件路径>，默认 stdout。
	Output string `yaml:"output"`
	// TimeFormat 是时间字段格式：iso8601|rfc3339|epoch，默认 iso8601。
	TimeFormat string `yaml:"time_format"`
}

// Load 从指定路径读取并解析 yaml 配置。
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// 补齐缺省字段的合理默认值（单一事实源），再对外暴露。
	// 不碰 fail-fast 字段（DSN/RootDir）与安全字段（口令/密钥），详见 ApplyDefaults。
	cfg.ApplyDefaults()

	Global = cfg
	return cfg, nil
}

// 默认/已知不安全凭证集合,release 模式下命中即拒绝启动(fail-closed)。
// config.yaml.example 的默认值即取自此处,故「拷贝模板未改即上线」会被直接拦下。
var insecurePasswords = map[string]bool{"admin123": true, "": true}
var insecureSecrets = map[string]bool{"please-change-me-to-a-long-random-string": true, "": true}

// Validate 在生产模式(release)下强制校验安全相关配置,拒绝以默认/空口令或弱 JWT 密钥启动,
// 闭合「拷贝模板未改口令即上线」的攻击链。debug/test 模式放过,保持开发开箱即跑。
// 用户名 admin 本身合法(只要密码非默认),故仅校验密码与密钥。
func (c *Config) Validate() error {
	if c.Server.Mode != "release" {
		return nil
	}
	if c.Auth.Username == "" {
		return fmt.Errorf("生产模式(release)要求 auth.username 非空")
	}
	if insecurePasswords[c.Auth.Password] {
		return fmt.Errorf("生产模式(release)拒绝默认/空密码:请修改 auth.password(勿用 admin123)")
	}
	if insecureSecrets[c.Auth.JWT.Secret] {
		return fmt.Errorf("生产模式(release)拒绝默认/空 JWT 密钥:请将 auth.jwt.secret 改为 32 位以上随机串")
	}
	if len(c.Auth.JWT.Secret) < 32 {
		return fmt.Errorf("auth.jwt.secret 长度 %d < 32,不满足生产安全要求", len(c.Auth.JWT.Secret))
	}
	return nil
}

// defaultAllowedMimeTypes 是 storage.allowed_mime_types 缺省时的默认白名单，
// 与 config.yaml 模板保持一致。集中在此作为单一事实源，避免消费方各自兜底导致默认值漂移。
var defaultAllowedMimeTypes = []string{
	"image/png",
	"image/jpeg",
	"image/gif",
	"image/webp",
}

// ApplyDefaults 把缺失（零值）的配置项补齐为合理默认值，建立「缺省默认」的单一事实源。
//
// 设计边界——以下字段刻意不兜底：
//   - fail-fast 字段（database.dsn / storage.root_dir）：空值时由 entdao.Open / NewSaver
//     返回 error、main.go log.Fatalf 拒绝启动。错配会把数据/图片写到错误位置，比启动
//     失败更难诊断，故强制显式配置而非静默填默认。
//   - 安全字段（auth.username / auth.password / auth.jwt.secret）：绝不兜底默认值，否则
//     会绕过 Validate 的 fail-closed（release 模式拒绝默认/空口令与弱密钥启动）。
//   - bool 字段（database.auto_migrate / apikey.https_only）：零值无法区分「缺失」与
//     「显式 false」，兜底会覆盖用户意图。auto_migrate 依赖 config.yaml 模板带 true。
//
// AllowedMimeTypes 用 == nil 判断而非 len==0：区分「缺失/null」（补默认白名单）与
// 「显式空列表」（保留空，尊重用户「禁止所有上传」的意图）。
//
// 由 Load 在 yaml.Unmarshal 后调用；导出是为了供测试与未来从 env/flag 加载配置的路径复用。
// 与 Validate 的关系：Load -> ApplyDefaults ->（main 调）Validate，二者职责互补不重叠。
func (c *Config) ApplyDefaults() {
	// Server
	if c.Server.Host == "" {
		c.Server.Host = "0.0.0.0"
	}
	if c.Server.Port <= 0 {
		c.Server.Port = 8080
	}
	if c.Server.Mode == "" {
		c.Server.Mode = "debug"
	}

	// App
	if c.App.Name == "" {
		c.App.Name = "irisImg"
	}
	if c.App.Version == "" {
		c.App.Version = "0.1.0"
	}

	// Database（DSN / AutoMigrate 不兜底，见方法注释）
	if c.Database.Driver == "" {
		c.Database.Driver = "sqlite"
	}

	// Auth（Username / Password / Secret 不兜底，见方法注释）
	if c.Auth.JWT.Issuer == "" {
		c.Auth.JWT.Issuer = "irisImg"
	}
	if c.Auth.JWT.ExpireHours <= 0 {
		c.Auth.JWT.ExpireHours = 24
	}

	// APIKey（HTTPSOnly 是 bool，零值 false 即合法默认，不兜底）
	if c.APIKey.RateLimitPerMinute <= 0 {
		c.APIKey.RateLimitPerMinute = 100
	}

	// Storage（RootDir fail-fast 不兜底；PublicBaseURL 空串是合法语义不兜底）
	if c.Storage.MaxUploadSizeMB <= 0 {
		c.Storage.MaxUploadSizeMB = 20
	}
	if c.Storage.AllowedMimeTypes == nil {
		c.Storage.AllowedMimeTypes = defaultAllowedMimeTypes
	}

	// Logger：与 internal/pkg/logger 的 parseLevel / openWriteSyncer / timeEncoder
	// 兜底逻辑保持一致，此处集中后那些构造期兜底降级为防御性二次校验。
	if c.Logger.Level == "" {
		c.Logger.Level = "info"
	}
	if c.Logger.Encoding != "console" {
		c.Logger.Encoding = "json"
	}
	if c.Logger.Output == "" {
		c.Logger.Output = "stdout"
	}
	if c.Logger.TimeFormat != "epoch" && c.Logger.TimeFormat != "rfc3339" {
		c.Logger.TimeFormat = "iso8601"
	}
}
