# `frontend/app/pages/content/audio.vue`

**音频中心**，对应路由 `/content/audio`，管理已上传的音频文件（MP3 / M4A / FLAC）。页面状态机与 [`img.vue`](./img.md) 完全对称（按 API Key 筛选、时间升序分页、多选批量删除、上传面板互斥），差异在**列表形态**（纵向堆叠的横向条卡）与**详情弹窗**（全量元数据 + 封面 URL 编辑）。

## 职责

- 顶部展示「音频」标题 + 副标题「管理已上传的音频文件（MP3 / M4A / FLAC）」，按钮组与图片页同款：
  - **默认态**：「多选」+「上传音频」（切换 [`ContentAudioUploadPanel`](../../components/content/audio-upload-panel.md)）。
  - **多选态**：「删除（N）」+「取消」。
- 上传栏：拖拽 / 点击选择 mp3/m4a/flac，走 JWT 后台直传（`POST /admin/audios`），不关联密钥；后端按内容魔数校验，扩展名伪造无效。
- 左侧按钮栏：与图片页完全同款（「全部」+ 各 Key；切换重置第 1 页并退出多选）。
- 右侧音频列表：**纵向堆叠**的 [`ContentAudioCard`](../../components/content/audio-card.md) 横向条卡（`space-y-3`），每页 24 个，分页复用 [`UiPagination`](../../components/ui/Pagination.md)；加载 / 空 / 错误态与重试同款。
- 点击卡片：默认模式弹出 [`ContentAudioDetailDialog`](../../components/content/audio-detail-dialog.md)（全量元数据 + 封面 URL 修改）；多选模式切换选中。卡片封面上的**播放 / 暂停按钮在组件内 `@click.stop`**，不触发本页的点击分叉。
- 批量删除：多选态点「删除」弹出 [`ContentAudioBatchDeleteDialog`](../../components/content/audio-batch-delete-dialog.md)（账号密码二次确认），成功后按「实际删除」数收敛页码并刷新（同图片页逻辑）。

## 鉴权逻辑

- `definePageMeta({ middleware: 'auth' })`；列表 / 上传 / 批删 / 改封面均走 JWT 通道（`/admin/audios*`），由 [`useApi`](../../composables/useApi.md) 自动附带 `Authorization`。

## 数据流

- `onMounted` 并行拉取 `GET /apikeys` 与 `GET /admin/audios`（经 [`useAudios`](../../composables/useAudios.md)，默认 `order=asc`、`page=1`、`page_size=24`）。
- `keyNameById` 映射解析来源 Key 名称；`key_id` 为空显示 `admin`。
- 上传成功 `onUploaded`：admin 直传只在不筛选时可见——若在按密钥筛选则 `selectKey(null)`，否则 `fetchAudios()` 刷新当前页。
- **改封面成功 `onAudioUpdated(updated)`**：`findIndex` 就地替换列表项并同步 `selectedAudio`（弹窗保持打开、数据即时刷新），无需整页重新拉取——这是与图片页唯一的增量数据流。

## 多选模式状态机

与 [`img.vue`](./img.md) 逐条相同（`multiSelect` / `selectedIds` 跨页累计 / `showDeleteConfirm`；翻页保留选中、切换 Key 退出多选、删除后保留多选模式并收敛页码；状态行显示「· 已选 N 个」）。

## 与其它文件的关系

```
content/audio.vue
  ├── useAudios.ts                                     → GET/POST/DELETE /admin/audios、PATCH /admin/audios/:id（updateCover）
  ├── useApi.ts                                        → GET /apikeys
  ├── components/content/AudioUploadPanel.vue          （上传栏，emit uploaded）
  ├── components/content/AudioCard.vue                （横向条卡，传入 selectable / selected；内部 useAudioPlayer）
  ├── components/content/AudioDetailDialog.vue         （详情弹窗，emit updated）
  └── components/content/AudioBatchDeleteDialog.vue    （批量删除确认弹窗，emit done）
layouts/default.vue（承载侧边栏与用户区）
```
