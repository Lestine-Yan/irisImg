# 音频管理 / 静态反代说明（irisImg 后端）

> 与图片管线完全对称的音频管线（存储、双通道上传、管理、审计、静态服务），外加音频特有的：**标准元数据解析**、**双层封面**（内嵌抽取 + 用户覆盖）、**视频拒收**。

## 1. 整体设计

```
multipart "file"
  └─ api.AudioAPI（readUploadFile：MaxBytesReader 限长 + 读取）
       └─ service.AudioService.Upload
            1. 空/超限守卫（ErrEmptyFile / ErrFileTooLarge）
            2. audiometa.Sniff 魔数嗅探 → 格式 + 规范化 MIME
               （mp3→audio/mpeg、m4a→audio/mp4、flac→audio/flac）
            3. 白名单比对（storage.audio_allowed_mime_types）
            4. m4a 家族额外 audiometa.HasAudioTrack 结构校验（拒收视频 mp4）
            5. sha256 → GetByHash 秒传（命中直接返回首传记录）
            6. audiometa.Parse 元数据解析（永不失败：标签/技术参数/封面缺失零值降级）
            7. saver.Save → <audio_root>/<YYYY>/<MM>/<sha256>.<format>
            8. 内嵌封面 best-effort 抽取 → SaveFlat covers/<sha256>.<ext>
            9. dao.Create 落库（含全部元数据与双层封面）
```

与图片的差异点：

| 维度 | 图片 | 音频 |
| --- | --- | --- |
| 存储根 / URL 前缀 | `storage.root_dir` / `/imgs` | `storage.audio_root_dir` / `/audios` |
| 格式判定 | `http.DetectContentType` | `audiometa.Sniff` 魔数（DetectContentType 不覆盖 m4a/flac） |
| 元数据 | 宽高（`image.DecodeConfig`） | 标签 + 时长/码率/采样率/声道/位深（4 个纯 Go 库） |
| 封面 | 无 | 内嵌抽取落盘 + 用户覆盖 URL（双层） |
| 大小上限 | `max_upload_size_mb`（20） | `audio_max_upload_size_mb`（100，无损体积大） |
| 仪表盘统计 | 有（总量/存储/趋势） | **无**（v1 刻意只统计图片；audio 事件会自然出现在日志直方图里） |

## 2. 配置

`storage` 段新增四个键（默认值集中在 [`config.ApplyDefaults`](../config/config.md)，见 [`config/config.yaml`](../../../backend/config/config.yaml)）：

| 键 | 默认 | 说明 |
| --- | --- | --- |
| `audio_root_dir` | `data/audios` | 音频落盘根目录；**刻意有默认值**（区别于 `root_dir` 的 fail-fast）——旧部署 config 缺键也能升级即用 |
| `audio_public_base_url` | `""` | 空 → 相对路径 `/audios/...`；非空须带协议（裸域名自动补 https://） |
| `audio_max_upload_size_mb` | `100` | Nginx `client_max_body_size` 需大于此值（模板按 150m 对齐） |
| `audio_allowed_mime_types` | mpeg/mp4/x-m4a/flac | 白名单匹配**嗅探规范化后**的 MIME；`audio/x-m4a` 仅别名，入库永远存 `audio/mp4` |

## 3. 接口

### `POST /api/v1/audios` —— 对外上传音频

API Key 鉴权（readwrite + 按密钥限流），multipart 字段 `file`。按内容魔数校验 mp3/m4a/flac，**扩展名/Content-Type 伪造无效**；带 `vide` 音轨的 mp4 按不支持类型拒收。响应即新增的 `model.Audio`（含元数据与 `key_id`）。

### `GET /api/v1/admin/audios` —— 后台音频列表

JWT + HTTPSOnly。query 同图片：`key_id` / `order`（默认 asc）/ `page` / `page_size`（默认 24）。响应 `{ items, total, page, page_size }`。

### `POST /api/v1/admin/audios` —— 后台直传音频

JWT + HTTPSOnly，`key_id` 为 nil（admin 直传）。

### `DELETE /api/v1/admin/audios` —— 批量删除音频

JWT + HTTPSOnly + 账号密码二次确认（失败 403，不触发前端全局登出）。body `{ username, password, ids }`（1..100）。响应 `{ deleted, ids }`，不存在的 ID 静默跳过。**内嵌封面（内容寻址共享）刻意不随删除清理**。

### `PATCH /api/v1/admin/audios/:id` —— 修改封面 URL

JWT + HTTPSOnly。body `{ "cover_url": "<URL 或空串>" }`（≤2048）：非空 = 设置用户覆盖；**空串 = 清除覆盖**（恢复展示内嵌封面）。响应更新后的完整 `model.Audio`。404 = 音频不存在。

