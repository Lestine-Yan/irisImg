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

// Global 是加载后的全局配置，便于其他包直接引用。
var Global *Config

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

	Global = cfg
	return cfg, nil
}
