# `internal/pkg/audiometa/audiometa.go`

> 文档路径约定：`internal/pkg/audiometa/audiometa.go` → `docs/backend/internal/pkg/audiometa/audiometa.md`。

提供音频文件的标准元数据解析（标签 + 技术参数 + 内嵌封面）。设计原则：**永不因解析失败阻断上传**——任何子解析器失败只让对应字段零值降级，格式合法性由 [`Sniff`](sniff.md) + service 白名单把关。四个职责分离的纯 Go 依赖集中在本包消化，service 层保持薄。

## 依赖分工

| 依赖 | 职责 |
| --- | --- |
| `github.com/dhowden/tag` | mp3(ID3)/m4a(MP4 atom)/flac(Vorbis comment) 标签字段 + 内嵌封面 |
| `github.com/tcolgate/mp3` | mp3 帧遍历（时长/采样率/声道） |
| `github.com/go-flac/go-flac` | FLAC STREAMINFO（时长/采样率/声道/位深） |
| `github.com/abema/go-mp4` | MP4 盒子结构（mvhd 时长、hdlr 音轨校验） |

## 类型

### `Format`

容器格式枚举，值即小写文件扩展名（落盘时直接用）：`FormatMP3 "mp3"` / `FormatM4A "m4a"` / `FormatFLAC "flac"` / `FormatUnknown ""`。

### `Info`

一份音频的标准元数据汇总：

- 标签字段：`Title` / `Artist` / `Album` / `AlbumArtist` / `Composer` / `Genre` / `Year` / `Track` / `Disc`（缺失为零值）；
- 技术参数：`DurationSeconds`（0 = 未知）/ `SampleRate`（Hz）/ `Channels` / `BitDepth`（仅 FLAC 有意义，有损格式恒 0）/ `BitrateKBPS`（平均码率）；
- 内嵌封面：`PictureData []byte`（nil = 无）、`PictureExt`（不含点号；空 = 无法识别，调用方应跳过落盘）。

## 函数

### `Parse(content []byte, f Format, size int64) Info`

按格式分派到 `parseMP3` / `parseM4A` / `parseFLAC`，**永不返回 error**。`FormatUnknown` 或未知格式返回零值 `Info`。

- `parseMP3`：`tag.ReadFrom` 取标签；tcolgate 帧遍历累计每帧样本数，时长 = `总样本/采样率`（VBR 也精确；比逐帧 `Duration()` 相加更稳——后者在毫秒粒度截断，长文件有累积误差）。首帧确定采样率与声道（`SingleChannel` → 1，其余 → 2）。
- `parseFLAC`：`tag.ReadFrom` 取标签；`flac.ParseMetadata` 只读元数据块（不读帧数据，大文件也只耗头部字节），`GetStreamInfo` 已把声道/位深从「存储值−1」归一化为真实值；时长 = `总样本/采样率`。
- `parseM4A`：`tag.ReadFrom` 取标签；go-mp4 盒子遍历找 `moov.mvhd`，时长 = `duration/timescale`（version 1 用 64 位 `DurationV1`）；全 1 哨兵值视为未知。m4a 的采样率/声道在 stsd 里且随编码器差异大，**v1 刻意不解析**（保持 0），时长与码率足够展示。
- `BitrateKBPS = size*8/duration/1000` 统一计算（平均码率，VBR 是均值）。

### `HasAudioTrack(content []byte) bool`

对 MP4 家族容器做「真音频」结构校验：存在 `soun` 音轨 handler 且无 `vide` 视轨。

背景：ffmpeg 默认把 `.m4a` 的 ftyp major_brand 编成 `isom`，**无法靠 brand 区分音频/视频**，只有走到 `moov.trak.mdia.hdlr` 看 HandlerType 才可靠。mp3/flac 的魔数本身就是格式的完整定义，无需此校验。结构损坏（盒子读不通）返回 false——交由上层按不支持的类型拒收。

## 封面扩展名推断链

`pictureExt(pic *tag.Picture)`：

1. 标签自带 `Ext`（MP4 covr 常见），`jpeg` 归一化为 `jpg`；
2. 按 `MIMEType` 映射（image/jpeg→jpg 等）；
3. 按**字节魔数**嗅探（`sniffImageExt`：JPEG `FF D8 FF`、PNG `\x89PNG`、WebP `RIFF...WEBP`、GIF、BMP）。Go 的 `http.DetectContentType` 不识别 WebP，故手工判定；
4. 仍未知返回空串——静态扩展名白名单服务不了未知类型，调用方（service）应跳过封面落盘。

## 与其它包的关系

```
service.AudioService ──► audiometa.Sniff        (格式嗅探 → 白名单)
                     ├─► audiometa.HasAudioTrack (m4a 音轨校验)
                     └─► audiometa.Parse         (元数据 + 封面)
```
