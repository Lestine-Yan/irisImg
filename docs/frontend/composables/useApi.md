# `frontend/app/composables/useApi.ts`

封装 `$fetch` 的通用 API 客户端组合函数。

## 职责

- 创建基于 `apiBase` 的 `$fetch` 实例。
- 自动附加 `Authorization: Bearer <token>` 请求头。
- 统一解析后端响应体 `{ code, message, data }`。
- 非 0 code 抛出 `ApiError`。
- HTTP 401 时自动登出并跳转 `/`。

## 关键类型

- `ApiResponse<T>`：后端统一响应结构。
- `ApiError`：继承 `Error`，携带 `code` 字段。

## 关键函数

- `api`：原始 `$fetch` 实例。
- `get<T>(url, opts)`：发起 GET 请求。
- `post<T>(url, body, opts)`：发起 POST 请求。

## 与其它文件的关系

```
useApi.ts
  ├── useRuntimeConfig().public.apiBase
  └── useState('auth-token')  ← 由 useAuth.ts 维护
```

## 修改建议

- 新增 PUT/DELETE/PATCH 方法时，按 `get`/`post` 模式扩展即可。
- 若需要更复杂的错误上报或重试逻辑，可在 `onResponseError` 中统一处理。
- 401 跳转逻辑使用 `navigateTo('/')`，后续如需跳转到登录页 `/login` 需同步修改。
