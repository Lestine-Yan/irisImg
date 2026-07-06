package router

import (
	"github.com/Lestine-Yan/irisImg/backend/config"
	"github.com/Lestine-Yan/irisImg/backend/internal/api"
	"github.com/Lestine-Yan/irisImg/backend/internal/dao"
	"github.com/Lestine-Yan/irisImg/backend/internal/middleware"
	"github.com/Lestine-Yan/irisImg/backend/internal/pkg/jwt"
	"github.com/Lestine-Yan/irisImg/backend/internal/pkg/ratelimit"
	"github.com/Lestine-Yan/irisImg/backend/internal/pkg/storage"
	"github.com/Lestine-Yan/irisImg/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// New 组装并返回 gin.Engine。
// 依赖在这里手动注入，规模变大后可以引入 wire 等工具自动生成。
//
// imageDAO / apiKeyDAO 由调用方（main）基于已打开的数据库构造后注入；
// saver 由调用方基于配置构造（启动失败可早期暴露权限/路径问题）。
func New(cfg *config.Config, imageDAO dao.ImageDAO, apiKeyDAO dao.APIKeyDAO, saver *storage.Saver) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), middleware.Logger(), middleware.CORS())

	// 静态图片服务：开发期由后端直接 serve 落盘目录，前端可加载 /imgs/<rel> 缩略图与大图。
	// 生产建议由 Nginx 反代 /imgs/（见 docs/backend/IMAGE.md），此处兜底，不影响 Nginx 优先拦截。
	r.Static("/imgs", cfg.Storage.RootDir)

	// 依赖装配：config -> jwt.Manager / service -> api
	jwtMgr := jwt.NewManager(cfg.Auth.JWT)
	authSvc := service.NewAuthService(cfg.Auth, jwtMgr)
	authAPI := api.NewAuthAPI(authSvc)

	apiKeySvc := service.NewAPIKeyService(apiKeyDAO)
	apiKeyAPI := api.NewAPIKeyAPI(apiKeySvc)

	imageSvc := service.NewImageService(imageDAO, saver, cfg.Storage)
	imageAPI := api.NewImageAPI(imageSvc)

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

			// 图片管理接口（后台）：受 JWT 保护，供内容中心拉取图片列表、后台直传上传。
			// 与对外 /images（API Key 鉴权）解耦，避免后台页面被迫注入 X-API-Key。
			// 路径用 /admin/images 而非 /images，后者已被 APIKeyAuth 组占用，重复注册会冲突。
			protected.GET("/admin/images", imageAPI.ListAdmin)
			protected.POST("/admin/images", imageAPI.CreateAdmin)
		}

		// 图片接口：由 API 密钥鉴权中间件保护（独立于 JWT）。
		// GET 任意有效密钥可访问，POST 需读写密钥；POST 已落地，GET 仍为占位。
		images := v1.Group("/images", middleware.APIKeyAuth(apiKeySvc, rateStore))
		{
			images.GET("", imageAPI.List)
			images.POST("", imageAPI.Create)
		}
	}

	return r
}
