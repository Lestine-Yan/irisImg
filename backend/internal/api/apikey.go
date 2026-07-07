package api

import (
	"errors"
	"strconv"

	"github.com/Lestine-Yan/irisImg/backend/internal/model"
	"github.com/Lestine-Yan/irisImg/backend/internal/pkg/response"
	"github.com/Lestine-Yan/irisImg/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// APIKeyAPI 是 API 密钥管理接口的控制器。
// 这些接口均挂在需 JWT 登录的受保护组下，并要求 HTTPS（由中间件保证）。
// 吊销 / 删除属于敏感操作，额外要求在请求体里携带账号密码做二次确认（authSvc 校验）。
type APIKeyAPI struct {
	svc     *service.APIKeyService
	authSvc *service.AuthService
}

// NewAPIKeyAPI 构造控制器。
func NewAPIKeyAPI(svc *service.APIKeyService, authSvc *service.AuthService) *APIKeyAPI {
	return &APIKeyAPI{svc: svc, authSvc: authSvc}
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

// Rename 处理 PATCH /apikeys/:id，重命名指定密钥。
func (h *APIKeyAPI) Rename(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "无效的密钥 ID")
		return
	}

	var req model.RenameAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	info, err := h.svc.Rename(c.Request.Context(), id, req.Name)
	if err != nil {
		if errors.Is(err, service.ErrKeyNotFound) {
			response.NotFound(c, "密钥不存在")
			return
		}
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, info)
}

// Reset 处理 POST /apikeys/:id/reset，重置密钥明文。
// 响应中包含新的明文密钥，仅此一次返回，调用方需自行妥善保存。
func (h *APIKeyAPI) Reset(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "无效的密钥 ID")
		return
	}

	resp, err := h.svc.Reset(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrKeyNotFound) {
			response.NotFound(c, "密钥不存在")
			return
		}
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, resp)
}

// Revoke 处理 POST /apikeys/:id/revoke，吊销指定密钥（软删除：仍展示但无法鉴权）。
// 需在请求体中携带账号密码做二次确认。
func (h *APIKeyAPI) Revoke(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "无效的密钥 ID")
		return
	}

	var req model.DestructiveAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.authSvc.VerifyCredentials(req.Username, req.Password); err != nil {
		response.Forbidden(c, "用户名或密码错误")
		return
	}

	if err := h.svc.Revoke(c.Request.Context(), id); err != nil {
		if errors.Is(err, service.ErrKeyNotFound) {
			response.NotFound(c, "密钥不存在")
			return
		}
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"id": id, "revoked": true})
}

// Delete 处理 DELETE /apikeys/:id，物理删除指定密钥并级联删除其关联图片。
// 需在请求体中携带账号密码做二次确认。响应附带被删除的图片数量。
func (h *APIKeyAPI) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "无效的密钥 ID")
		return
	}

	var req model.DestructiveAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.authSvc.VerifyCredentials(req.Username, req.Password); err != nil {
		response.Forbidden(c, "用户名或密码错误")
		return
	}

	removed, err := h.svc.Delete(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrKeyNotFound) {
			response.NotFound(c, "密钥不存在")
			return
		}
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"id": id, "deleted": true, "images_removed": removed})
}
