# `frontend/app/pages/logs/index.vue`

日志中心占位页，对应路由 `/logs`。

## 职责

- 后台「日志中心」导航目标的占位页面。
- 暂只展示标题、副标题与「正在开发中」占位卡片。

## 鉴权逻辑

- `definePageMeta({ middleware: 'auth' })`：未登录跳转 `/`。

## 与其它文件的关系

```
logs/index.vue
  └── middleware/auth.ts → useAuth.ts
layouts/default.vue（承载侧边栏与用户区）
```

## 修改建议

- 后续接入上传 / 访问日志查询接口，呈现表格 + 筛选 + 分页。
