# `frontend/app/components/layout/UserBadge.vue`

侧边栏底部当前用户组件，Nuxt 自动导入名为 `<LayoutUserBadge>`。

## 职责

- 展示用户头像：浅紫渐变圆 + 用户名首字符。
- 完整侧边栏（`>= xl`）下展示用户名与「已登录 · 私人空间」。
- 提供退出登录按钮；窄屏下面板纵向排列，仍保留退出入口。

## 状态

- `initial`：由 `user.username` 首字符大写计算得到；未登录时显示 `?`。

## 视觉

- 容器：窄屏 `flex-col` 居中；`>= xl` 转为横向玻璃卡片（`border-white/70 bg-white/45`）。
- 头像：`h-9 w-9` 圆形，`bg-gradient-to-br from-white to-[#F0EBFC]`，`text-iris-dark` + 白色细边与轻阴影。
- 用户名：`12.5px font-semibold text-[#1C1A27]`；副文案：`10px text-[#8A8696]`。
- 退出按钮：低对比灰色，hover 转白底与深色；带 `title` / `aria-label`。

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
