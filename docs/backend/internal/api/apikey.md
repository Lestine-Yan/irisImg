# `internal/api/apikey.go`

API 密钥管理接口的 Gin 控制器。控制器层只做参数解析、调用 service、组装响应。这些接口均挂在需 **JWT 登录**的受保护组下，并要求 **HTTPS**（由 [`HTTPSOnly`](../middleware/https.md) 中间件保证）。成功路径上的密钥管理操作经 `rec` 记录业务事件到日志中心。

## 类型

### `APIKeyAPI`

- 字段：
  - `svc *service.APIKeyService`
  - `authSvc *service.AuthService` -- 用于吊销 / 删除前的密码二次确认。
  - `rec service.LogRecorder` -- 业务事件记录器，把密钥管理操作写入日志中心；为 `nil` 时静默跳过，便于测试。
- 由 [`router`](../router/router.md) 通过 `NewAPIKeyAPI(svc, authSvc, rec)` 注入。

## 处理函数

### `Create(c *gin.Context)` -- `POST /api/v1/apikeys`

1. `c.ShouldBindJSON(&req)` 解析 [`model.CreateAPIKeyRequest`](../model/apikey.md)；失败 -> `response.BadRequest`。
2. 调 `svc.Create`：
   - `errors.Is(err, service.ErrInvalidScope)` -> `response.BadRequest("scope 仅支持 readonly 或 readwrite")`
   - 其它 err -> `response.ServerError`
3. 成功后调 `record(c, model.EventAPIKeyCreate, model.LevelInfo, "create apikey: "+resp.Name)` 记一条 `apikey.create`（info）业务事件，再 `response.Success(c, resp)`，`resp` 含**一次性明文密钥** `key`。

### `List(c *gin.Context)` -- `GET /api/v1/apikeys`

调 `svc.List`，返回 `gin.H{"items": infos}`（[`APIKeyInfo`](../model/apikey.md) 列表，不含明文与哈希）。只读查询，不记录业务事件。

### `Rename(c *gin.Context)` -- `PATCH /api/v1/apikeys/:id`

1. 解析 ID 与 [`model.RenameAPIKeyRequest`](../model/apikey.md)。
2. 调 `svc.Rename`：`ErrKeyNotFound` -> 404；其它 err -> 500。
3. 成功后调 `record(c, model.EventAPIKeyRename, model.LevelInfo, "rename apikey #<id>")` 记 `apikey.rename`（info），再返回更新后的 `APIKeyInfo`。

### `Reset(c *gin.Context)` -- `POST /api/v1/apikeys/:id/reset`

调 `svc.Reset`，成功后调 `record(c, model.EventAPIKeyReset, model.LevelInfo, "reset apikey #<id>")` 记 `apikey.reset`（info），响应含**一次性新明文** `key`（重置同时取消吊销）。`ErrKeyNotFound` -> 404。

### `Revoke(c *gin.Context)` -- `POST /api/v1/apikeys/:id/revoke`

吊销密钥（软删除）。需在请求体携带 [`model.DestructiveAPIKeyRequest`](../model/apikey.md) 做密码二次确认：

1. 解析 ID 与 `DestructiveAPIKeyRequest`。
2. `authSvc.VerifyCredentials` 失败 -> `response.Forbidden("用户名或密码错误")`（**403 而非 401**，避免触发前端 `useApi` 的全局登出）。
3. 调 `svc.Revoke`：`ErrKeyNotFound` -> 404；其它 err -> 500。
4. 成功后调 `record(c, model.EventAPIKeyRevoke, model.LevelWarn, "revoke apikey #<id>")` 记 `apikey.revoke`（warn），返回 `gin.H{"id": id, "revoked": true}`。

### `Delete(c *gin.Context)` -- `DELETE /api/v1/apikeys/:id`

物理删除密钥并级联删除关联图片。同样需 `DestructiveAPIKeyRequest` 二次确认：

1. 解析 ID 与 `DestructiveAPIKeyRequest`。
2. `authSvc.VerifyCredentials` 失败 -> 403。
3. 调 `svc.Delete`：返回被删除的图片数量；`ErrKeyNotFound` -> 404；其它 err -> 500。
4. 成功后调 `record(c, model.EventAPIKeyDelete, model.LevelWarn, "delete apikey #<id>")` 记 `apikey.delete`（warn），返回 `gin.H{"id": id, "deleted": true, "images_removed": removed}`。

> 路由语义：`POST /:id/revoke` 为软吊销，`DELETE /:id` 为硬删除。旧的 `DELETE /:id`（软吊销）已被 `POST /:id/revoke` 取代。

### `record(c *gin.Context, event, level, msg string)` -- 辅助方法

密钥管理业务事件的统一记录入口。`rec` 为 `nil` 时直接返回（便于测试）；否则调 `rec.Record(model.NewEventLog(event, level, msg, middleware.LogContextFromGin(c)))`，把请求上下文（操作者 / 来源 IP / request_id 等）一并带入日志。`Create` / `Rename` / `Reset` 在成功路径上用 `LevelInfo` 记 `apikey.create` / `rename` / `reset`；`Revoke` / `Delete` 属敏感操作，用 `LevelWarn` 记 `apikey.revoke` / `delete`。`List` 为只读查询，不记录业务事件。仅在操作**成功**后触发，失败分支直接写响应返回。

## 错误码约定

| 场景 | HTTP | code |
|------|------|------|
| 入参非法 / scope 错 / ID 非数字 | 400 | 40000 |
| 密码二次确认失败（吊销 / 删除） | 403 | 40300 |
| 密钥不存在 | 404 | 40400 |
| 内部错误 | 500 | 50000 |

## 修改建议

- 不要在控制器里生成密钥或算哈希--那是 [`service.APIKeyService`](../service/apikey.md) 的职责。
- 新增管理接口挂到 router 的 `keys` 分组下即可，自动继承 JWT + HTTPS 保护。
- 敏感操作的密码校验放在控制器层（边界关注点），service 保持纯粹。
- 业务事件统一经 `record` + `rec` 写入日志中心（`apikey.create/rename/reset` 用 info，`apikey.revoke/delete` 用 warn）；`rec` 为 `nil` 时自动跳过，单测无需 mock。新增密钥类事件沿用 `model.EventAPIKey*` 命名并在此处调 `record`，不要散落 `fmt.Println`。
