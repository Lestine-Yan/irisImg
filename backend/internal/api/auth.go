package api

import (
	"errors"

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

// NewAuthAPI 构造控制器。rec 用于记录登录成功 / 失败业务事件到日志中心。
func NewAuthAPI(authSvc *service.AuthService, rec service.LogRecorder) *AuthAPI {
	return &AuthAPI{authSvc: authSvc, rec: rec}
}

// Login 处理 POST /auth/login。
// 校验用户名密码，通过则返回 JWT，并记录 auth.login_success / auth.login_failed 事件。
func (h *AuthAPI) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := h.authSvc.Login(&req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			h.recordLogin(c, req.Username, false)
			response.Unauthorized(c, "用户名或密码错误")
			return
		}
		response.ServerError(c, err.Error())
		return
	}
	h.recordLogin(c, req.Username, true)
	response.Success(c, resp)
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

// Me 处理 GET /auth/me，返回当前登录用户信息。
// 由 JWT 中间件确保上下文中的 username 一定存在。
func (h *AuthAPI) Me(c *gin.Context) {
	username, _ := c.Get("username")
	response.Success(c, gin.H{"username": username})
}
