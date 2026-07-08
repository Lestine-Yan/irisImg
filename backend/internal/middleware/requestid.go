package middleware

import (
	"github.com/Lestine-Yan/irisImg/backend/internal/model"
	"github.com/Lestine-Yan/irisImg/backend/internal/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ContextKeyRequestID 是存放 request id 的 gin.Context 键。
const ContextKeyRequestID = "request_id"

// HeaderRequestID 是 request id 请求 / 响应头名称。
const HeaderRequestID = "X-Request-Id"

// RequestID 为每个请求生成或透传 request id：
//   - 优先沿用客户端传入的 X-Request-Id，否则用 uuid 生成；
//   - 回写响应头 X-Request-Id，并写入 gin.Context（供 LogContextFromGin 读取）；
//   - 写入 c.Request.Context()（供 logger 自动附加 request_id 字段）。
//
// 使访问日志、业务事件、panic 经 request id 关联同一请求。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(HeaderRequestID)
		if rid == "" {
			rid = uuid.NewString()
		}
		c.Header(HeaderRequestID, rid)
		c.Set(ContextKeyRequestID, rid)
		c.Request = c.Request.WithContext(logger.ContextWithRequestID(c.Request.Context(), rid))
		c.Next()
	}
}

// LogContextFromGin 从 gin.Context 抽取写日志时附带的关联信息
// （request id / 用户名 / 来源密钥 / 客户端 IP），供访问日志中间件与业务事件控制器复用。
func LogContextFromGin(c *gin.Context) model.LogContext {
	lc := model.LogContext{
		RequestID: c.GetString(ContextKeyRequestID),
		Username:  c.GetString(ContextKeyUsername),
		ClientIP:  c.ClientIP(),
	}
	if v, ok := c.Get(ContextKeyAPIKeyID); ok {
		if id, ok := v.(int); ok {
			lc.APIKeyID = &id
		}
	}
	return lc
}
