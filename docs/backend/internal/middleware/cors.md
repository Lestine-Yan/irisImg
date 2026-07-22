# `internal/middleware/cors.go`

基于 [`github.com/gin-contrib/cors`](https://github.com/gin-contrib/cors) 的跨域中间件，按配置 [`cors.allow_origins`](../../config/config.md) 收紧来源白名单，替换原先无条件的 `Access-Control-Allow-Origin: *`。

## 函数

### `CORS(allowOrigins []string) gin.HandlerFunc`

- **含 `"*"`**：回显 `Access-Control-Allow-Origin: *`，开发联调用。release 模式被 [`config.Validate`](../../config/config.md) 拒绝启动，仅 debug 可达。
- **确切 origin 列表**（如 `["https://img.example.com"]`）：仅命中项回显具体 origin；未命中 gin-contrib/cors 直接返回 403 拒绝跨域。
- **空**：关闭跨域，返回 no-op（不写任何 CORS 头）。生产同域部署无跨域需求，留空即可。

统一配置：

```go
cors.Config{
    AllowOrigins:     allowOrigins,
    AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
    AllowCredentials: false,   // 鉴权走 Authorization: Bearer 头，无 cookie/session
    MaxAge:           12 * time.Hour,
}
```

OPTIONS 预检由 gin-contrib/cors 自动处理并返回 204。

## 鉴权模型与安全边界

- JWT 走 `Authorization: Bearer` 头（[`service/auth`](../service/auth.md)），API Key 走 `X-API-Key` 头（[`middleware/apikey`](./apikey.md)），**全程无 cookie/session**。
- 故不启用 `AllowCredentials`：浏览器不会自动附带凭据到跨域请求，恶意网站无法借用受害者已登录会话发起带凭据的跨域请求。
- `AllowHeaders` 保留 `Authorization` 以放行前端 Bearer 请求；跨域 JS 要发送它必须显式填值，而受害者的 JWT 存在 localStorage（同源策略下跨域读不到），故不构成 CSRF。

## release fail-closed

[`config.Validate`](../../config/config.md) 在 release 模式下拒绝 `allow_origins` 含 `*`，闭合「通配 CORS 上线」攻击链：`Allow-Origin: *` 配合 `Authorization` 透传是定时炸弹，一旦将来引入 cookie 鉴权即升为高危。生产同域部署留空关闭跨域最安全。

## 与其它包的关系

```
router.New ──CORS(cfg.CORS.AllowOrigins)──► 全局中间件链（RequestID -> Recovery -> CORS -> Logger）
```
