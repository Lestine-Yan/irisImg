package router

import (
	"github.com/Lestine-Yan/irisImg/backend/config"
	"github.com/Lestine-Yan/irisImg/backend/internal/api"
	"github.com/Lestine-Yan/irisImg/backend/internal/dao"
	"github.com/Lestine-Yan/irisImg/backend/internal/middleware"
	"github.com/Lestine-Yan/irisImg/backend/internal/pkg/jwt"
	"github.com/Lestine-Yan/irisImg/backend/internal/pkg/ratelimit"
	"github.com/Lestine-Yan/irisImg/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// New 组装并返回 gin.Engine。
// 依赖在这里手动注入，规模变大后可以引入 wire 等工具自动生成。
//
// imageDAO / apiKeyDAO 由调用方（main）基于已打开的数据库构造后注入。
func New(cfg *config.Config, imageDAO dao.ImageDAO, apiKeyDAO dao.APIKeyDAO) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), middleware.Logger(), middleware.CORS())

	// 依赖装配：config -> jwt.Manager / service -> api
	jwtMgr := jwt.NewManager(cfg.Auth.JWT)
	authSvc := service.NewAuthService(cfg.Auth, jwtMgr)
	authAPI := api.NewAuthAPI(authSvc)

	apiKeySvc := service.NewAPIKeyService(apiKeyDAO)
	apiKeyAPI := api.NewAPIKeyAPI(apiKeySvc)
	imageAPI := api.NewImageAPI(imageDAO)

	// 按密钥维度限流的内存令牌桶，默认阈值来自配置。
	rateStore := ratelimit.NewStore(cfg.APIKey.RateLimitPerMinute)

	// 路由分组：所有业务接口统一挂在 /api/v1 下
	v1 := r.Group("/api/v1")
	{
		v1.GET("/ping", api.Ping)

		// 公开的认证入口
		v1.POST("/auth/login", authAPI.Login)

		// 需要 JWT 登录后才能访问的受保护接口
		protected := v1.Group("", middleware.JWTAuth(jwtMgr))
		{
			protected.GET("/auth/me", authAPI.Me)

			// 密钥管理接口：受 JWT 保护 + 强制 HTTPS（生产由配置开启）
			keys := protected.Group("/apikeys", middleware.HTTPSOnly(cfg.APIKey.HTTPSOnly))
			{
				keys.POST("", apiKeyAPI.Create)
				keys.GET("", apiKeyAPI.List)
				keys.DELETE("/:id", apiKeyAPI.Revoke)
			}
		}

		// 图片接口：由 API 密钥鉴权中间件保护（独立于 JWT）。
		// GET 任意有效密钥可访问，POST 需读写密钥；当前为占位实现。
		images := v1.Group("/images", middleware.APIKeyAuth(apiKeySvc, rateStore))
		{
			images.GET("", imageAPI.List)
			images.POST("", imageAPI.Create)
		}
	}

	return r
}
