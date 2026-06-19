# `internal/model/auth.go`

认证相关的 DTO。不放任何业务行为，只描述「请求/响应在 JSON 上长什么样」以及 Gin validator 的约束。

## `LoginRequest`

```go
type LoginRequest struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
}
```

- 用 Gin 自带 validator 校验非空；缺字段会被 [`AuthAPI.Login`](../api/auth.md) 转成 400。
- **不**做用户名/密码格式正则校验：账号是配置文件里写死的，请求里只要任一项为空就视为非法即可，其余对错由 `AuthService` 判定。

## `LoginResponse`

```go
type LoginResponse struct {
    Token     string `json:"token"`
    TokenType string `json:"token_type"`   // 固定 "Bearer"
    ExpiresAt int64  `json:"expires_at"`   // Unix 秒
}
```

- `token` 由 [`pkg/jwt`](../pkg/jwt.md) 签发的 HS256 JWT 字符串。
- `token_type` 写死 `"Bearer"`，前端拼接 `Authorization: Bearer <token>` 时一致。
- `expires_at` 是过期时刻（Unix 秒），前端可据此判断是否需要重新登录，或对照本地时钟刷新。

## 修改建议

- 后续要返回更多上下文（昵称、头像、权限），优先扩 `LoginResponse`，不要新增 `/auth/userinfo` 类似接口。
- 添加新的入参 DTO 时，文件就近放在本目录下，命名遵循 `<feature>.go`。
