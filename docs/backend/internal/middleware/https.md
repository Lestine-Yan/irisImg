# `internal/middleware/https.go`

HTTPS 强制中间件。挂在密钥管理等敏感接口分组上，按配置项决定是否要求请求经由 HTTPS。

## 部署形态

生产部署由 **Nginx 统一做 HTTPS 反向代理**，后端本地走 HTTP。因此中间件不依赖后端自身是否监听 TLS，而是通过反代写入的 `X-Forwarded-Proto` 头做二次校验，同时兼容后端直接监听 TLS 的场景。

## 函数

### `HTTPSOnly(enabled bool) gin.HandlerFunc`

```go
if !enabled { c.Next(); return }
isHTTPS := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
if !isHTTPS {
    response.Fail(c, http.StatusForbidden, response.CodeForbidden, "该接口要求使用 HTTPS 访问")
    c.Abort()
    return
}
c.Next()
```

- `enabled` 为 false 时直接放行（本地开发）。
- `enabled` 为 true 时，`c.Request.TLS != nil`（直连 TLS）或 `X-Forwarded-Proto == "https"`（Nginx 反代）任一满足即放行；否则 403 `CodeForbidden`(40300)。
- `enabled` 来自配置 [`apikey.https_only`](../../config/config.md)，由 [`router`](../router/router.md) 注入：本地开发置 false，生产置 true。

## 与其它包的关系

```
router.New ──HTTPSOnly(cfg.APIKey.HTTPSOnly)──► 挂在 /apikeys 管理组（JWT 之后）
```

## 注意

- `X-Forwarded-Proto` 可被客户端伪造，**只有在请求确实经过受信任的反代时才安全**。生产环境务必让 Nginx 覆盖（而非透传）该头。
- 当前仅密钥管理组开启；如需对全站强制，可在 router 顶层 `r.Use` 挂载。
