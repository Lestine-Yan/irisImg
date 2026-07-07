# `internal/middleware/cors.go`

最简单的跨域中间件，方便前后端分离开发联调。

## 函数

### `CORS() gin.HandlerFunc`

每个请求统一写入：

- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Methods: GET, POST, PUT, PATCH, DELETE, OPTIONS`
- `Access-Control-Allow-Headers: Origin, Content-Type, Accept, Authorization`

当方法是 `OPTIONS` 时直接 `c.AbortWithStatus(204)` 返回预检响应，否则继续 `c.Next()`。

## 注意

- `Allow-Origin: *` 仅适合开发期，**不要用于生产**：携带 `Authorization` 凭证的请求在浏览器侧会被严格策略拦截。
- 上线前替换为白名单方案，建议引入 `github.com/gin-contrib/cors` 或在此基础上读取 `config` 里的允许域名。
