package middleware

import (
	"net/http"

	"github.com/Lestine-Yan/irisImg/backend/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

// HTTPSOnly 在 enabled 为 true 时强制请求经由 HTTPS。
//
// 部署形态为 Nginx 统一做 HTTPS 反向代理、后端本地走 HTTP，
// 因此这里通过反代写入的 X-Forwarded-Proto 头做二次校验；
// 同时兼容后端直接监听 TLS 的场景（c.Request.TLS != nil）。
// 本地开发将配置项 apikey.https_only 置为 false 即可关闭。
func HTTPSOnly(enabled bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !enabled {
			c.Next()
			return
		}

		isHTTPS := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
		if !isHTTPS {
			response.Fail(c, http.StatusForbidden, response.CodeForbidden, "该接口要求使用 HTTPS 访问")
			c.Abort()
			return
		}

		c.Next()
	}
}
