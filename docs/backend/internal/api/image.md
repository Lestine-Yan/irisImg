# `internal/api/image.go`

图片相关接口的 Gin 控制器。

- **`POST /api/v1/images`** —— 添加图片，**已实现**，由 [API 密钥鉴权中间件](../middleware/apikey.md) 保护，需 `readwrite` 密钥。
- **`GET  /api/v1/images`** —— 任意有效密钥可访问，**目前为占位**，待接入前端列表页时再实现，显式返回 501 与鉴权失败状态码区分开。

## 类型

### `ImageAPI`

```go
type ImageAPI struct {
    svc *service.ImageService
}
```

由 [`router`](../router/router.md) 通过 `NewImageAPI(svc)` 注入；不再直接持有 DAO，便于在 service 层统一编排上传链路。

## 处理函数

### `Create(c *gin.Context)` —— `POST /api/v1/images`

**请求**：`multipart/form-data`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `file` | file | 图片文件，必填。MIME 白名单见 `storage.allowed_mime_types` |

**Header**：`X-API-Key: <readwrite 密钥>`（由 [`middleware.APIKeyAuth`](../middleware/apikey.md) 校验）。

**响应**：200 + 完整 `model.Image`（含 `id / url / stored_path / size / mime_type / width / height / hash / key_id / created_at` 等）。

**错误映射**：

| HTTP | 业务码 | 触发 |
| --- | --- | --- |
| 400 | `CodeBadRequest` | 缺少 `file` 字段、上传内容为空、MIME 不在白名单 |
| 401 / 403 / 429 | 见 [`APIKEY.md`](../../APIKEY.md) | 鉴权失败 / 权限不足 / 限流 |
| 413 | `CodePayloadTooLarge` | 文件超过 `storage.max_upload_size_mb` |
| 500 | `CodeServerError` | 内部错误（落盘 / 落库失败等） |

**关键实现**：

1. `http.MaxBytesReader(...)` 把 `c.Request.Body` 包一层，超大请求体在更早阶段就被拦下。上限取自 `svc.MaxBytes()`。
2. `c.FormFile("file")` 解析出 `*multipart.FileHeader`；若被 `MaxBytesReader` 触发，`errors.As(err, *http.MaxBytesError)` 命中 → 413。
3. 读完文件全部字节后调 `svc.Upload`，按 sentinel error 分支映射状态码。
4. `key_id` 取自 `c.GetInt(middleware.ContextKeyAPIKeyID)`（中间件保证一定 >0），透传给 service 落库，记录图片由哪把密钥添加。

### `List(c *gin.Context)` —— `GET /api/v1/images`（占位）

显式返回 `501 Not Implemented`（`code = CodeServerError`，message「图片列表接口尚未实现（占位）」），便于前端 / 调用方与鉴权失败的 401/403/429 区分。

## 与其它包的关系

```
images 组 ── APIKeyAuth ──► ImageAPI.Create ──► ImageService.Upload ──► dao.ImageDAO + pkg/storage.Saver
                                              │
                                              └─ c.GetInt(ContextKeyAPIKeyID) → model.UploadImageInput.KeyID
```

## 注意

- 控制器**不再做权限/限流判定**，这些全部在 [`middleware.APIKeyAuth`](../middleware/apikey.md) 内完成。
- `MaxBytesReader` 的兜底校验在 service 层用 `len(content) > MaxBytes` 又做一次，避免接入新调用路径时遗漏。
- 落盘 / URL 生成的细节见 [`service/image.md`](../service/image.md) 与 [`pkg/storage.md`](../pkg/storage.md)；运维侧的 Nginx 静态反代约定见 [`IMAGE.md`](../../IMAGE.md)。
