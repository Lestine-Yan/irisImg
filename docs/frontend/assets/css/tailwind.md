# `frontend/app/assets/css/tailwind.css`

Tailwind CSS 指令入口文件。

## 职责

- 引入 Tailwind 的三个核心指令：
  - `@tailwind base`：重置与基础样式。
  - `@tailwind components`：组件类（当前未自定义）。
  - `@tailwind utilities`：工具类。

## 与其它文件的关系

- 被 `frontend/nuxt.config.ts` 的 `css` 配置项引用，由 Nuxt 在构建时自动注入。

## 修改建议

- 若需要自定义全局样式（如 `html`、`body`、滚动条），可在此文件底部追加。
- 避免在此处写入大量组件样式，优先使用 Tailwind utility class 或 scoped CSS。
