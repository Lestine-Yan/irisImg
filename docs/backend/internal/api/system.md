# `internal/api/system.go`

系统配置只读接口的 Gin 控制器。控制器层只做调用 service 与组装响应，**无参数解析、无写操作、无业务事件记录**。该接口挂在需 **JWT 登录**的受保护组下（见 [`router`](../router/router.md)），直接注册于 `protected` 而非 `HTTPSOnly` 子组（与密钥管理 / 日志清理等敏感接口不同），故开发环境 HTTP 亦可访问。

## 类型

### `SystemAPI`

- 字段：`svc *service.SystemService`。
- 由 [`router`](../router/router.md) 通过 `NewSystemAPI(svc)` 注入。
- 不持有 `LogRecorder`：系统配置查询属运维查看、非敏感操作，不记业务事件，因此构造时不注入 `logSvc`（与 `AuthAPI` / `APIKeyAPI` / `ImageAPI` / `LogAPI` 不同）。

## 处理函数

### `Config(c *gin.Context)` -- `GET /system/config`

返回当前系统配置的只读视图。流程：

1. 无入参（不解析 query / body）。
2. 调 `h.svc.Config()` 拿到 [`model.SystemConfigResponse`](../model/system.md)。
3. `response.Success(c, h.svc.Config())` 返回 200 + 配置快照。

由于 [`service.SystemService.Config()`](../service/system.md) 是纯内存映射、不返回 error，handler 内无错误分支，整段实现即一行透传。

## 错误码约定

| 场景 | HTTP | code |
|------|------|------|
| 未携带 / 无效 / 过期 JWT | 401 | 40100（由 [`middleware.JWTAuth`](../middleware/auth.md) 中间件统一返回，非本 handler 处理） |
| 成功 | 200 | 0 |

本 handler 自身不产生 4xx / 5xx；唯一可能的失败是 JWT 鉴权未通过，由 `JWTAuth` 中间件在进入 handler 前拦截并 `c.Abort()`。

## 修改建议

- 不要在控制器里读取 config 或拼装 DTO--那是 [`service.SystemService`](../service/system.md) 的职责；控制器保持一行透传。
- 如需返回更细粒度的配置（如 `logger` 段），在 service 与 model 同步新增字段，控制器无须改动。
- 若后续把系统配置升级为可写（需慎重），应另起 handler 并加密码二次确认 + 审计事件，不要复用本只读 handler。
