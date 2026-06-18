package router

import (
	"github.com/Lestine-Yan/irisImg/backend/internal/api"
	"github.com/Lestine-Yan/irisImg/backend/internal/dao"
	"github.com/Lestine-Yan/irisImg/backend/internal/middleware"
	"github.com/Lestine-Yan/irisImg/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// New 组装并返回 gin.Engine。
// 依赖在这里手动注入，规模变大后可以引入 wire 等工具自动生成。
func New() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), middleware.Logger(), middleware.CORS())

	// 依赖装配：dao -> service -> api
	userDAO := dao.NewMemoryUserDAO()
	userSvc := service.NewUserService(userDAO)
	userAPI := api.NewUserAPI(userSvc)

	// 路由分组：所有业务接口统一挂在 /api/v1 下
	v1 := r.Group("/api/v1")
	{
		v1.GET("/ping", api.Ping)

		users := v1.Group("/users")
		{
			users.POST("", userAPI.Create)
			users.GET("", userAPI.List)
			users.GET("/:id", userAPI.Get)
		}
	}

	return r
}
