# `internal/service/audio.go`

> 文档路径约定：`internal/service/audio.go` → `docs/backend/internal/service/audio.md`。

音频上传的业务编排：嗅探 → 去重 → 元数据解析 → 落盘 → 落库。与 [image.go](image.md) 平行；大小/空文件/MIME 校验哨兵（`ErrEmptyFile` / `ErrFileTooLarge` / `ErrUnsupportedMime`）与图片共用，本文件新增 `ErrAudioNotFound`（按 ID 操作时目标不存在）。

## 类型

- `AudioService{ dao dao.AudioDAO, saver *storage.Saver, cfg, maxBytes, allowedMime }`——`saver` 是**音频存储器**（`storage.NewAudioSaver` 构造，URL 前缀 `/audios`，根目录 `storage.audio_root_dir`）。
- `NewAudioService(d, s, cfg)`：构造期把 `AudioMaxUploadSizeMB`（<=0 兜底 100）与 `AudioAllowedMimeTypes` 预处理成 `maxBytes` 与小写查表 `allowedMime`。
- `MaxBytes()`：供 api 层设置 `http.MaxBytesReader` 阈值。

## 方法

### `Upload(ctx, in *model.UploadAudioInput) (*model.Audio, error)`

主流程（错误 → api 层映射 400/413/500）：

1. 空/超限守卫（复用图片哨兵）；
2. **魔数嗅探** [`audiometa.Sniff`](../pkg/audiometa/sniff.md)：`FormatUnknown` 或规范化 MIME 不在白名单 → `ErrUnsupportedMime`；**MP4 家族额外 `HasAudioTrack` 结构校验**（拒收视频文件与损坏结构）；
3. sha256 → `GetByHash` **秒传**（返回首传记录，含当时的元数据与封面，不重复写盘落库）；
4. [`audiometa.Parse`](../pkg/audiometa/audiometa.md) 解析元数据（**永不失败**，缺失零值降级）；标签无标题时**回退为文件名去扩展名**（音乐库常见兜底）；
5. `saver.Save(content, hash, string(format), now)`——扩展名即格式名，路径 `<audio_root>/<YYYY>/<MM>/<sha256>.<format>`；`PublicURL` 生成对外地址（`/audios/...`）；
6. **内嵌封面抽取落盘**（best-effort，失败不阻断）：封面字节 sha256 命名，`saver.SaveFlat("covers/<hash>.<ext>")` 平铺存储，`embedded_cover_url = PublicURL(coverRel)`；同一专辑封面被多个音频共享时天然去重；
7. `dao.Create` 落库，回填 ID / CreatedAt。

### `List(ctx, q)` / `GetByID(ctx, id)`

列表兜底 `Limit<=0 → 24`；`GetByID` 未找到映射 `ErrAudioNotFound`。

### `BatchDelete(ctx, ids) (int, []int, error)`

与图片完全对称：`ListByIDs` 取现存记录 → 逐条 best-effort `saver.Delete(StoredPath)` → `DeleteByIDs` 批删记录；非事务（同款权衡）。**刻意不删内嵌封面**——内容寻址共享文件，删一个音频不应影响其余音频的封面展示；孤儿清扫留作后续运维操作。

### `UpdateCover(ctx, id, coverURL string) (*model.Audio, error)`

trim 后空串转 `nil`（清除覆盖、恢复内嵌封面），非空直接覆盖；未找到 `ErrAudioNotFound`。长度上限由 api 层 binding（`max=2048`）把关。

## 与其它包的关系

```
api.AudioAPI ──► service.AudioService ──► dao.AudioDAO（接口）
                      │        └─► audiometa（Sniff/HasAudioTrack/Parse）
                      └─► storage.Saver（NewAudioSaver：Save/SaveFlat/PublicURL/Delete）
```
