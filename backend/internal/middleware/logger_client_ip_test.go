package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestLogContextFromGin_ClientIPTrustedProxy 验证可信代理边界：
// 仅当 TCP 对端属于 trusted_proxies 时才采信 X-Forwarded-For，
// 且从右往左跳过可信代理、取第一个不可信地址（客户端自带的伪造链被忽略）。
func TestLogContextFromGin_ClientIPTrustedProxy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	if err := r.SetTrustedProxies([]string{"127.0.0.1/8", "::1"}); err != nil {
		t.Fatalf("set trusted proxies: %v", err)
	}
	var got string
	r.GET("/ip", func(c *gin.Context) {
		got = LogContextFromGin(c).ClientIP
		c.Status(http.StatusOK)
	})

	cases := []struct {
		name   string
		remote string
		xff    string
		want   string
	}{
		{
			name:   "可信反代取最右侧真实地址",
			remote: "127.0.0.1:12345",
			xff:    "1.2.3.4, 203.0.113.5",
			want:   "203.0.113.5",
		},
		{
			name:   "不可信对端忽略伪造 XFF",
			remote: "203.0.113.9:12345",
			xff:    "1.2.3.4",
			want:   "203.0.113.9",
		},
		{
			name:   "直连无请求头取 RemoteAddr",
			remote: "198.51.100.4:12345",
			want:   "198.51.100.4",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/ip", nil)
			req.RemoteAddr = c.remote
			if c.xff != "" {
				req.Header.Set("X-Forwarded-For", c.xff)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if got != c.want {
				t.Fatalf("ClientIP = %q, want %q", got, c.want)
			}
		})
	}
}
