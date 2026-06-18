package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger 记录每个请求的关键信息。
// 这里只用标准库 log，后续可替换为 zap / logrus。
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		log.Printf("[%s] %s %d %s",
			c.Request.Method,
			path,
			c.Writer.Status(),
			time.Since(start),
		)
	}
}
