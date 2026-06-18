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

// UserAPI 是用户相关接口的控制器。
// 控制器只做：参数解析/校验、调用 service、组装响应。
type UserAPI struct {
	userSvc *service.UserService
}

// NewUserAPI 构造控制器。
func NewUserAPI(userSvc *service.UserService) *UserAPI {
	return &UserAPI{userSvc: userSvc}
}

// Create 处理 POST /users。
func (h *UserAPI) Create(c *gin.Context) {
	var req model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	u, err := h.userSvc.Create(&req)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, u)
}

// Get 处理 GET /users/:id。
func (h *UserAPI) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	u, err := h.userSvc.GetByID(id)
	if err != nil {
		if errors.Is(err, dao.ErrNotFound) {
			response.NotFound(c, "user not found")
			return
		}
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, u)
}

// List 处理 GET /users。
func (h *UserAPI) List(c *gin.Context) {
	users, err := h.userSvc.List()
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, users)
}
