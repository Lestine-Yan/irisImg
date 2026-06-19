package router

import (
	"github.com/Lestine-Yan/irisImg/backend/config"
	"github.com/Lestine-Yan/irisImg/backend/internal/api"
	"github.com/Lestine-Yan/irisImg/backend/internal/middleware"
	"github.com/Lestine-Yan/irisImg/backend/internal/pkg/jwt"
	"github.com/Lestine-Yan/irisImg/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// New 组装并返回 gin.Engine。
// 依赖在这里手动注入，规模变大后可以引入 wire 等工具自动生成。
func New(cfg *config.Config) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), middleware.Logger(), middleware.CORS())

	// 依赖装配：config -> jwt.Manager -> service -> api
	jwtMgr := jwt.NewManager(cfg.Auth.JWT)
	authSvc := service.NewAuthService(cfg.Auth, jwtMgr)
	authAPI := api.NewAuthAPI(authSvc)

	// 路由分组：所有业务接口统一挂在 /api/v1 下
	v1 := r.Group("/api/v1")
	{
		v1.GET("/ping", api.Ping)

		// 公开的认证入口
		v1.POST("/auth/login", authAPI.Login)

		// 需要登录后才能访问的接口都挂到 protected 下
		protected := v1.Group("", middleware.JWTAuth(jwtMgr))
		{
			protected.GET("/auth/me", authAPI.Me)
		}
	}

	return r
}
