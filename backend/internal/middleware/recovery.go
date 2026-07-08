package middleware

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/Lestine-Yan/irisImg/backend/internal/model"
	"github.com/Lestine-Yan/irisImg/backend/internal/pkg/logger"
	"github.com/Lestine-Yan/irisImg/backend/internal/pkg/response"
	"github.com/Lestine-Yan/irisImg/backend/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery 捕获 panic：用 zap 记录堆栈，并落库一条 event=panic 的日志，再返回 500。
// 替换 gin.Recovery()，使 panic 也进入日志中心可查。
func Recovery(l *logger.Logger, svc *service.LogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				if l != nil {
					l.Error(c.Request.Context(), "panic recovered",
						zap.Any("panic", r),
						zap.String("stack", string(stack)),
						zap.String("method", c.Request.Method),
						zap.String("path", c.Request.URL.Path),
					)
				}
				if svc != nil {
					lc := LogContextFromGin(c)
					svc.Record(&model.Log{
						Timestamp: time.Now(),
						Level:     model.LevelError,
						Event:     model.EventPanic,
						Method:    c.Request.Method,
						Path:      c.Request.URL.Path,
						Message:   fmt.Sprintf("panic: %v", r),
						RequestID: lc.RequestID,
						Username:  lc.Username,
						APIKeyID:  lc.APIKeyID,
						ClientIP:  lc.ClientIP,
					})
				}
				response.ServerError(c, "服务器内部错误")
				c.Abort()
			}
		}()
		c.Next()
	}
}
