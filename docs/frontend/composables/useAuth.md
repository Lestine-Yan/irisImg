# `frontend/app/composables/useAuth.ts`

维护前端登录态的组合函数。

## 职责

- 通过 `useState` 在运行时持有 token、过期时间、用户信息。
- 提供 `login()`、`logout()`、`fetchMe()`、`initAuth()`。
- 在客户端将 token 持久化到 `localStorage`。

## 关键类型

- `LoginRequest`：`{ username, password }`。
- `LoginResponse`：`{ token, token_type, expires_at }`。
- `UserProfile`：`{ username }`。

## 关键函数

- `login(username, password)`：调用 `/auth/login`，成功后保存 token 并拉取用户信息。
- `logout()`：清空内存状态与 `localStorage`。
- `fetchMe()`：调用 `/api/v1/auth/me` 获取当前用户。
- `initAuth()`：客户端启动时从 `localStorage` 恢复 token（仅客户端）。

## 与其它文件的关系

```
useAuth.ts
  ├── useApi.ts        # 发起 /auth/login 与 /auth/me 请求
  ├── localStorage     # 持久化 token / expires_at
  └── auth.client.ts   # 启动时调用 initAuth()
```

## 修改建议

- `localStorage` key（`irisimg_token`、`irisimg_expires_at`）修改时需同步更新本文件与 `auth.client.ts`。
- 若后续使用 Pinia 或 Cookie，可在此集中替换存储策略。
