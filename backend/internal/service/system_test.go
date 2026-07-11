package service

import (
	"reflect"
	"testing"

	"github.com/Lestine-Yan/irisImg/backend/config"
)

// TestSystemService_Config 覆盖 config -> DTO 的关键字段映射与兜底逻辑：
// host/port 透传、dsn 剥离 ? 得纯路径、0 值回退生效默认（限速 100 / 上传 20 MiB）、
// nil AllowedMimeTypes 兜底为非 nil 空切片（避免 JSON null）。
func TestSystemService_Config(t *testing.T) {
	cfg := &config.Config{
		Server:   config.ServerConfig{Host: "0.0.0.0", Port: 8080},
		Database: config.DatabaseConfig{Driver: "sqlite", DSN: "data/irisImg.db?_pragma=busy_timeout(5000)"},
		APIKey:   config.APIKeyConfig{RateLimitPerMinute: 0, HTTPSOnly: false},
		Storage: config.StorageConfig{
			RootDir:          "data/imgs",
			PublicBaseURL:    "",
			MaxUploadSizeMB:  0,
			AllowedMimeTypes: nil,
		},
	}

	got := NewSystemService(cfg).Config()

	if got.Server.Host != "0.0.0.0" || got.Server.Port != 8080 {
		t.Errorf("server = %+v, want host 0.0.0.0 port 8080", got.Server)
	}
	if got.Database.Driver != "sqlite" {
		t.Errorf("database.driver = %q, want sqlite", got.Database.Driver)
	}
	if want := "data/irisImg.db"; got.Database.Path != want {
		t.Errorf("database.path = %q, want %q (dsn 剥离 ? 之后)", got.Database.Path, want)
	}
	if got.APIKey.RateLimitPerMinute != 100 {
		t.Errorf("apikey.rate_limit_per_minute = %d, want 100 (0 回退默认)", got.APIKey.RateLimitPerMinute)
	}
	if got.APIKey.HTTPSOnly != false {
		t.Errorf("apikey.https_only = %v, want false", got.APIKey.HTTPSOnly)
	}
	if got.Storage.MaxUploadSizeMB != 20 {
		t.Errorf("storage.max_upload_size_mb = %d, want 20 (0 回退默认)", got.Storage.MaxUploadSizeMB)
	}
	if got.Storage.AllowedMimeTypes == nil {
		t.Errorf("storage.allowed_mime_types = nil, want 非 nil 空切片（避免 JSON null）")
	}
	if len(got.Storage.AllowedMimeTypes) != 0 {
		t.Errorf("storage.allowed_mime_types len = %d, want 0", len(got.Storage.AllowedMimeTypes))
	}
	if got.Storage.RootDir != "data/imgs" {
		t.Errorf("storage.root_dir = %q, want data/imgs", got.Storage.RootDir)
	}
	if got.Storage.PublicBaseURL != "" {
		t.Errorf("storage.public_base_url = %q, want empty", got.Storage.PublicBaseURL)
	}
}

// TestSystemService_ConfigDSNStripping 表驱动覆盖 dsn 剥离 ? 的边界。
func TestSystemService_ConfigDSNStripping(t *testing.T) {
	tests := []struct {
		name string
		dsn  string
		want string
	}{
		{"no question mark", "data/irisImg.db", "data/irisImg.db"},
		{"question in middle", "data/irisImg.db?_pragma=wal", "data/irisImg.db"},
		{"question at start", "?_pragma=wal", ""},
		{"empty dsn", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewSystemService(&config.Config{Database: config.DatabaseConfig{Driver: "sqlite", DSN: tt.dsn}}).Config()
			if got.Database.Path != tt.want {
				t.Errorf("path = %q, want %q", got.Database.Path, tt.want)
			}
		})
	}
}

// TestSystemService_ConfigExplicitValues 显式非零值不被默认覆盖。
func TestSystemService_ConfigExplicitValues(t *testing.T) {
	mimes := []string{"image/png", "image/jpeg"}
	got := NewSystemService(&config.Config{
		APIKey: config.APIKeyConfig{RateLimitPerMinute: 30, HTTPSOnly: true},
		Storage: config.StorageConfig{
			MaxUploadSizeMB:  5,
			AllowedMimeTypes: mimes,
		},
	}).Config()

	if got.APIKey.RateLimitPerMinute != 30 {
		t.Errorf("rate_limit_per_minute = %d, want 30", got.APIKey.RateLimitPerMinute)
	}
	if got.Storage.MaxUploadSizeMB != 5 {
		t.Errorf("max_upload_size_mb = %d, want 5", got.Storage.MaxUploadSizeMB)
	}
	if !got.APIKey.HTTPSOnly {
		t.Errorf("https_only = false, want true")
	}
	if !reflect.DeepEqual(got.Storage.AllowedMimeTypes, mimes) {
		t.Errorf("allowed_mime_types = %v, want %v", got.Storage.AllowedMimeTypes, mimes)
	}
}
