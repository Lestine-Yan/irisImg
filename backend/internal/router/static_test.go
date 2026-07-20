package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAllowedImageExtensions(t *testing.T) {
	cases := []struct {
		name  string
		mimes []string
		want  map[string]struct{}
	}{
		{
			name:  "defaults",
			mimes: []string{"image/png", "image/jpeg", "image/gif", "image/webp"},
			want:  map[string]struct{}{"png": {}, "jpg": {}, "jpeg": {}, "gif": {}, "webp": {}},
		},
		{
			name:  "unknown mime ignored",
			mimes: []string{"application/pdf", "image/png"},
			want:  map[string]struct{}{"png": {}},
		},
		{
			name:  "empty list yields empty set",
			mimes: []string{},
			want:  map[string]struct{}{},
		},
		{
			name:  "trim space and lowercase",
			mimes: []string{"  IMAGE/PNG  "},
			want:  map[string]struct{}{"png": {}},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := allowedImageExtensions(c.mimes)
			if len(got) != len(c.want) {
				t.Fatalf("allowedImageExtensions(%v) = %v, want %v", c.mimes, got, c.want)
			}
			for k := range c.want {
				if _, ok := got[k]; !ok {
					t.Fatalf("allowedImageExtensions(%v) missing %q, got %v", c.mimes, k, got)
				}
			}
		})
	}
}

func TestIsAllowedExt(t *testing.T) {
	allowed := map[string]struct{}{"png": {}, "jpg": {}}
	cases := []struct {
		path string
		want bool
	}{
		{"/2026/06/abc.png", true},
		{"/2026/06/abc.jpg", true},
		{"/2026/06/abc.PNG", true}, // 大小写不敏感
		{"/2026/06/abc.jpeg", false},
		{"/2026/06/abc.yaml", false},
		{"/2026/06/abc.db", false},
		{"/config/config.yaml", false},
		{"/2026/06/abc", false},     // 无扩展名
		{"/2026/06/.gitignore", false}, // 以点开头
		{"/", false},                // 根
		{"", false},                 // 空
		{"/2026/06/", false},        // 目录
	}
	for _, c := range cases {
		t.Run(c.path, func(t *testing.T) {
			if got := isAllowedExt(c.path, allowed); got != c.want {
				t.Fatalf("isAllowedExt(%q) = %v, want %v", c.path, got, c.want)
			}
		})
	}
}

// TestServeImages 验证 /imgs 静态服务的扩展名白名单兜底：
// 即便 root 目录里混入了 config.yaml / .db，未认证访客也只能拿到图片扩展名的文件。
func TestServeImages(t *testing.T) {
	gin.SetMode(gin.TestMode)

	root := t.TempDir()
	imgDir := filepath.Join(root, "2026", "06")
	if err := os.MkdirAll(imgDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(imgDir, "abc.png"), []byte("png-bytes"), 0o644); err != nil {
		t.Fatalf("write png: %v", err)
	}
	if err := os.WriteFile(filepath.Join(imgDir, "secret.db"), []byte("db-bytes"), 0o644); err != nil {
		t.Fatalf("write db: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "config.yaml"), []byte("password: leak"), 0o644); err != nil {
		t.Fatalf("write yaml: %v", err)
	}

	exts := allowedImageExtensions([]string{"image/png", "image/jpeg", "image/gif", "image/webp"})
	r := gin.New()
	h := serveImages(root, exts)
	r.GET("/imgs/*filepath", h)
	r.HEAD("/imgs/*filepath", h)

	cases := []struct {
		name   string
		method string
		target string
		want   int
	}{
		{"png served", http.MethodGet, "/imgs/2026/06/abc.png", http.StatusOK},
		{"yaml blocked", http.MethodGet, "/imgs/config.yaml", http.StatusNotFound},
		{"db blocked", http.MethodGet, "/imgs/2026/06/secret.db", http.StatusNotFound},
		{"no ext blocked", http.MethodGet, "/imgs/2026/06/abc", http.StatusNotFound},
		{"dir listing blocked", http.MethodGet, "/imgs/2026/06/", http.StatusNotFound},
		{"root listing blocked", http.MethodGet, "/imgs/", http.StatusNotFound},
		{"head png served", http.MethodHead, "/imgs/2026/06/abc.png", http.StatusOK},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(c.method, c.target, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != c.want {
				t.Fatalf("%s %s -> %d, want %d (body=%q)", c.method, c.target, w.Code, c.want, w.Body.String())
			}
		})
	}
}
