# `frontend/app/components/content/AudioUploadPanel.vue`

音频中心的拖拽上传栏，Nuxt 自动导入标签为 `<ContentAudioUploadPanel />`。与 [`UploadPanel.vue`](./upload-panel.md) 完全同构（拖拽 / 点击选择、深度计数防高亮闪烁、顺序队列上传、逐项状态点与错误文案、移除与清空已完成），仅接收范围与接口不同。

## 与图片上传栏的差异

| 维度 | UploadPanel（图片） | AudioUploadPanel（音频） |
| --- | --- | --- |
| `accept` | `image/*` | `.mp3,.m4a,.flac,audio/*` |
| 拖入过滤 | `f.type.startsWith('image/')` | **扩展名 mp3/m4a/flac 或 `audio/*` MIME**（`isAudioFile`：部分 OS 对 .flac 报空 MIME，不能只信 `f.type`） |
| 提示文案 | 「支持 PNG / JPEG / GIF / WebP」 | 「支持 MP3 / M4A / FLAC」 |
| 上传接口 | `useImages().upload` → `POST /admin/images` | `useAudios().upload` → `POST /admin/audios` |
| 图标 | 云上传 | 音符（feather music 形状） |

## 队列状态机

`pending → uploading → success | error`，`processQueue` 顺序上传（一次一个，避免并发冲击），每个成功即 `emit('uploaded', audio)` 让父页面刷新列表；错误优先取后端业务 message（ofetch 挂在 `err.data` 上），与图片上传栏同款 `resolveErrorMsg`。

## Props / Emits

- 无 props。
- emits：`uploaded(audio: AudioItem)`。

## 与其它文件的关系

- 父组件：[`pages/content/audio.vue`](../../pages/content/audio.md)（与多选模式互斥显隐）。
- 依赖 [`useAudios`](../../composables/useAudios.md)（`upload`）、[`useImages`](../../composables/useImages.md)（`formatBytes` 复用）。
