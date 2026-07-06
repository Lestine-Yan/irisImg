# `frontend/app/components/layout/UserBadge.vue`

侧边栏底部当前用户组件，Nuxt 自动导入名为 `<LayoutUserBadge>`。

## 职责

- 展示用户头像（白底圆 + 浅紫首字符）。
- 展示当前用户名。
- 提供退出登录按钮。

## 状态

- `initial`：由 `user.username` 首字符大写计算得到；未登录时显示 `?`。

## 视觉

- 头像：`h-9 w-9` 圆形，`bg-white border border-gray-200`，文字 `text-iris-violet`（白底 + 浅色首字符，与浅白底侧边栏区分）。
- 退出按钮：`text-gray-400`，hover `bg-gray-100 text-gray-700`。

## 与其它文件的关系

```
UserBadge.vue
  └── useAuth.ts
        ├── user.username
        └── logout() → navigateTo('/')
```

## 修改建议

- 后续可补用户菜单下拉（修改密码、切换账号等）。
- 若引入头像上传，可替换首字符占位为真实头像 URL。
