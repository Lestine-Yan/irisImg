# `frontend/app/components/content/AudioDetailDialog.vue`

音频详情弹窗，Nuxt 自动导入标签为 `<ContentAudioDetailDialog />`。Teleport 双栏布局沿用 [`ImageDetailDialog.vue`](./image-detail-dialog.md)（遮罩 + ESC 关闭 + watch 清理），**右栏额外携带封面 URL 编辑表单**（输入样式与校验模式取自 [`RenameKeyDialog`](../apikeys/RenameKeyDialog.md)）。

## 布局

- **左栏（2/5，深色底）**：生效封面大图（`effectiveCover` = `cover_url || embedded_cover_url`，经 `resolveAudioUrl` 解析）；无封面显示居中音符 SVG 占位。
- **右栏（3/5，可滚动）**：
  - 标题（`title || filename`）+ 关闭按钮；
  - 全量元数据 `dl`（`rows` computed）：来源 Key / 标题 / 艺术家 / 专辑 / 专辑艺术家 / 作曲 / 流派 / 年份·音轨 / 碟号 / 时长 / 码率 / 采样率 / 声道（`channelLabel`：单声道 / 双声道 / N 声道）/ 位深 / 格式·大小 / MIME / 上传时间 / 文件 ID / 哈希 / 访问 URL / 内嵌封面——零值统一展示 `—`；
  - **封面 URL 表单**：输入框（预填**当前覆盖值** `cover_url ?? ''`，留空即恢复内嵌封面）、轻校验（非空须以 `http(s)://` 或 `/` 开头且 ≤ 2048 字符，与后端 binding 对齐）、内联错误与 `serverError` 横幅、底部三按钮：取消 / 恢复默认 / 保存。

## 封面编辑语义

- **输入框值是「用户覆盖」而非「生效封面」**：留空保存 = 清除覆盖（恢复展示内嵌封面），与后端「空串清除」语义一致；「恢复默认」按钮等价于清空输入框后提交。
- `handleSave`（提交输入值）/ `handleReset`（提交空串）都走 `useAudios().updateCover(id, coverUrl)` → `PATCH /admin/audios/:id`。
- 成功 `emit('updated', audio)`，父页面就地替换列表项（无需整页刷新）；失败内联报错可重试（非 401 场景不触发全局登出）。

## Props / Emits

- props：`audio: AudioItem | null`（null 时整个弹窗不渲染）；`keyName?: string`（来源 Key 名称，admin 直传时为 `'admin'`）。
- emits：`close: []`；`updated: [audio: AudioItem]`。

## 与其它文件的关系

- 父组件：[`pages/content/audio.vue`](../../pages/content/audio.md)（`onAudioUpdated` 就地替换列表项并同步弹窗内数据）。
- 依赖 [`useAudios`](../../composables/useAudios.md)（`updateCover` / `effectiveCover` / `formatDuration` / `resolveAudioUrl`）、[`useImages`](../../composables/useImages.md)（`formatBytes` / `formatDate`）。
