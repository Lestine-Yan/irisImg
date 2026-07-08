# `internal/api/auth.go`

认证相关的 Gin 控制器。控制器层只做参数解析、调用 service、组装响应，不写业务规则。

## 类型

### `AuthAPI`

- 字段：
  - `authSvc *service.AuthService`
  - `rec service.LogRecorder` -- 业务事件记录器，用于把登录成功 / 失败写入日志中心。
- 由 [`router`](../router/router.md) 通过 `NewAuthAPI(authSvc, rec)` 注入；`rec` 可为 `nil`，单测中可省略。

## 处理函数

### `Login(c *gin.Context)` -- `POST /api/v1/auth/login`

1. `c.ShouldBindJSON(&req)` 解析 [`model.LoginRequest`](../model/auth.md)；失败 -> `response.BadRequest(err)`。
2. 调 `authSvc.Login(&req)`：
   - `errors.Is(err, service.ErrInvalidCredentials)` -> 调 `recordLogin(c, req.Username, false)` 记一次失败事件，再 `response.Unauthorized("用户名或密码错误")`。
   - 其它非 nil err -> `response.ServerError(err)`。
3. 成功时调 `recordLogin(c, req.Username, true)` 记一次成功事件，`response.Success(c, resp)`，`resp` 是 `*model.LoginResponse`，含 `token / token_type / expires_at`。

### `recordLogin(c, attemptedUsername, success)`（辅助）

- 从 `middleware.LogContextFromGin(c)` 取出请求上下文，把请求体里填写的 `attemptedUsername` 写入 `lc.Username`（即便鉴权未通过也记录，供审计），再调 `rec.Record` 写入一条 `model.EventLog`。
- 成功 -> `model.EventAuthLoginOK` / `model.LevelInfo`，事件 `auth.login_success`。
- 失败 -> `model.EventAuthLoginFail` / `model.LevelWarn`，事件 `auth.login_failed`。
- `rec == nil` 时直接返回，便于测试时跳过日志记录。

### `Me(c *gin.Context)` -- `GET /api/v1/auth/me`

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

- 不要在控制器里直接操作 jwt 或读配置--这些都是 service 的职责。
- 新增受保护接口时，把它挂到 router 里 `protected` 分组下即可，无需在控制器里手动校验 token。
- 需要审计的鉴权类事件统一走 `recordLogin` + `rec`，不要散落 `fmt.Println`；若新增其它登录态变更事件（登出、改密、token 刷新等），沿用 `model.EventAuth*` 命名并在此处调 `rec.Record`。
