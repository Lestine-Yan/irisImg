# `frontend/app/pages/apikeys/index.vue`

APIkey 管理占位页，对应路由 `/apikeys`。

## 职责

- 后台「APIkey 管理」导航目标的占位页面。
- 暂只展示标题、副标题与「正在开发中」占位卡片。

## 鉴权逻辑

- `definePageMeta({ middleware: 'auth' })`：未登录跳转 `/`。

## 与其它文件的关系

```
apikeys/index.vue
  └── middleware/auth.ts → useAuth.ts
layouts/default.vue（承载侧边栏与用户区）
```

## 修改建议

- 后续接入 APIkey 增删改查接口，呈现密钥列表、创建表单与权限范围选择。
