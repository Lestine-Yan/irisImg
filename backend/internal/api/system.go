package api

import (
	"github.com/Lestine-Yan/irisImg/backend/internal/pkg/response"
	"github.com/Lestine-Yan/irisImg/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// SystemAPI 是系统配置只读接口的控制器，挂在 JWT 受保护组下。
//
// 仅暴露当前 config 的非敏感快照，不支持修改 / 热更新；配置变更需改 config 文件并重启。
type SystemAPI struct {
	svc *service.SystemService
}

// NewSystemAPI 构造控制器。
func NewSystemAPI(svc *service.SystemService) *SystemAPI {
	return &SystemAPI{svc: svc}
}

// Config 处理 GET /system/config，返回当前系统配置只读视图。
func (h *SystemAPI) Config(c *gin.Context) {
	response.Success(c, h.svc.Config())
}
