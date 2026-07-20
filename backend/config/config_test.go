package config

import (
	"reflect"
	"testing"
)

// TestConfig_Validate 覆盖启动安全校验:release 模式强制拒绝默认/空口令与弱 JWT 密钥,
// debug/test 放过(开发开箱即跑)。闭合「拷贝模板未改口令即上线」的攻击链。
func TestConfig_Validate(t *testing.T) {
	strongSecret := "01234567890123456789012345678901abcdef" // 40 字符,>=32
	defaultSecret := "please-change-me-to-a-long-random-string"

	cases := []struct {
		name             string
		mode, user, pass string
		secret           string
		wantErr          bool
	}{
		{"release 默认口令+默认密钥(应拒)", "release", "admin", "admin123", defaultSecret, true},
		{"release 改口令+强密钥(应过)", "release", "admin", "real-pass-9x", strongSecret, false},
		{"debug 默认口令(应过,开发友好)", "debug", "admin", "admin123", defaultSecret, false},
		{"test 模式放过", "test", "admin", "admin123", defaultSecret, false},
		{"release 短密钥(应拒)", "release", "admin", "real-pass-9x", "short", true},
		{"release 空口令(应拒)", "release", "admin", "", strongSecret, true},
		{"release 空用户名(应拒)", "release", "", "real-pass-9x", strongSecret, true},
		{"release 改口令但密钥仍默认(应拒)", "release", "admin", "real-pass-9x", defaultSecret, true},
		{"release 用户名 admin 合法(密码非默认,应过)", "release", "admin", "real-pass-9x", strongSecret, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cfg := &Config{
				Server: ServerConfig{Mode: c.mode},
				Auth: AuthConfig{
					Username: c.user,
					Password: c.pass,
					JWT:      JWTConfig{Secret: c.secret},
				},
			}
			err := cfg.Validate()
			gotErr := err != nil
			if gotErr != c.wantErr {
				t.Fatalf("Validate() err=%v, wantErr=%v", err, c.wantErr)
			}
		})
	}
}

