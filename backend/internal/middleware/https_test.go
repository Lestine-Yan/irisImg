package middleware

import (
	"crypto/tls"
	"net"
	"net/http/httptest"
	"testing"

	"github.com/Lestine-Yan/irisImg/backend/config"
	"github.com/gin-gonic/gin"
)

// TestHTTPSOnly 覆盖信任边界：仅当 TCP 对端（RemoteAddr）属于 trustedProxies 时才认
// X-Forwarded-Proto，否则只认 c.Request.TLS。核心断言：不可信 peer 伪造 XFP=https 仍被拒。
func TestHTTPSOnly(t *testing.T) {
	localNets, err := config.ParseCIDRList([]string{"127.0.0.1/8", "::1"})
	if err != nil {
		t.Fatalf("parse local nets: %v", err)
	}
	crossNets, err := config.ParseCIDRList([]string{"10.0.0.0/8"})
	if err != nil {
		t.Fatalf("parse cross nets: %v", err)
	}

	cases := []struct {
		name       string
		enabled    bool
		nets       []*net.IPNet
		remoteAddr string
		xfp        string // X-Forwarded-Proto 头值，"" 表示不设
		tls        bool   // 模拟 c.Request.TLS != nil
		wantPass   bool
	}{
		{"enabled=false 放行", false, localNets, "127.0.0.1:1234", "http", false, true},
		{"可信peer+XFP=https 放行", true, localNets, "127.0.0.1:1234", "https", false, true},
		{"可信peer+无XFP+无TLS 拒", true, localNets, "127.0.0.1:1234", "", false, false},
		{"不可信peer+伪造XFP=https 拒", true, localNets, "8.8.8.8:1234", "https", false, false},
		{"不可信peer+无XFP+无TLS 拒", true, localNets, "8.8.8.8:1234", "", false, false},
		{"不可信peer+TLS 放行(只认TLS)", true, localNets, "8.8.8.8:1234", "http", true, true},
		{"跨机反代CIDR命中 放行", true, crossNets, "10.0.0.5:1234", "https", false, true},
		{"跨机CIDR未命中+伪造XFP 拒", true, crossNets, "8.8.8.8:1234", "https", false, false},
		{"空trustedProxies+伪造XFP 拒(只认TLS)", true, nil, "127.0.0.1:1234", "https", false, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.Use(HTTPSOnly(c.enabled, c.nets))
			passed := false
			r.GET("/", func(ctx *gin.Context) { passed = true; ctx.Status(200) })

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = c.remoteAddr
			if c.xfp != "" {
				req.Header.Set("X-Forwarded-Proto", c.xfp)
			}
			if c.tls {
				req.TLS = &tls.ConnectionState{}
			}
			r.ServeHTTP(w, req)

			if passed != c.wantPass {
				t.Errorf("passed=%v wantPass=%v status=%d", passed, c.wantPass, w.Code)
			}
		})
	}
}

// TestIsFromTrustedProxy 直接覆盖网段匹配：本地回环命中、跨网段不命中、空列表不命中、
// 无端口 RemoteAddr 兜底、非法 IP 不命中。
func TestIsFromTrustedProxy(t *testing.T) {
	localNets, _ := config.ParseCIDRList([]string{"127.0.0.1/8", "::1"})

	cases := []struct {
		name string
		addr string
		nets []*net.IPNet
		want bool
	}{
		{"回环命中", "127.0.0.1:1234", localNets, true},
		{"回环无端口兜底", "127.0.0.1", localNets, true},
		{"IPv6回环命中", "[::1]:1234", localNets, true},
		{"外部IP不命中", "8.8.8.8:1234", localNets, false},
		{"空列表不命中", "127.0.0.1:1234", nil, false},
		{"非法IP不命中", "not-an-ip:1234", localNets, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = c.addr
			if got := isFromTrustedProxy(req, c.nets); got != c.want {
				t.Errorf("isFromTrustedProxy(%q) = %v, want %v", c.addr, got, c.want)
			}
		})
	}
}
