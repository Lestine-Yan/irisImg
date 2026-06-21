# `internal/api/image.go`

图片相关接口的 Gin 控制器。

> **当前为占位实现**：真实的「申请图片(GET) / 添加图片(POST)」业务逻辑后续单独实现。这里保留占位处理函数，主要用于**挂载并演示 [API 密钥鉴权中间件](../middleware/apikey.md)**：只读密钥可访问 GET，读写密钥才能 POST，超出限流会被拒。

## 类型

### `ImageAPI`

- 字段：`imageDAO dao.ImageDAO`（占位阶段未真正读写，预留给后续实现）。
- 由 [`router`](../router/router.md) 通过 `NewImageAPI(imageDAO)` 注入。

## 处理函数

### `List(c *gin.Context)` —— `GET /api/v1/images`（占位）

- 受 [`middleware.APIKeyAuth`](../middleware/apikey.md) 保护；任意有效密钥（readonly / readwrite）均可访问。
- 当前从 `c.GetInt(middleware.ContextKeyAPIKeyID)` 取出密钥 ID 后，返回 `501 Not Implemented`（`code = CodeServerError`，message「图片列表接口尚未实现（占位）」）。

### `Create(c *gin.Context)` —— `POST /api/v1/images`（占位）

- 受中间件保护；非 GET 请求需 **readwrite** 密钥（否则中间件返回 403）。
- 当前返回 `501 Not Implemented`（占位）。
- **真实实现时**应将 `middleware.ContextKeyAPIKeyID` 写入 [`model.Image.KeyID`](../model/image.md) 落库，以记录图片由哪个密钥添加（对应 [`Image` 的 `key` edge](../../ent/schema/image.md)）。

## 与其它包的关系

```
images 组 ── APIKeyAuth ──► ImageAPI.List / Create ──► (c.GetInt "api_key_id")
```

## 注意

- 占位函数刻意返回 501，方便前端 / 调用方区分「鉴权通过但功能未上线」与真正的鉴权失败（401/403/429）。
- 鉴权链路与权限矩阵的完整说明见 [`APIKEY.md`](../../APIKEY.md)。
