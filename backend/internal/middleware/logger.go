package middleware

import (
	"time"

	"github.com/Lestine-Yan/irisImg/backend/internal/model"
	"github.com/Lestine-Yan/irisImg/backend/internal/pkg/logger"
	"github.com/Lestine-Yan/irisImg/backend/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger 是结构化访问日志中间件：
//   - c.Next() 后用 zap 输出结构化访问日志（method/path/status/duration/bytes/username 等）
//     到 stdout / 文件，request_id 由 logger 从 ctx 自动附加；
//   - 同时把一条 event=http.request 的日志异步落库，供日志中心查询。
//
// 跳过 /api/v1/ping 健康检查以减噪。
func Logger(l *logger.Logger, svc *service.LogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/api/v1/ping" {
			c.Next()
			return
		}

		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		status := c.Writer.Status()
		duration := time.Since(start)
		level := levelFromStatus(status)
		lc := LogContextFromGin(c)

		if l != nil {
			fields := []zap.Field{
				zap.String("method", c.Request.Method),
				zap.String("path", path),
				zap.Int("status", status),
				zap.Duration("duration", duration),
				zap.Int("bytes", c.Writer.Size()),
			}
			if lc.Username != "" {
				fields = append(fields, zap.String("username", lc.Username))
			}
			if lc.APIKeyID != nil {
				fields = append(fields, zap.Int("api_key_id", *lc.APIKeyID))
			}
			switch level {
			case model.LevelError:
				l.Error(c.Request.Context(), "http.request", fields...)
			case model.LevelWarn:
				l.Warn(c.Request.Context(), "http.request", fields...)
			default:
				l.Info(c.Request.Context(), "http.request", fields...)
			}
		}

		if svc != nil {
			statusVal := status
			durationMs := int(duration.Milliseconds())
			svc.Record(&model.Log{
				Timestamp:  start,
				Level:      level,
				Event:      model.EventHTTPRequest,
				Method:     c.Request.Method,
				Path:       path,
				Status:     &statusVal,
				DurationMs: &durationMs,
				ClientIP:   lc.ClientIP,
				RequestID:  lc.RequestID,
				APIKeyID:   lc.APIKeyID,
				Username:   lc.Username,
			})
		}
	}
}

// levelFromStatus 按 HTTP 状态码推导日志级别：≥500 error、≥400 warn、否则 info。
func levelFromStatus(status int) string {
	switch {
	case status >= 500:
		return model.LevelError
	case status >= 400:
		return model.LevelWarn
	default:
		return model.LevelInfo
	}
}
