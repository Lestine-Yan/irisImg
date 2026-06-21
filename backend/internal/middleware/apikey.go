package middleware

import (
	"errors"
	"net/http"

	"github.com/Lestine-Yan/irisImg/backend/internal/model"
	"github.com/Lestine-Yan/irisImg/backend/internal/pkg/ratelimit"
	"github.com/Lestine-Yan/irisImg/backend/internal/pkg/response"
	"github.com/Lestine-Yan/irisImg/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// HeaderAPIKey 是携带 API 密钥的请求头名称。
const HeaderAPIKey = "X-API-Key"

// 上下文键，业务侧可通过 c.GetInt(ContextKeyAPIKeyID) 取出当前密钥 ID，
// 用于落库「图片由哪个密钥添加」。
const ContextKeyAPIKeyID = "api_key_id"

// APIKeyAuth 是 API 密钥鉴权中间件，按以下顺序校验，任一步失败即中止并返回区分的错误码：
//  1. header 缺失 / 为空       -> 401 CodeAPIKeyMissing
//  2. 格式非法                 -> 401 CodeAPIKeyInvalid
//  3. 密钥不存在 / 已吊销      -> 401 CodeAPIKeyInvalid
//  4. 权限不足（非 GET 需读写）-> 403 CodeForbidden
//  5. 触发限流                 -> 429 CodeTooManyRequests
//
// 权限规则：GET 请求任意有效密钥均可访问；非 GET（POST 等）必须为 readwrite 密钥。
func APIKeyAuth(svc *service.APIKeyService, limiter *ratelimit.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 取出 header
		raw := c.GetHeader(HeaderAPIKey)
		if raw == "" {
			response.Fail(c, http.StatusUnauthorized, response.CodeAPIKeyMissing, "缺少 "+HeaderAPIKey+" 请求头")
			c.Abort()
			return
		}

		// 2 & 3. 格式校验 + 查库 + 吊销判定
		key, err := svc.Authenticate(c.Request.Context(), raw)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrInvalidKeyFormat):
				response.Fail(c, http.StatusUnauthorized, response.CodeAPIKeyInvalid, "密钥格式非法")
			case errors.Is(err, service.ErrKeyNotFound):
				response.Fail(c, http.StatusUnauthorized, response.CodeAPIKeyInvalid, "密钥不存在")
			case errors.Is(err, service.ErrKeyRevoked):
				response.Fail(c, http.StatusUnauthorized, response.CodeAPIKeyInvalid, "密钥已吊销")
			default:
				response.ServerError(c, "密钥校验失败")
			}
			c.Abort()
			return
		}

		// 4. 权限校验：非 GET 请求需要读写密钥
		if c.Request.Method != http.MethodGet && key.Scope != model.ScopeReadWrite {
			response.Forbidden(c, "只读密钥无权访问该接口")
			c.Abort()
			return
		}

		// 5. 限流校验：按密钥维度令牌桶
		if !limiter.Allow(key.ID, key.RateLimit) {
			response.TooManyRequests(c, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}

		// 通过：记录来源密钥 ID，并尽力更新最近使用时间（失败不阻断主流程）。
		c.Set(ContextKeyAPIKeyID, key.ID)
		_ = svc.Touch(c.Request.Context(), key.ID)
		c.Next()
	}
}
