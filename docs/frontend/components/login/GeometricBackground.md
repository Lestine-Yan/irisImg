# `frontend/app/components/login/GeometricBackground.vue`

登录页全屏 SVG 几何背景组件。

## 职责

- 使用固定定位覆盖整个视口。
- 通过 SVG 多边形、线性渐变与图案（pattern）还原「斜四边形拼接 + 渐变」的鸢尾花风格背景。

## 视觉元素

- 三块斜四边形（平行斜边）拼接，从左到右：
  - 左块：暖米白留白（`creamFade`，`#FAFBE6` → `#F5F8DA`，`iris-cream`），作为左侧文案底色。
  - 中块：浅天蓝 → 紫罗兰渐变（`skyViolet`，`#86C9EC` → `#6D4FD8`）。
  - 右块：紫罗兰 → 浅天蓝渐变（`violetSky`，`#6D4FD8` → `#86C9EC`）。
- 斜 45° 菱形网格（`grid45`，`#B9A6E6`）**仅在留白区域**叠加；右块内另有一小块暖米白留白（`#FAFBE6`）同样带网格作为点缀。
- 金黄色斑纹（`#F4C430`）：留白区右缘一条主斜斑纹，以及中块、右块各一条短斜斑纹，呼应鸢尾花须状斑纹。
- 外层容器底色为 `bg-iris-cream`，与留白区一致。

## 与其它文件的关系

- 被 `frontend/app/pages/index.vue` 引用，作为登录页最底层视觉层。
- 配色与 `tailwind.config.ts` 的鸢尾色系保持一致；留白底色对应 `iris-cream`。

## 修改建议

- 颜色值与多边形坐标直接写死在 SVG 中，若需复用到其他页面，建议将色值抽取到 Tailwind 主题或 CSS 变量。
- 左侧留白区为文字可读性服务；调整其右缘斜边时需同步确认 `LoginHero.vue` 字标不越界进入渐变块。
- `aria-hidden="true"` 避免装饰性 SVG 干扰屏幕阅读器。