// TestApplyDefaults 覆盖缺省字段集中兜底：零值 -> 合理默认，显式非零值 -> 保留。
// 重点验证四类边界：
//   - fail-fast 字段（DSN/RootDir）不兜底，保持空（由 entdao.Open/NewSaver fail-fast）。
//   - 安全字段（Username/Password/Secret）不兜底，保持原值（避免绕过 Validate 的 fail-closed）。
//   - AllowedMimeTypes 用 == nil 判断：nil 补默认白名单，显式空列表保留空。
//   - Logger 非法值（非 console 的 encoding、未知 time_format）归默认，与 logger 包兜底一致。
func TestApplyDefaults(t *testing.T) {
	t.Run("全零值补默认", func(t *testing.T) {
		c := &Config{}
		c.ApplyDefaults()

		if c.Server.Host != "0.0.0.0" || c.Server.Port != 8080 || c.Server.Mode != "debug" {
			t.Errorf("server = %+v, want 0.0.0.0/8080/debug", c.Server)
		}
		if c.App.Name != "irisImg" || c.App.Version != "0.1.0" {
			t.Errorf("app = %+v, want irisImg/0.1.0", c.App)
		}
		if c.Database.Driver != "sqlite" {
			t.Errorf("database.driver = %q, want sqlite", c.Database.Driver)
		}
		// DSN fail-fast 不兜底
		if c.Database.DSN != "" {
			t.Errorf("database.dsn = %q, want empty (fail-fast 不兜底)", c.Database.DSN)
		}
		// Auth: 仅 Issuer/ExpireHours 兜底，安全字段保持空
		if c.Auth.JWT.Issuer != "irisImg" || c.Auth.JWT.ExpireHours != 24 {
			t.Errorf("auth.jwt = %+v, want irisImg/24", c.Auth.JWT)
		}
		if c.Auth.Username != "" || c.Auth.Password != "" || c.Auth.JWT.Secret != "" {
			t.Errorf("auth 安全字段被兜底: username=%q password=%q secret=%q (应保持空)",
				c.Auth.Username, c.Auth.Password, c.Auth.JWT.Secret)
		}
		if c.APIKey.RateLimitPerMinute != 100 {
			t.Errorf("apikey.rate_limit_per_minute = %d, want 100", c.APIKey.RateLimitPerMinute)
		}
		// RootDir fail-fast 不兜底
		if c.Storage.RootDir != "" {
			t.Errorf("storage.root_dir = %q, want empty (fail-fast 不兜底)", c.Storage.RootDir)
		}
		if c.Storage.MaxUploadSizeMB != 20 {
			t.Errorf("storage.max_upload_size_mb = %d, want 20", c.Storage.MaxUploadSizeMB)
		}
		if !reflect.DeepEqual(c.Storage.AllowedMimeTypes, defaultAllowedMimeTypes) {
			t.Errorf("storage.allowed_mime_types = %v, want %v (nil 补默认白名单)",
				c.Storage.AllowedMimeTypes, defaultAllowedMimeTypes)
		}
		if c.Logger.Level != "info" || c.Logger.Encoding != "json" ||
			c.Logger.Output != "stdout" || c.Logger.TimeFormat != "iso8601" {
			t.Errorf("logger = %+v, want info/json/stdout/iso8601", c.Logger)
		}
	})

	t.Run("显式非零值保留", func(t *testing.T) {
		mimes := []string{"image/png", "image/bmp"}
		c := &Config{
			Server:   ServerConfig{Host: "127.0.0.1", Port: 9000, Mode: "release"},
			App:      AppConfig{Name: "myImg", Version: "2.0.0"},
			Database: DatabaseConfig{Driver: "sqlite", DSN: "/data/x.db", AutoMigrate: false},
			Auth: AuthConfig{
				Username: "alice",
				Password: "strong-pass",
				JWT:      JWTConfig{Secret: "abcdef", Issuer: "myImg", ExpireHours: 12},
			},
			APIKey: APIKeyConfig{RateLimitPerMinute: 30, HTTPSOnly: true},
			Storage: StorageConfig{
				RootDir:          "/data/imgs",
				PublicBaseURL:    "https://img.example.com",
				MaxUploadSizeMB:  5,
				AllowedMimeTypes: mimes,
			},
			Logger: LoggerConfig{Level: "debug", Encoding: "console", Output: "stderr", TimeFormat: "epoch"},
		}
		c.ApplyDefaults()

		if c.Server.Host != "127.0.0.1" || c.Server.Port != 9000 || c.Server.Mode != "release" {
			t.Errorf("server 被覆盖: %+v", c.Server)
		}
		if c.App.Name != "myImg" || c.App.Version != "2.0.0" {
			t.Errorf("app 被覆盖: %+v", c.App)
		}
		if c.Database.DSN != "/data/x.db" || c.Database.AutoMigrate != false {
			t.Errorf("database 被覆盖: %+v", c.Database)
		}
		if c.Auth.Username != "alice" || c.Auth.Password != "strong-pass" || c.Auth.JWT.Secret != "abcdef" {
			t.Errorf("auth 安全字段被覆盖: %+v", c.Auth)
		}
		if c.Auth.JWT.Issuer != "myImg" || c.Auth.JWT.ExpireHours != 12 {
			t.Errorf("auth.jwt 被覆盖: %+v", c.Auth.JWT)
		}
		if c.APIKey.RateLimitPerMinute != 30 || c.APIKey.HTTPSOnly != true {
			t.Errorf("apikey 被覆盖: %+v", c.APIKey)
		}
		if c.Storage.RootDir != "/data/imgs" || c.Storage.PublicBaseURL != "https://img.example.com" ||
			c.Storage.MaxUploadSizeMB != 5 || !reflect.DeepEqual(c.Storage.AllowedMimeTypes, mimes) {
			t.Errorf("storage 被覆盖: %+v", c.Storage)
		}
		if c.Logger.Level != "debug" || c.Logger.Encoding != "console" ||
			c.Logger.Output != "stderr" || c.Logger.TimeFormat != "epoch" {
			t.Errorf("logger 被覆盖: %+v", c.Logger)
		}
	})

	t.Run("AllowedMimeTypes nil 补默认 显式空列表保留", func(t *testing.T) {
		// nil -> 默认白名单
		c1 := &Config{}
		c1.ApplyDefaults()
		if c1.Storage.AllowedMimeTypes == nil {
			t.Fatal("nil MIME 兜底后仍为 nil")
		}
		if !reflect.DeepEqual(c1.Storage.AllowedMimeTypes, defaultAllowedMimeTypes) {
			t.Errorf("nil MIME = %v, want %v", c1.Storage.AllowedMimeTypes, defaultAllowedMimeTypes)
		}

		// 显式空列表 -> 保留空（尊重用户「禁止所有上传」意图，不被默认白名单覆盖）
		c2 := &Config{Storage: StorageConfig{AllowedMimeTypes: []string{}}}
		c2.ApplyDefaults()
		if c2.Storage.AllowedMimeTypes == nil {
			t.Fatal("显式空列表被改成 nil")
		}
		if len(c2.Storage.AllowedMimeTypes) != 0 {
			t.Errorf("显式空列表 = %v, want 空切片", c2.Storage.AllowedMimeTypes)
		}
	})

	t.Run("Logger 非法值归默认", func(t *testing.T) {
		// encoding 非 console -> json；time_format 非 epoch/rfc3339 -> iso8601
		c := &Config{Logger: LoggerConfig{Encoding: "xml", TimeFormat: "unknown"}}
		c.ApplyDefaults()
		if c.Logger.Encoding != "json" {
			t.Errorf("encoding = %q, want json (非法值归默认)", c.Logger.Encoding)
		}
		if c.Logger.TimeFormat != "iso8601" {
			t.Errorf("time_format = %q, want iso8601 (非法值归默认)", c.Logger.TimeFormat)
		}
	})
}
