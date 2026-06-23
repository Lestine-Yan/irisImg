# 图片上传 / 静态反代说明（irisImg 后端）

> 本文档跨文件讲清楚「一张图片是怎么从客户端跑到磁盘、再被外部 URL 访问到的」。
> 各 `.go` 文件的逐文件文档见各自目录的 `.md`；API 密钥鉴权链路见 [`APIKEY.md`](./APIKEY.md)。

---

## 1. 整体设计

- 上传通道挂在 **API 密钥鉴权**之下（独立于后台 JWT），由 [`middleware.APIKeyAuth`](./internal/middleware/apikey.md) 保证 POST 必为 `readwrite` 密钥。
- 文件名由内容 **SHA256** 计算得出，**天然唯一、天然去重**：同一张图二次上传自动秒传，复用首次记录。
- 落盘目录按 `<root>/<YYYY>/<MM>/<sha256>.<ext>` 排布，避免单目录文件过多。
- **真实 MIME 嗅探**（`http.DetectContentType`）+ 白名单，不信任客户端 `Content-Type`，扩展名由后端推导。
- **对外访问 URL** 由配置 `storage.public_base_url` + 相对路径拼接：空 → `/imgs/...`（前端/Nginx 同域反代）；填了如 `https://img.example.com` → 绝对地址。
- 元数据落库 `images` 表（含 `key_id`），记录是哪把密钥添加的图片，便于审计与后续按密钥维度展示。

参与的代码文件：

| 角色 | 文件 |
| --- | --- |
| Schema | `ent/schema/image.go` |
| 配置 | `config/config.go`、`config/config.yaml`（`storage` 段） |
| 存储工具 | `internal/pkg/storage/storage.go` |
| DTO | `internal/model/image.go`（`Image` / `UploadImageInput`） |
| DAO | `internal/dao/dao.go`、`internal/dao/entdao/image.go` |
| 业务逻辑 | `internal/service/image.go` |
| 中间件 | `internal/middleware/apikey.go`（鉴权 + 限流，已存在） |
| 控制器 | `internal/api/image.go` |
| 统一响应 | `internal/pkg/response/response.go`（`CodePayloadTooLarge` / `PayloadTooLarge`） |
| 路由装配 | `internal/router/router.go` |
| 启动入口 | `cmd/server/main.go`（启动期构造 `storage.Saver`） |

## 2. 配置

```yaml
storage:
  root_dir: "data/imgs"             # 落盘根目录；相对路径相对进程 cwd，生产建议改绝对路径
  public_base_url: ""               # 空 → 返回 /imgs/<rel>，前端/Nginx 同域反代
                                    # 非空 → 例如 "https://img.example.com"（结尾不带 /）
  max_upload_size_mb: 20            # 单次上传字节上限（MiB），<=0 回退 20
  allowed_mime_types:               # 真实 MIME 白名单（后端嗅探，不信任客户端 Content-Type）
    - "image/png"
    - "image/jpeg"
    - "image/gif"
    - "image/webp"
```

详见 [`config.md`](./config/config.md)。

## 3. 接口

### `POST /api/v1/images` —— 添加图片

- **鉴权**：`X-API-Key: <readwrite 密钥>`（由 [`middleware.APIKeyAuth`](./internal/middleware/apikey.md) 校验；只读密钥被该中间件直接 403）。
- **请求体**：`multipart/form-data`，唯一字段 `file`（图片二进制）。
- **成功响应**：`200` + `data` 为完整 `model.Image`（`id / url / stored_path / size / mime_type / width / height / hash / key_id / created_at`）。
- **错误**：

| HTTP | 业务码 | 场景 |
| --- | --- | --- |
| 400 | `CodeBadRequest` | 缺少 `file` 字段 / 内容为空 / 嗅探出的 MIME 不在白名单 |
| 401 | `CodeAPIKeyMissing` / `CodeAPIKeyInvalid` | 缺密钥 / 格式非法 / 已吊销 |
| 403 | `CodeForbidden` | 只读密钥写入 |
| 413 | `CodePayloadTooLarge` | 超过 `storage.max_upload_size_mb` |
| 429 | `CodeTooManyRequests` | 触发该密钥限流 |
| 500 | `CodeServerError` | 落盘 / 落库失败等内部错误 |

### `GET /api/v1/images` —— 申请图片（占位）

任意有效密钥可访问，**当前固定返回 501**，待前端列表页接入时再实现。

