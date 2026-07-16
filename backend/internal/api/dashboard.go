package api

import (
	"strconv"

	"github.com/Lestine-Yan/irisImg/backend/internal/pkg/response"
	"github.com/Lestine-Yan/irisImg/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// DashboardAPI 是仪表盘统计接口的控制器，挂在 JWT 受保护组下，只读聚合。
type DashboardAPI struct {
	svc *service.DashboardService
}

// NewDashboardAPI 构造控制器。
func NewDashboardAPI(svc *service.DashboardService) *DashboardAPI {
	return &DashboardAPI{svc: svc}
}

// Overview 处理 GET /admin/dashboard，返回仪表盘首页所需的聚合统计。
//
//	Query 参数:
//	  - days: 趋势窗口天数，默认 30，上限 90（防滥用）；非法值（非数字 / <=0 / >90）一律回退到 30。
//
// 响应: model.DashboardOverview（见 internal/model/dashboard.go）。
func (h *DashboardAPI) Overview(c *gin.Context) {
	days, err := strconv.Atoi(c.DefaultQuery("days", "30"))
	if err != nil || days <= 0 || days > 90 {
		days = 30
	}
	overview, err := h.svc.Overview(c.Request.Context(), days)
	if err != nil {
		response.ServerError(c, "查询仪表盘数据失败："+err.Error())
		return
	}
	response.Success(c, overview)
}
