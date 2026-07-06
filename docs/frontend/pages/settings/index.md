# `frontend/app/pages/settings/index.vue`

系统配置占位页，对应路由 `/settings`。

## 职责

- 后台「系统配置」导航目标的占位页面。
- 暂只展示标题、副标题与「正在开发中」占位卡片。

## 鉴权逻辑

- `definePageMeta({ middleware: 'auth' })`：未登录跳转 `/`。

## 与其它文件的关系

```
settings/index.vue
  └── middleware/auth.ts → useAuth.ts
layouts/default.vue（承载侧边栏与用户区）
```

## 修改建议

- 后续接入站点参数、存储后端、CORS 等配置项的读写接口，呈现表单 + 保存。
