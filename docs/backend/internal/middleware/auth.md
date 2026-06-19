# `internal/middleware/auth.go`

JWT 鉴权中间件。挂在受保护路由分组上，校验请求头里的 Bearer token 并把用户名注入 `gin.Context`。

## 常量

### `ContextKeyUsername = "username"`

业务代码在受保护路由里取当前用户名时统一用这个键，避免到处硬编码字符串：

```go
username := c.GetString(middleware.ContextKeyUsername)
```

## 函数

### `JWTAuth(m *jwt.Manager) gin.HandlerFunc`

- 闭包持有 [`*jwt.Manager`](../pkg/jwt.md)，由 [`router`](../router/router.md) 注入。
- 处理流程：
  1. 读 `Authorization` 请求头；为空 → `response.Unauthorized("缺少 Authorization 请求头")` + `c.Abort()`。
  2. 必须以 `"Bearer "` 前缀开头；否则 → `response.Unauthorized("Authorization 格式应为 Bearer <token>")`。
  3. 截掉前缀并 `TrimSpace`；空 token → `response.Unauthorized("token 不能为空")`。
  4. 调 `m.Parse(tokenStr)`；任意错误 → `response.Unauthorized("token 无效或已过期")`。
  5. 成功后 `c.Set(ContextKeyUsername, claims.Username)`，再 `c.Next()`。

所有失败分支统一走 [`response.Unauthorized`](../pkg/response.md) → HTTP 401，`code = 40100`。

## 注意

- 错误信息有意做得粗粒度（无效或过期），不暴露具体原因（签名错？过期？发行者不对？），减少猜测空间。
- 中间件不区分用户角色——本项目只有一个用户，进得来就是管理员；引入多用户/角色时应在解析后再做角色判断。
