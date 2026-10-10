# `internal/pkg/audiometa/sniff.go`

> 文档路径约定：`internal/pkg/audiometa/sniff.go` → `docs/backend/internal/pkg/audiometa/sniff.md`。

基于字节魔数识别音频容器格式。**刻意不复用 `http.DetectContentType`**：它的嗅探表覆盖 mp3（ID3）但不覆盖 m4a（ftyp 盒）与 flac（`fLaC`），而本项目坚持「嗅探内容而非信任客户端 Content-Type / 扩展名」的纵深防御，故自持一份最小判定集。

## `Sniff(head []byte) (Format, string)`

返回 `(格式, 规范化 MIME)`；未识别返回 `(FormatUnknown, "")`，由上层按不支持的类型拒收（400）。判定规则（顺序即优先级）：

| 规则 | 判定 | 结果 |
| --- | --- | --- |
| FLAC | 头 4 字节 `fLaC` | `(flac, "audio/flac")` |
| MP4 家族 | `[4:8] == "ftyp"` | `(m4a, "audio/mp4")` |
| MP3 | 头 3 字节 `ID3`（ID3v2 标签在前） | `(mp3, "audio/mpeg")` |
| MP3 | 裸 MPEG 帧同步 + 头部位合法 | `(mp3, "audio/mpeg")` |

- MP4 家族**不按 ftyp major brand 区分** M4A/isom：ffmpeg 默认把 `.m4a` 编成 major_brand=isom，严格按 brand 会误拒大量正常文件；真正的音频/视频区分由 [`HasAudioTrack`](audiometa.md) 走到 `moov.trak.mdia.hdlr` 做结构校验。
- 裸 MPEG 帧同步校验 `validMPEGFrameHeader`（对齐 tcolgate/mp3 的帧合法性判定）：帧同步字 `0xFF 0xEx` + version/layer 非保留值 + 码率索引非 0（free）非 0xF（bad）+ 采样率索引 != 3（reserved），降低把任意 `0xFF` 开头数据误判为 mp3 的概率。
- `head` 传入内容前若干字节即可（建议 ≥ 16），函数内部对所有访问做了长度保护。

## 规范化 MIME 映射

入库永远存规范化值（扩展名即格式名）：mp3→`audio/mpeg`、m4a→`audio/mp4`、flac→`audio/flac`。`audio/x-m4a` 仅作为 `storage.audio_allowed_mime_types` 白名单里的历史习惯别名被接受，嗅探结果永远是 canonical 值——存储与展示稳定，不受白名单别名影响。

## 与其它包的关系

- `service.AudioService.Upload` 第一步调用（先于白名单判断与元数据解析）；
- 三方对齐之一：本判定结果 ↔ `config.yaml storage.audio_allowed_mime_types` ↔ nginx `/audios/` 扩展名白名单（见 `docs/backend/AUDIO.md`）。
