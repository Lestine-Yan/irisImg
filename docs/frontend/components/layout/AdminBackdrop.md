# `frontend/app/components/layout/AdminBackdrop.vue`

后台全屏背景层，Nuxt 自动导入标签为 `<LayoutAdminBackdrop />`。

## 职责

- 以 `fixed inset-0 z-0` 铺满视口，作为后台布局与加载过渡层的共同背景。
- 延续登录页的暖白纸色、细网格、柔和径向光晕与轨道圆。
- 渲染 `utils/transitionCurves.ts` 中四条曲线的 `settled` 落点，作为加载动画结束后的静态背景。
- 可通过 `showCurves` 控制是否显示静态曲线（加载过渡层传 `false`，避免与漂浮曲线重叠）。

## 视觉元素

- 底色 `#F7F7F2`；右上浅天蓝、下方紫罗兰、左上少量金色光晕。
- 80×80 的 1px 网格，径向遮罩只保留中部。
- 以 `(1120, 420)` 为圆心的四层低透明度轨道圆。
- 四条落定曲线：左下长弧、底部扫弧、上方进入右侧、右上边缘弧。
- 稀疏紫点、两个十字标记，作为画面细节。
- 底部 160px 渐隐，避免内容区底部出现生硬边界。

## 与加载过渡的关系

- 加载层使用同一组件作为背景（`show-curves=false`），所以背景底色、网格、轨道线完全一致。
- 加载层曲线动画到 `settled` 落点后淡出，后台背景中的同一组曲线随即显现，形成「曲线迁移完成页面切换」的观感。

## 与其它文件的关系

- 被 `layouts/default.vue` 作为后台底层背景。
- 被 `components/AppLoadingTransition.vue` 复用为加载层背景。
- 曲线数据来自 `utils/transitionCurves.ts`。

## 修改建议

- 后台内容使用半透明白色卡片，背景元素需保持低透明度，优先保证正文对比度。
- 若新增曲线或改变落点，必须同步 `transitionCurves.ts`，否则加载层与后台会错位。
- 该组件 `pointer-events-none` + `aria-hidden`，不参与交互和读屏。
