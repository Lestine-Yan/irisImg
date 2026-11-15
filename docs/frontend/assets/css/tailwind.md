# `frontend/app/assets/css/tailwind.css`

Tailwind CSS 指令入口与全局基础样式文件。

## 职责

- 引入 Tailwind 三个核心指令：`@tailwind base` / `components` / `utilities`。
- 定义全局页面底色、正文色与选区色。
- 定义后台复用的玻璃表面与页面标题类：
  - `.admin-surface`：20px 圆角、白色 85% 细边、白色 74% 背景、低对比多层阴影、18px backdrop blur。用于仪表盘卡片、表格容器、配置分组、图片/音频卡片等。
  - `.admin-surface-flat`：与 `.admin-surface` 同款圆角、细边与阴影，但不带 backdrop blur；用于图片 / 音频网格这类大量重复卡片，避免滚动时 GPU 压力过大。
  - `.admin-surface-quiet`：虚线边 + 白色 45% 背景，用于加载 / 空态占位。
  - `.admin-page-title`：26px、600 字重、`-0.02em` 字距、近黑墨色。
  - `.admin-page-subtitle`：13px、暖灰正文色。

## 与其它文件的关系

- 被 `frontend/nuxt.config.ts` 的 `css` 配置项引用，由 Nuxt 构建时注入。
- `.admin-surface` / `.admin-page-title` 等类被后台页面与公共组件使用。

## 修改建议

- 新的通用卡片或标题样式优先扩展这里的组件类，避免每个页面重复长串 utility。
- 避免在此处写业务选择器；页面特有的布局仍放在页面组件内。
