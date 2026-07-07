package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORS 是一个最简的跨域中间件，方便前后端分离开发联调。
// 生产环境建议替换为更严格的策略或使用 github.com/gin-contrib/cors。
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
