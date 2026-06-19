package middleware

import (
	"strings"

	"github.com/Lestine-Yan/irisImg/backend/internal/pkg/jwt"
	"github.com/Lestine-Yan/irisImg/backend/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

// 上下文键，业务侧可通过 c.GetString(ContextKeyUsername) 取出当前用户名。
const ContextKeyUsername = "username"

// JWTAuth 校验请求头中的 Bearer token。
// 通过则把 username 写入 gin.Context；否则中止请求并返回 401。
func JWTAuth(m *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			response.Unauthorized(c, "缺少 Authorization 请求头")
			c.Abort()
			return
		}

		// 严格要求 "Bearer <token>" 格式
		const prefix = "Bearer "
		if !strings.HasPrefix(header, prefix) {
			response.Unauthorized(c, "Authorization 格式应为 Bearer <token>")
			c.Abort()
			return
		}
		tokenStr := strings.TrimSpace(strings.TrimPrefix(header, prefix))
		if tokenStr == "" {
			response.Unauthorized(c, "token 不能为空")
			c.Abort()
			return
		}

		claims, err := m.Parse(tokenStr)
		if err != nil {
			response.Unauthorized(c, "token 无效或已过期")
			c.Abort()
			return
		}

		c.Set(ContextKeyUsername, claims.Username)
		c.Next()
	}
}
