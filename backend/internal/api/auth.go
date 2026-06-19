package api

import (
	"errors"

	"github.com/Lestine-Yan/irisImg/backend/internal/model"
	"github.com/Lestine-Yan/irisImg/backend/internal/pkg/response"
	"github.com/Lestine-Yan/irisImg/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// AuthAPI 是认证相关接口的控制器。
type AuthAPI struct {
	authSvc *service.AuthService
}

// NewAuthAPI 构造控制器。
func NewAuthAPI(authSvc *service.AuthService) *AuthAPI {
	return &AuthAPI{authSvc: authSvc}
}

// Login 处理 POST /auth/login。
// 校验用户名密码，通过则返回 JWT。
func (h *AuthAPI) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := h.authSvc.Login(&req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.Unauthorized(c, "用户名或密码错误")
			return
		}
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, resp)
}

// Me 处理 GET /auth/me，返回当前登录用户信息。
// 由 JWT 中间件确保上下文中的 username 一定存在。
func (h *AuthAPI) Me(c *gin.Context) {
	username, _ := c.Get("username")
	response.Success(c, gin.H{"username": username})
}
