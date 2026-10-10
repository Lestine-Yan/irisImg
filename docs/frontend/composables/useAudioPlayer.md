# `frontend/app/composables/useAudioPlayer.ts`

全局**单实例**的音频播放 / 暂停控制，供音频卡片封面上的播放按钮共用——任意时刻至多一个音频在播，切歌即换源。

## 设计

- **模块级单例 `HTMLAudioElement`**：`audioEl` 在首次 `toggle` 时才懒创建（`preload = 'none'`，不预载不占资源）。本项目 SPA 模式（`ssr: false`），模块级单例无 SSR 状态串扰问题。
- **响应式状态走 `useState`**（`audio-player:current` / `audio-player:playing`）：跨组件共享且 Nuxt 语义安全。`currentId` 是「正在播（或暂停中保留进度）」的音频 ID；暂停时 `playing=false` 但保留 `currentId`，按钮显示「继续播放」。
- `ended` 事件把 `playing` 归 false（播完按钮回到播放态）。
- `play()` 的 Promise 被 catch：自动播放策略 / 网络失败静默回退，不打断界面。

## API

| 成员 | 说明 |
| --- | --- |
| `toggle(id, url)` | 同 ID：播放中 → 暂停；已暂停 → 从保留进度继续。新 ID：停掉旧的、换源从头播放 |
| `isPlaying(id)` | 该音频是否播放中（供卡片按钮图标 / 常显态切换） |
| `currentId` / `playing` | 响应式状态（一般经 `isPlaying` 间接消费） |

刻意不做进度条 / 拖动 UI（浏览器原生支持 Range 请求，需要精听时在详情弹窗复制直链）。

## 与其它文件的关系

- 消费方：[`AudioCard.vue`](../components/content/audio-card.md) 封面上的播放 / 暂停按钮（`@click.stop` 避免触发卡片选中/详情）。
- 音频 URL 来自 [`useAudios`](./useAudios.md) 的 `resolveAudioUrl(item.url)`。
