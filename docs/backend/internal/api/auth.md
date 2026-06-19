# `internal/api/auth.go`

认证相关的 Gin 控制器。控制器层只做参数解析、调用 service、组装响应，不写业务规则。

## 类型

### `AuthAPI`

- 字段：`authSvc *service.AuthService`
- 由 [`router`](../router/router.md) 通过 `NewAuthAPI(authSvc)` 注入。

## 处理函数

### `Login(c *gin.Context)` —— `POST /api/v1/auth/login`

1. `c.ShouldBindJSON(&req)` 解析 [`model.LoginRequest`](../model/auth.md)；失败 → `response.BadRequest(err)`。
2. 调 `authSvc.Login(&req)`：
   - `errors.Is(err, service.ErrInvalidCredentials)` → `response.Unauthorized("用户名或密码错误")`
   - 其它非 nil err → `response.ServerError(err)`
3. 成功时 `response.Success(c, resp)`，`resp` 是 `*model.LoginResponse`，含 `token / token_type / expires_at`。

### `Me(c *gin.Context)` —— `GET /api/v1/auth/me`

- 受 [`middleware.JWTAuth`](../middleware/auth.md) 保护，进入此函数时 token 已校验通过。
- 从 `c.Get("username")` 取出当前用户名，返回 `gin.H{"username": ...}`。
- 同时也是「校验 token 是否有效」的接口：返回 200 即代表 token 合法。

## 错误码约定

| 场景 | HTTP | code | message 示例 |
| --- | --- | --- | --- |
| 入参 JSON 缺失/不符合 binding | 400 | 40000 | validator 自带描述 |
| 用户名或密码错误 | 401 | 40100 | `用户名或密码错误` |
| jwt 签发失败等内部错误 | 500 | 50000 | 透传 err.Error() |

## 修改建议

- 不要在控制器里直接操作 jwt 或读配置——这些都是 service 的职责。
- 新增受保护接口时，把它挂到 router 里 `protected` 分组下即可，无需在控制器里手动校验 token。
