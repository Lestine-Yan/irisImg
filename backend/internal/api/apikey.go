package api

import (
	"errors"
	"strconv"

	"github.com/Lestine-Yan/irisImg/backend/internal/dao"
	"github.com/Lestine-Yan/irisImg/backend/internal/model"
	"github.com/Lestine-Yan/irisImg/backend/internal/pkg/response"
	"github.com/Lestine-Yan/irisImg/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// APIKeyAPI 是 API 密钥管理接口的控制器。
// 这些接口均挂在需 JWT 登录的受保护组下，并要求 HTTPS（由中间件保证）。
type APIKeyAPI struct {
	svc *service.APIKeyService
}

// NewAPIKeyAPI 构造控制器。
func NewAPIKeyAPI(svc *service.APIKeyService) *APIKeyAPI {
	return &APIKeyAPI{svc: svc}
}

// Create 处理 POST /apikeys，创建一把新密钥。
// 响应中包含明文密钥，仅此一次返回，调用方需自行妥善保存。
func (h *APIKeyAPI) Create(c *gin.Context) {
	var req model.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := h.svc.Create(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidScope) {
			response.BadRequest(c, "scope 仅支持 readonly 或 readwrite")
			return
		}
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, resp)
}

// List 处理 GET /apikeys，返回全部密钥（不含明文与哈希）。
func (h *APIKeyAPI) List(c *gin.Context) {
	infos, err := h.svc.List(c.Request.Context())
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"items": infos})
}

// Revoke 处理 DELETE /apikeys/:id，吊销指定密钥。
func (h *APIKeyAPI) Revoke(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "无效的密钥 ID")
		return
	}

	if err := h.svc.Revoke(c.Request.Context(), id); err != nil {
		if errors.Is(err, dao.ErrNotFound) {
			response.NotFound(c, "密钥不存在")
			return
		}
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"id": id, "revoked": true})
}
