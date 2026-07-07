# `frontend/app/pages/apikeys/index.vue`

APIkey 管理页，对应路由 `/apikeys`，后台「APIkey 管理」导航目标。

## 职责

- 顶栏：标题 + 「创建 Key」按钮（`justify-between` 布局）。
- 标题下说明文本：APIkey 仅在创建时可见可复制，请妥善保存，不要与他人共享或暴露在客户端代码中。
- 密钥列表（[`ApiKeyTable`](../../components/apikeys/ApiKeyTable.md)）：操作 / 名称 / 明文前缀 / 创建时间 / 最近使用。
- 弹窗编排：创建 / 明文展示（创建与重置共用）/ 重命名 / 重置 / 吊销·删除。同一时刻只开一个操作弹窗。

## 鉴权逻辑

- `definePageMeta({ middleware: 'auth' })`：未登录跳转 `/`。

## 数据流

- `onMounted` 调 [`useApiKeys`](../../composables/useApiKeys.md).`list()` 拉取列表。
- 创建 / 重命名 / 重置 / 吊销 / 删除 成功后均刷新列表；创建与重置成功后弹出明文展示弹窗（一次性）。

## 与其它文件的关系

```
pages/apikeys/index.vue
  ├── composables/useApiKeys.ts ── useApi.ts
  ├── components/apikeys/ApiKeyTable.vue
  ├── components/apikeys/CreateKeyDialog.vue
  ├── components/apikeys/PlaintextKeyDialog.vue
  ├── components/apikeys/RenameKeyDialog.vue
  ├── components/apikeys/ResetKeyDialog.vue
  └── components/apikeys/RevokeDeleteDialog.vue
layouts/default.vue（承载侧边栏与用户区）
```
