# `frontend/app/components/login/LoginHero.vue`

登录页左侧品牌字标组件。

## 职责

- 以**内联 SVG 字标**展示品牌名称 `irisImg`，且仅展示该字标（无副标题、无功能标签、无装饰圆点）。
- 字标由 `iris` 与 `Img` 两部分**上下两行**布局构成，两行统一使用浅紫罗兰（`#A48CE6`，`iris-violet`）。

## 对外行为

- 纯展示组件，无 props、无 emits、无内部状态（仅模板）。

## 视觉与定位

- SVG `viewBox="0 0 240 200"`，两行 `font-size="96"`、`font-weight="800"`、`letter-spacing="-3"`，相较初版明显放大。
- 响应式宽度：`w-56 sm:w-64 md:w-[360px] lg:w-[420px]`，高度自适应；桌面端有效字号约 168px。
- 组件自身 `flex flex-col items-start` 左对齐；实际「左上方」定位由 `frontend/app/pages/index.vue` 通过 `lg:self-start lg:pt-24` 顶部对齐实现。
- 字标落在 `GeometricBackground.vue` 左侧暖米白留白区（`iris-cream` `#FAFBE6`）内，右缘不越过留白斜边。
- `iris-violet`（`#A48CE6`）于米白底上对比度偏低（约 2.7:1），系刻意取色以追求观感舒适，非可读性优先。

## 与其它文件的关系

- 被 `frontend/app/pages/index.vue` 放置在左侧分栏，桌面端顶部对齐。
- 字标配色与 `GeometricBackground.vue` 的鸢尾色系一致；可读性依赖背景左侧保持留白。

## 修改建议

- 字标内文字使用 `<text>` 渲染，依赖 Inter 字体已加载；若需脱离字体依赖可转为路径（path）。
- 调整字号或宽度时需同步确认下沿不越过 `GeometricBackground.vue` 留白的斜边（留白区上宽下窄）。
- 若后续需要更高可读性，可将字标色改回 `iris-dark`（`#6D4FD8`）。
