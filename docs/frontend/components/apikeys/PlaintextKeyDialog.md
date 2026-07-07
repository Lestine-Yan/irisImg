# `frontend/app/components/apikeys/PlaintextKeyDialog.vue`

一次性明文密钥展示弹窗（创建 / 重置共用），Nuxt 自动导入标签为 `<ApikeysPlaintextKeyDialog />`。

## 职责

- 展示明文密钥（只读输入框）+ 复制按钮 + 「仅此一次」警示。
- 复制用 `navigator.clipboard.writeText`，成功后按钮变 emerald「已复制」2 秒；clipboard 不可用时兜底选中文本。

## Props / Emits

- props：`open: boolean`、`plaintext: string`、`title?: string`。
- emits：`close()`。

## 实现要点

- `watch(open)` 重置「已复制」状态。
- 输入框 `@focus` 自动全选，便于手动复制。

## 与其它文件的关系

- 父组件：[`pages/apikeys/index.vue`](../../pages/apikeys/index.vue)（创建成功 / 重置成功后复用）。
- 外壳：[`ui/BaseDialog`](../ui/BaseDialog.md)。
