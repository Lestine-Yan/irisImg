package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestCORS 覆盖 origin 白名单：* 回显通配、确切 origin 命中回显、未命中不写、
// 空列表关闭跨域、OPTIONS 预检 204。
func TestCORS(t *testing.T) {
	cases := []struct {
		name       string
		origins    []string
		method     string
		reqOrigin  string // 请求 Origin 头，"" 表示不设
		preflight  bool   // OPTIONS 是否带 Access-Control-Request-Method（触发预检）
		wantHeader string // 期望 Access-Control-Allow-Origin，"" 表示不写
		wantStatus int
	}{
		{"* 回显通配", []string{"*"}, "GET", "https://evil.com", false, "*", 200},
		{"白名单命中回显", []string{"https://app.example.com"}, "GET", "https://app.example.com", false, "https://app.example.com", 200},
		{"白名单未命中返回403", []string{"https://app.example.com"}, "GET", "https://evil.com", false, "", 403},
		{"空列表关闭跨域不写", []string{}, "GET", "https://evil.com", false, "", 200},
		{"OPTIONS预检204", []string{"*"}, "OPTIONS", "https://evil.com", true, "*", 204},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.Use(CORS(c.origins))
			r.GET("/", func(ctx *gin.Context) { ctx.Status(200) })
			r.OPTIONS("/", func(ctx *gin.Context) { ctx.Status(200) })

			w := httptest.NewRecorder()
			req := httptest.NewRequest(c.method, "/", nil)
			if c.reqOrigin != "" {
				req.Header.Set("Origin", c.reqOrigin)
			}
			if c.preflight {
				req.Header.Set("Access-Control-Request-Method", "GET")
			}
			r.ServeHTTP(w, req)

			got := w.Header().Get("Access-Control-Allow-Origin")
			if got != c.wantHeader {
				t.Errorf("Allow-Origin = %q, want %q", got, c.wantHeader)
			}
			if w.Code != c.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, c.wantStatus)
			}
		})
	}
}
