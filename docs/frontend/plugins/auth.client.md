# `frontend/app/plugins/auth.client.ts`

客户端认证恢复插件。

## 职责

- 应用仅在客户端启动时执行。
- 调用 `useAuth().initAuth()`，从 `localStorage` 恢复已保存的 token 与过期时间。

## 命名约定

- 文件名后缀 `.client.ts` 确保该插件只在浏览器端运行，避免 SSR 阶段访问 `localStorage` 报错。

## 与其它文件的关系

```
auth.client.ts
  └── useAuth.ts
        └── initAuth() 读取 localStorage
```

## 修改建议

- 若后续改用 Cookie 或服务器端会话，可删除此插件或改为从 Cookie 读取。
- 恢复 token 后若需要同步拉取用户信息，当前已在 `initAuth()` 内部调用 `fetchMe()`。
