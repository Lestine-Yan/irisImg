# `frontend/app/pages/index.vue`

登录落地页，对应路由 `/`。

## 职责

- 组合登录页的三个核心视觉组件：
  - `LoginGeometricBackground`：全屏留白背景、细线 SVG 与鸢尾花图层。
  - `LoginHero`：左侧品牌字标、产品定位文案与能力说明。
  - `LoginForm`：右侧半透明登录卡片。
- 提供顶部的品牌条 / 自托管状态，以及底部版权信息。
- 登录成功后先触发 `useAppTransition().begin()`，显示全屏加载层；等待约 520ms 后跳转 `/dashboard`，再等待约 320ms 调用 `reveal()`，由加载层完成曲线迁移并自行结束。

## 布局与层级

- 使用 `definePageMeta({ layout: false })` 禁用默认布局，让背景铺满整个视口。
- 页面最外层为 `relative min-h-screen overflow-hidden`，**不设置页面背景色**；背景由 `LoginGeometricBackground` 以 `fixed inset-0 z-0` 承担。若在页面根节点再加不透明背景，会盖住 SVG 与花朵图层。
- 内容层使用 `relative z-10 flex min-h-screen flex-col`，依次为：
  - `header`：左侧 `irisImg` 花形标记与 `private cloud`，右侧 `self-hosted` 状态胶囊。
  - `main`：左侧 Hero、右侧登录卡片。
  - `footer`：版权与「为个人而建，由你所有」。
- 桌面端（`lg`）为两栏网格 `1.08fr / 0.92fr`，垂直居中对齐；移动端上下堆叠。
- 移动端隐藏 Hero 的能力说明列表，压缩字标与间距，让登录表单尽量在首屏内出现。

## 响应式

- 移动端：品牌条 → Hero（字标、标题、描述）→ 登录卡片，单列居中；卡片宽度受 `max-w-[420px]` 约束。
- 桌面端：左右分栏，卡片右对齐；左下能力说明恢复为三列。
- 页面在 390px 视口下 `scrollWidth` 与视口一致，无横向溢出。

## 与其它文件的关系

```
index.vue
  ├── LoginGeometricBackground.vue   // 背景层（z-0）
  ├── LoginHero.vue                  // 左侧文案
  └── LoginForm.vue                  // 右侧表单
        └── 登录成功 emit('success')
              └── begin() → 520ms → navigateTo('/dashboard')
                    → 320ms → reveal() → AppLoadingTransition 曲线过渡
```

## 修改建议

- 调整背景层 `z-index` 或页面根节点背景时，需同时确认花朵与 SVG 未被遮挡。
- 过渡时长与 `AppLoadingTransition` 的动画时长耦合，修改等待时间时需同步复查曲线结束时机。
- 若后续登录页改为独立路由 `/login`，需要同步调整此处路由与 `dashboard.vue` 的跳转目标。
- 移动端首屏高度依赖 Hero 的间距与字标宽度，改动时建议在 390×844 视口复查登录按钮是否仍可尽快触达。
