# `frontend/app/pages/content/index.vue`

内容中心占位页，对应路由 `/content`。

## 职责

- 后台「内容中心」导航目标的占位页面。
- 暂只展示标题、副标题与「正在开发中」占位卡片。

## 鉴权逻辑

- `definePageMeta({ middleware: 'auth' })`：未登录跳转 `/`。

## 与其它文件的关系

```
content/index.vue
  └── middleware/auth.ts → useAuth.ts
layouts/default.vue（承载侧边栏与用户区）
```

## 修改建议

- 后续替换为图片资源列表 / 网格视图，接入 `GET /images` 等接口。