## 4. 上传链路

```
client           api.ImageAPI             service.ImageService           dao.ImageDAO         pkg/storage.Saver
  │ POST /images    │                          │                              │                       │
  │ X-API-Key       │                          │                              │                       │
  │ form file=…     │                          │                              │                       │
  │ ───────────────►│ MaxBytesReader+FormFile  │                              │                       │
  │                 │ ──── Upload(input) ────► │                              │                       │
  │                 │                          │ DetectContentType + 白名单    │                       │
  │                 │                          │ sha256                       │                       │
  │                 │                          │ GetByHash ─────────────────► │                       │
  │                 │                          │ ◄───── (existing / NotFound) │                       │
  │                 │                          │ DecodeConfig (W/H, 失败=0,0)  │                       │
  │                 │                          │ Save(content, hash, ext) ─── │ ────────────────────► │
  │                 │                          │                              │       <root>/YYYY/MM  │
  │                 │                          │ PublicURL(rel)               │                       │
  │                 │                          │ Create(*model.Image) ──────► │                       │
  │                 │ ◄────── *model.Image ─── │                              │                       │
  │ ◄── 200 Body ───│                          │                              │                       │
```

### 关键节点

1. **MaxBytesReader 早拦**：`api.ImageAPI.Create` 用 `http.MaxBytesReader(c.Writer, c.Request.Body, svc.MaxBytes())` 把超大请求体拦在 Multipart 解析之前，节省内存。
2. **真实 MIME 嗅探**：service 内 `http.DetectContentType` 看头部 512 字节，剥掉 `;charset=...` 等参数后比对白名单。
3. **秒传**：`dao.GetByHash` 命中即直接返回该记录，**不重复写盘、不重复落库**（连 `key_id` 也保持首次写入的值；后续若需记录"谁又传过一次"再加关联表）。
4. **写盘原子**：`pkg/storage.Saver.Save` 写到同目录临时文件再 `os.Rename`，避免半写状态被读取。
5. **URL 拼接**：`Saver.PublicURL`；空 base_url 返回 `/imgs/<rel>`，配 base_url 返回 `<base>/<rel>`。
6. **落库**：`dao.ImageDAO.Create` 写入 `images` 表，并通过 schema 上的 `key` edge 关联到 `api_keys.id`。

## 5. 部署与 Nginx 反代约定

- `storage.root_dir` 与 Nginx `location /imgs/` 暴露的物理路径**必须一致**。例如：

  ```yaml
  storage:
    root_dir: "/var/lib/irisImg/imgs"
    public_base_url: ""
  ```

  ```nginx
  location /imgs/ {
    alias /var/lib/irisImg/imgs/;
    expires 30d;
    add_header Cache-Control "public, immutable";
  }
  ```

- 想用独立图片域名（如 `https://img.example.com`）：把 `public_base_url` 配上（**结尾不带斜杠**），并在那个域名同样反代到 `root_dir`。
- 备份 / 迁移时同步处理 `root_dir` 与数据库；hash 文件名让"按需补图"也很简单。
- **本次未提供 Nginx 配置**，仅在文档里约定路径，等部署时再写。

## 6. 常见排错

| 现象 | 原因 |
| --- | --- |
| 405 `Method Not Allowed` | 用了未注册的方法（如 PUT），中间件之前就被 Gin 拒 |
| 401 `CodeAPIKeyMissing` | 没带 `X-API-Key` |
| 401 `CodeAPIKeyInvalid` | 密钥格式不对 / 不存在 / 被吊销 |
| 403 `CodeForbidden` | 用只读密钥 POST |
| 413 `CodePayloadTooLarge` | 文件大于 `max_upload_size_mb` |
| 400 + "不支持的图片类型" | 内容嗅探结果不在 `allowed_mime_types`，伪造 Content-Type 无效 |
| 上传成功但 `width=0,height=0` | 是 webp/avif 等标准库未注册解码器的格式；不影响存储与 URL |
| 上传成功但 URL 拼接异常 | 检查 `public_base_url` 是否带了尾斜杠（应当不带） |

## 7. 不在本次范围

- Nginx 配置文件本身（仅在本文档里约定路径）。
- `GET /api/v1/images` 列表 / 单图查询接口。
- 后台 JWT 直传通道（目前所有上传都走 API Key；后续接前端时再补 JWT 入口，`KeyID` 留空）。
- 缩略图、EXIF 清理、防盗链、对象存储后端。
