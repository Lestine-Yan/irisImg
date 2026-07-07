# `frontend/app/composables/useApiKeys.ts`

APIkey 管理页用到的密钥接口封装 + 类型定义，供 [`pages/apikeys/index.vue`](../pages/apikeys/index.md) 与 `components/apikeys/*` 使用。

## 导出类型

- `APIKeyScope`：`'readonly' | 'readwrite'`。
- `APIKeyInfo`：密钥列表项，对应后端 `model.APIKeyInfo`（`id / name / prefix / scope / rate_limit / revoked / last_used_at / created_at`，不含明文与哈希）。
- `CreateAPIKeyResponse` / `ResetAPIKeyResponse`：创建 / 重置响应，含**一次性明文** `key`。
- `DestructiveAPIKeyRequest`：吊销 / 删除的请求体（`username / password`，二次确认）。
- `DeleteAPIKeyResult`：删除响应，含 `images_removed`（级联删除的图片数）。

## 导出函数

### `useApiKeys()`

返回（均走后台 JWT 通道，由 [`useApi`](./useApi.md) 自动附带 Bearer、自动解包 `data`）：

- `list()`：`GET /apikeys` → `APIKeyInfo[]`。
- `create({name, scope})`：`POST /apikeys`，返回含一次性明文。
- `rename(id, name)`：`PATCH /apikeys/:id`。
- `reset(id)`：`POST /apikeys/:id/reset`，返回含一次性新明文。
- `revoke(id, creds)`：`POST /apikeys/:id/revoke`（需账号密码）。
- `purge(id, creds)`：`DELETE /apikeys/:id`（硬删 + 级联删图，需账号密码）。

PATCH / DELETE 走 `useApi` 的 `api`（`$fetch` 实例），因为导出的 `get` / `post` 不支持这两个方法。

> 后端密码校验失败返回 **403**（非 401），不会触发 `useApi` 的全局登出，由调用方就地提示「用户名或密码错误」。

## 与其它文件的关系

- 被 [`pages/apikeys/index.vue`](../pages/apikeys/index.md) 与 `components/apikeys/*` 使用。
- 底层依赖 [`useApi.ts`](./useApi.md)。时间格式化复用 [`useImages.ts`](./useImages.md) 的 `formatDate`（自动导入）。
