# `internal/middleware/https.go`

HTTPS 强制中间件。挂在密钥管理等敏感接口分组上，按配置项决定是否要求请求经由 HTTPS。

## 部署形态

生产部署由 **Nginx 统一做 HTTPS 反向代理**，后端本地走 HTTP。因此中间件不依赖后端自身是否监听 TLS，而是通过反代写入的 `X-Forwarded-Proto` 头做二次校验，同时兼容后端直接监听 TLS 的场景。

## 信任边界（防伪造）

`X-Forwarded-Proto` 是可被客户端任意伪造的请求头。中间件**不无条件信任**它：仅当请求的 TCP 对端（`RemoteAddr`）属于 `trustedProxies` 网段时才采信该头；否则只认 `c.Request.TLS`。这样即便后端端口被误暴露公网，攻击者伪造 `X-Forwarded-Proto: https` 也无法绕过--必须真 TLS。

`trustedProxies` 由 [`config.Server.TrustedProxies`](../../config/config.md) 经 [`config.ParseCIDRList`](../../config/config.md) 在启动期解析（[`cmd/server/main.go`](../../cmd/server.md) fail-fast），默认本地回环（同机反代）；跨机反代需在配置里追加反代所在 CIDR。空列表时退化为只认 `c.Request.TLS`（仍安全）。

## 函数

### `HTTPSOnly(enabled bool, trustedProxies []*net.IPNet) gin.HandlerFunc`

```go
if !enabled { c.Next(); return }
isHTTPS := c.Request.TLS != nil ||
    (isFromTrustedProxy(c.Request, trustedProxies) && c.GetHeader("X-Forwarded-Proto") == "https")
if !isHTTPS {
    response.Fail(c, http.StatusForbidden, response.CodeForbidden, "该接口要求使用 HTTPS 访问")
    c.Abort()
    return
}
c.Next()
```

- `enabled` 为 false 时直接放行（本地开发）。
- `enabled` 为 true 时，`c.Request.TLS != nil`（直连 TLS）或「来自可信反代且 `X-Forwarded-Proto == "https"`」任一满足即放行；否则 403 `CodeForbidden`(40300)。
- `enabled` 来自 [`apikey.https_only`](../../config/config.md)，`trustedProxies` 来自 [`server.trusted_proxies`](../../config/config.md)，由 [`router`](../router/router.md) 注入。

### `isFromTrustedProxy(r *http.Request, trustedProxies []*net.IPNet) bool`

取 `r.RemoteAddr` 经 `net.SplitHostPort` 去端口，`net.ParseIP` 后对各 CIDR 做 `Contains`。`trustedProxies` 为空时返回 false（直连场景，HTTPSOnly 只认 `c.Request.TLS`）。

## 与其它包的关系

```
main.go ──config.ParseCIDRList(cfg.Server.TrustedProxies)──► []*net.IPNet
router.New ──HTTPSOnly(cfg.APIKey.HTTPSOnly, trustedProxies)──► 挂在 /apikeys 与 /admin/logs 管理组（JWT 之后）
```

## 注意

- 信任边界以 TCP 对端为准，不依赖 Nginx 是否覆盖该头：即使反代未覆盖（透传客户端原值），不可信 peer 仍被拒。Nginx 用 `proxy_set_header X-Forwarded-Proto $scheme` 覆盖该头是额外的纵深防御，非必需。
- 当前仅密钥管理组与日志组开启；如需对全站强制，可在 router 顶层 `r.Use` 挂载。
