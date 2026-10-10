# `frontend/app/components/content/AudioCard.vue`

音频中心的单个**横向条卡**，Nuxt 自动导入标签为 `<ContentAudioCard />`。与 [`ImageCard.vue`](./image-card.md) 平行的哑组件（点击语义由父页面分叉），形态差异：图片是正方形网格卡，音频是「左封面 + 右简介数据」的横向条卡（行式布局更适合展示文本元数据）。

## 布局

- 根为 `div`（`role="button"` + `tabindex` + enter 键支持）而非 button——封面上的播放按钮是嵌套 `<button>`，button 内嵌 button 是非法 HTML（详见下）。
- **左：80px 方形封面区**（`h-20 w-20 rounded-xl`，渐变占位底）：
  - 有生效封面（`effectiveCover(audio)` = `cover_url || embedded_cover_url`，经 `resolveAudioUrl` 解析）→ `object-cover` 图片，hover 轻微放大；
  - 无封面 → 居中音符 SVG（feather music 形状：竖线 + 双符头圆）；
  - **播放 / 暂停按钮**居中覆盖：默认 hover 才显形（`opacity-0 group-hover:opacity-100`），播放中常显；实心三角（播放）/ 双竖条（暂停）；`@click.stop` 避免触发卡片的选中 / 详情点击；
  - **多选徽标**在右上角（`-right-1.5 -top-1.5`）：两态切换，与 ImageCard 的右下角徽标同款交互逻辑（常驻占位避免布局跳动）。
- **右：简介数据三行**（`min-w-0 flex-1`）：
  1. 标题（`title || filename`，truncate）；
  2. 「艺术家 — 专辑」（两者只有其一时单独显示，都缺显示 `—`）；
  3. 元数据行：`formatDuration(时长) · FORMAT · formatBytes(大小) · formatDate(上传时间)`。

## Props / Emits

- props：`audio: AudioItem`（类型见 [`useAudios`](../../composables/useAudios.md)）；`selectable?: boolean`；`selected?: boolean`。
- emits：`click(audio: AudioItem)`。

## 播放控制

经 [`useAudioPlayer`](../../composables/useAudioPlayer.md) 的 `toggle(id, resolveAudioUrl(audio.url))` 与 `isPlaying(id)`——全局单实例（`currentId` / `playing` 走 `useState` 共享），任意时刻至多一个音频在播，切歌即换源。卡片自身不持有任何播放状态。

## 选中态样式

边框三态互斥（computed `borderClass`）：选中（`border-iris-dark ring-2 ring-iris-violet/30`）> 多选未选中（`hover:border-iris-dark/50`）> 默认（`hover:border-iris-violet/60 hover:shadow-sm`），规则与 ImageCard 相同。

## 与其它文件的关系

- 父组件：[`pages/content/audio.vue`](../../pages/content/audio.md)（传入 `selectable` / `selected`，点击分叉）。
- 依赖 [`useAudios`](../../composables/useAudios.md)（`effectiveCover` / `formatDuration` / `resolveAudioUrl`）、[`useImages`](../../composables/useImages.md)（`formatBytes` / `formatDate` 复用）、[`useAudioPlayer`](../../composables/useAudioPlayer.md)（自动导入）。
