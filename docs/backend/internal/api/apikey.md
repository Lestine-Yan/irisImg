# `internal/api/apikey.go`

API 密钥管理接口的 Gin 控制器。控制器层只做参数解析、调用 service、组装响应。这些接口均挂在需 **JWT 登录**的受保护组下，并要求 **HTTPS**（由 [`HTTPSOnly`](../middleware/https.md) 中间件保证）。

## 类型

### `APIKeyAPI`

- 字段：`svc *service.APIKeyService`
- 由 [`router`](../router/router.md) 通过 `NewAPIKeyAPI(svc)` 注入。

## 处理函数

### `Create(c *gin.Context)` —— `POST /api/v1/apikeys`

1. `c.ShouldBindJSON(&req)` 解析 [`model.CreateAPIKeyRequest`](../model/apikey.md)；失败 → `response.BadRequest`。
2. 调 `svc.Create`：
   - `errors.Is(err, service.ErrInvalidScope)` → `response.BadRequest("scope 仅支持 readonly 或 readwrite")`
   - 其它 err → `response.ServerError`
3. 成功 `response.Success(c, resp)`，`resp` 含**一次性明文密钥** `key`。

### `List(c *gin.Context)` —— `GET /api/v1/apikeys`

调 `svc.List`，返回 `gin.H{"items": infos}`（[`APIKeyInfo`](../model/apikey.md) 列表，不含明文与哈希）。

### `Revoke(c *gin.Context)` —— `DELETE /api/v1/apikeys/:id`

1. `strconv.Atoi(c.Param("id"))` 解析 ID；失败 → `response.BadRequest("无效的密钥 ID")`。
2. 调 `svc.Revoke`：
   - `errors.Is(err, dao.ErrNotFound)` → `response.NotFound("密钥不存在")`
   - 其它 err → `response.ServerError`
3. 成功返回 `gin.H{"id": id, "revoked": true}`。

## 错误码约定

| 场景 | HTTP | code |
|------|------|------|
| 入参非法 / scope 错 / ID 非数字 | 400 | 40000 |
| 吊销时密钥不存在 | 404 | 40400 |
| 内部错误 | 500 | 50000 |

## 修改建议

- 不要在控制器里生成密钥或算哈希——那是 [`service.APIKeyService`](../service/apikey.md) 的职责。
- 新增管理接口挂到 router 的 `keys` 分组下即可，自动继承 JWT + HTTPS 保护。
