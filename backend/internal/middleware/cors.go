package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS 构造跨域中间件。allowOrigins 为来源白名单：
//   - 含 "*"：开发联调用，回显 Access-Control-Allow-Origin: *（release 模式被 Validate
//     拒绝启动，仅 debug 可达）。
//   - 确切 origin 列表（如 ["https://img.example.com"]）：仅命中项回显具体 origin，
//     未命中不写 Allow-Origin（浏览器拒绝跨域读）。
//   - 空：关闭跨域，返回 no-op（不写任何 CORS 头）。生产同域部署无跨域需求，留空即可。
//
// 鉴权走 Authorization: Bearer 头（无 cookie/session），故不启用 AllowCredentials；
// AllowHeaders 保留 Authorization 以放行前端 Bearer 请求。预检（OPTIONS）由 gin-contrib/cors
// 自动处理并返回 204。
func CORS(allowOrigins []string) gin.HandlerFunc {
	if len(allowOrigins) == 0 {
		// 生产同域部署：无跨域需求，不写任何 CORS 头，浏览器默认拒绝跨域读。
		return func(c *gin.Context) { c.Next() }
	}
	return cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	})
}
