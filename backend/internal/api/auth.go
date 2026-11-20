package api

import (
	"errors"
	"math"
	"strconv"

	"github.com/Lestine-Yan/irisImg/backend/internal/middleware"
	"github.com/Lestine-Yan/irisImg/backend/internal/model"
	"github.com/Lestine-Yan/irisImg/backend/internal/pkg/response"
	"github.com/Lestine-Yan/irisImg/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// AuthAPI 是认证相关接口的控制器。
type AuthAPI struct {
	authSvc *service.AuthService
	rec     service.LogRecorder
}

// NewAuthAPI 构造控制器。rec 用于记录登录成功 / 失败 / 锁定业务事件到日志中心。
func NewAuthAPI(authSvc *service.AuthService, rec service.LogRecorder) *AuthAPI {
	return &AuthAPI{authSvc: authSvc, rec: rec}
}

// Login 处理 POST /auth/login。
// 校验用户名密码，通过则返回 JWT，并记录 auth.login_success / auth.login_failed 事件。
//
// 连续失败达到阈值后进入锁定：锁定期间返回 429 + Retry-After，即使密码正确也拒绝。
// 触发锁定的那一次仍按 401 返回（便于前端提示「用户名或密码错误」），下一次起变为 429。
func (h *AuthAPI) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	clientIP := c.ClientIP()
	resp, err := h.authSvc.Login(&req, clientIP)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrLoginLocked):
			// 锁定期间被拒的请求只进访问日志（429），不重复写业务事件，避免爆破流量放大日志。
			h.writeRetryAfter(c, clientIP)
			response.TooManyRequests(c, "登录尝试过多，请稍后再试")
		case errors.Is(err, service.ErrInvalidCredentials):
			h.recordLogin(c, req.Username, false)
			// 本次失败可能刚好触发锁定，补记一条锁定事件。
			if _, locked := h.authSvc.LockoutRemaining(clientIP); locked {
				h.recordLocked(c, req.Username)
			}
			response.Unauthorized(c, "用户名或密码错误")
		default:
			response.ServerError(c, err.Error())
		}
		return
	}
	h.recordLogin(c, req.Username, true)
	response.Success(c, resp)
}

// writeRetryAfter 按该 IP 的剩余锁定时间写入 Retry-After 响应头（秒，向上取整）。
func (h *AuthAPI) writeRetryAfter(c *gin.Context, clientIP string) {
	remaining, locked := h.authSvc.LockoutRemaining(clientIP)
	if !locked {
		return
	}
	seconds := int(math.Ceil(remaining.Seconds()))
	if seconds < 1 {
		seconds = 1
	}
	c.Header("Retry-After", strconv.Itoa(seconds))
}

// recordLogin 记录一次登录尝试。attemptedUsername 是请求体里填写的用户名
// （未通过鉴权也记录，供审计）。rec 为空时直接返回，便于测试。
func (h *AuthAPI) recordLogin(c *gin.Context, attemptedUsername string, success bool) {
	if h.rec == nil {
		return
	}
	lc := middleware.LogContextFromGin(c)
	lc.Username = attemptedUsername
	event, level, msg := model.EventAuthLoginOK, model.LevelInfo, "login success"
	if !success {
		event, level, msg = model.EventAuthLoginFail, model.LevelWarn, "login failed"
	}
	h.rec.Record(model.NewEventLog(event, level, msg, lc))
}

// recordLocked 记录一次「登录锁定」事件。只在触发锁定的那次失败 / 锁定期间的请求上调用；
// 锁定期间的重复请求也要记录，便于在日志中心看到被挡下的爆破流量。
func (h *AuthAPI) recordLocked(c *gin.Context, attemptedUsername string) {
	if h.rec == nil {
		return
	}
	lc := middleware.LogContextFromGin(c)
	lc.Username = attemptedUsername
	h.rec.Record(model.NewEventLog(model.EventAuthLoginLocked, model.LevelWarn, "login locked", lc))
}

// Me 处理 GET /auth/me，返回当前登录用户信息。
// 由 JWT 中间件确保上下文中的 username 一定存在。
func (h *AuthAPI) Me(c *gin.Context) {
	username, _ := c.Get("username")
	response.Success(c, gin.H{"username": username})
}