### `GET /api/v1/audios` —— 对外列表（占位 501）

与 `GET /images` 同款占位。

### 错误码（上传两通道共用）

| 状态码 | 场景 |
| --- | --- |
| 400 | 缺 `file` 字段 / 内容为空 / 格式不在白名单（含视频 mp4、损坏结构） |
| 401 | API Key 缺失/无效（`X-API-Key`）；JWT 缺失或失效 |
| 403 | 只读密钥 POST；二次确认密码错误；生产 HTTPSOnly 拦截 |
| 413 | 超过 `storage.audio_max_upload_size_mb`（MaxBytesReader 先拦） |
| 429 | 密钥限流（对外通道） |
| 500 | 其它内部错误 |

## 4. 元数据与封面语义

- **解析永不阻断上传**：标签缺失 → 空串/0；时长未知 → 0；标题缺失 → 回退「文件名去扩展名」。
- **码率**是 `文件大小×8/时长` 的平均值（VBR 亦然）。
- **双层封面**：生效封面 = `cover_url || embedded_cover_url`（前端 `effectiveCover` 计算）；`embedded_cover_url` 上传时抽取后不变，`cover_url` 是用户覆盖（详情页可改/可清除）。
- **封面落盘**在 `<audio_root>/covers/<sha256>.<ext>`（[`storage.SaveFlat`](internal/pkg/storage.md) 平铺、内容寻址）：同一专辑封面被多个音频共享天然去重；删除音频**不删封面**——孤儿清扫是后续可选的运维操作。
- **秒传语义**：同哈希音频返回首传记录（含当时的元数据与已改的封面），与图片一致。
- m4a 的采样率/声道 v1 刻意不解析（stsd 随编码器差异大），详情页展示 0；mp3/flac 均有。

## 5. 部署与 Nginx 反代约定

生产由 Nginx 直接静态服务 `/audios/`（alias 指向 `storage.audio_root_dir`，与 `/imgs/` 同款），后端 `serveFiles("/audios", ...)` 仅为开发期兜底。**三方对齐**（新增类型时三处都要改）：

1. nginx `/audios/` 的扩展名正则：`\.(mp3|m4a|flac|png|jpg|jpe?g|gif|webp)$`（封面图片也放行——封面与音频共用一棵树）；
2. `config.yaml` `storage.audio_allowed_mime_types`（+ `allowed_mime_types` 决定封面扩展名）；
3. 后端 [`internal/router/static.go`](internal/router/static.md) 的 `audioMimeExt` 闭合映射。

其它要点：

- `client_max_body_size` 须大于 `audio_max_upload_size_mb`（模板 25m → **150m**）；`/api/` 的 `proxy_read_timeout` 60s → **300s**（100MB 慢速上传会超 60s）。
- Nginx 与 `http.FileServer` 原生支持 Range 请求，音频拖动进度条无需额外配置。
- HTTPS 模板的 `/audios/` location 内需重申 server 级安全头（nginx add_header 继承规则，同 `/imgs/`）。

## 6. 常见排错

| 现象 | 原因 / 处置 |
| --- | --- |
| 正常 m4a 上传被 400 拒收 | 内容无 `ftyp` 盒（如 ADTS 裸 AAC——**不在支持范围**）；或文件结构损坏导致 `HasAudioTrack` 判 false。 |
| mp4 视频被拒（预期行为） | 嗅探识别为 mp4 家族后结构校验发现 `vide` 轨；仅收音频轨文件。 |
| 音频直链 404 | 扩展名不在 nginx / `audioMimeExt` 白名单；检查三方对齐。 |
| 上传 100MB 音频 413 | 后端 `audio_max_upload_size_mb` 与 nginx `client_max_body_size` 都要够大。 |
| 卡片封面空白但详情有内嵌封面 | 前端 `effectiveCover` 应取 `cover_url || embedded_cover_url`；检查是否有旧缓存数据。 |
| 升级后启动失败 `audio_root_dir 不能为空` | 该键**有默认值**，此报错说明显式配了空串；填值或删掉该行走默认。 |

## 7. 不在本次范围（刻意）

- 仪表盘音频统计（总量/占用/趋势）——v1 仪表盘只统计图片，后续可加统计卡；
- 封面孤儿清扫脚本（covers/ 下内容寻址，删除音频不清理）；
- WAV / OGG / OPUS 等其它格式（白名单与解析器都只覆盖 mp3/m4a/flac）；
- m4a 采样率/声道解析（stsd 复杂度高、展示价值低）；
- 歌词、波形、播放进度条（前端仅播放/暂停控制）；
- 音频转码 / 压缩。
