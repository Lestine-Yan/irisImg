package middleware

import (
	"net"
	"net/http"

	"github.com/Lestine-Yan/irisImg/backend/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

// HTTPSOnly 在 enabled 为 true 时强制请求经由 HTTPS。
//
// 信任边界：X-Forwarded-Proto 是可被客户端任意伪造的请求头，故仅当请求的 TCP 对端
// （RemoteAddr）属于 trustedProxies 网段时才信任该头；否则只认 c.Request.TLS
// （后端直接监听 TLS 的场景）。这样即便后端端口被误暴露公网，攻击者伪造
// X-Forwarded-Proto: https 也无法绕过——必须真 TLS。
//
// trustedProxies 由 config.Server.TrustedProxies 经 ParseCIDRList 在启动期解析，
// 默认本地回环（同机反代）；跨机反代需在配置里追加反代所在 CIDR。空列表时退化为
// 只认 c.Request.TLS（仍安全）。本地开发将配置项 apikey.https_only 置为 false 即可关闭。
func HTTPSOnly(enabled bool, trustedProxies []*net.IPNet) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !enabled {
			c.Next()
			return
		}

		isHTTPS := c.Request.TLS != nil ||
			(isFromTrustedProxy(c.Request, trustedProxies) && c.GetHeader("X-Forwarded-Proto") == "https")
		if !isHTTPS {
			response.Fail(c, http.StatusForbidden, response.CodeForbidden, "该接口要求使用 HTTPS 访问")
			c.Abort()
			return
		}

		c.Next()
	}
}

// isFromTrustedProxy 判断请求的 TCP 对端地址是否落在受信任反代网段内。
// 取 RemoteAddr 经 net.SplitHostPort 去端口，解析为 IP 后对各 CIDR 做 Contains。
// trustedProxies 为空时返回 false（直连场景，HTTPSOnly 只认 c.Request.TLS）。
func isFromTrustedProxy(r *http.Request, trustedProxies []*net.IPNet) bool {
	if len(trustedProxies) == 0 {
		return false
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// RemoteAddr 无端口时（形如 "1.2.3.4"）直接当 host 处理。
		host = r.RemoteAddr
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	for _, n := range trustedProxies {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}
