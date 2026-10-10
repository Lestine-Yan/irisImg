# `internal/api/audio.go`

> 文档路径约定：`internal/api/audio.go` → `docs/backend/internal/api/audio.md`。

音频相关接口的控制器，与 [image.go](image.md) 平行。完整契约（含请求/响应示例与错误表）见 `docs/backend/AUDIO.md`。

## 类型

- `AudioAPI{ svc *service.AudioService, authSvc *service.AuthService, rec service.LogRecorder }`；`NewAudioAPI(svc, authSvc, rec)`。
- `recordUpload`（`audio.upload`，LevelInfo）/ `recordDelete`（`audio.delete`，LevelWarn）/ `recordUpdateCover`（`audio.update_cover`，LevelInfo，clear 时消息为 `reset`）——审计事件常量定义在 [model/log.go](../model/log.md)。

## 端点

| 路由 | Handler | 鉴权 | 说明 |
| --- | --- | --- | --- |
| `GET /api/v1/audios` | `List` | API Key（任意有效） | **501 占位**（与 GET /images 对称，语义待定） |
| `POST /api/v1/audios` | `Create` | API Key（readwrite，含限流） | 对外上传；`KeyID` 取自 `ContextKeyAPIKeyID`（中间件保证 >0） |
| `GET /api/v1/admin/audios` | `ListAdmin` | JWT + HTTPSOnly | 列表；query `key_id`/`order`/`page`/`page_size`，响应 `{items, total, page, page_size}` |
| `POST /api/v1/admin/audios` | `CreateAdmin` | JWT + HTTPSOnly | 后台直传；`KeyID` 传 nil（admin） |
| `DELETE /api/v1/admin/audios` | `BatchDeleteAdmin` | JWT + HTTPSOnly | 批量删除；`VerifyCredentials` 失败 → **403**（不触发前端全局登出） |
| `PATCH /api/v1/admin/audios/:id` | `UpdateCoverAdmin` | JWT + HTTPSOnly | 改封面；body `{cover_url}`（≤2048，空串=清除覆盖）；`ErrAudioNotFound` → 404 |

## 复用

`readUploadFile`（MaxBytesReader + multipart "file" 字段 + 413 映射）、`uploadFormField`、`respondUploadErrorWith(c, err, "音频")`（与图片共用的参数化错误映射，只影响文案不影响状态码：空 400 / 超限 413 / 格式 400「不支持的音频类型」/ 其余 500）。

## 与其它文件的关系

- 路由注册：[`internal/router/router.go`](../router/router.md)（`/audios` APIKeyAuth 组 + `/admin/audios` JWT/HTTPSOnly 组）；
- 业务：[`internal/service/audio.go`](../service/audio.md)；
- DTO：[`internal/model/audio.md`](../model/audio.md)。
