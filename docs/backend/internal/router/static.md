# `internal/router/static.go`

`/imgs` 静态图片服务的 handler 工厂，在原 `r.Static` 的基础上叠加「图片扩展名白名单」前置过滤，作为 `storage.root_dir` 误配时的纵深防御。

## 背景

原 `r.Static("/imgs", cfg.Storage.RootDir)` 把整个落盘目录无差别地交给 `http.FileServer`，且挂在全局 engine、无认证中间件。默认 `root_dir: data/imgs` 是安全的，但一旦用户把 `root_dir` 配成 `.` / `..` / `/` / 工作目录或其祖先，未认证访客就能直接 `GET /imgs/config/config.yaml` 下载配置（含 `auth.password` + `auth.jwt.secret`）、`GET /imgs/data/irisImg.db` 下载整库。这是「用户误配 + 代码无防护」的组合漏洞。

本文件提供即便误配也不暴露非图片文件的兜底：**只放行图片扩展名**。

## 导出/内部符号

### `imageMimeExt`

```go
var imageMimeExt = map[string][]string{
    "image/png":     {"png"},
    "image/jpeg":    {"jpg", "jpeg"},
    "image/gif":     {"gif"},
    "image/webp":    {"webp"},
    "image/bmp":     {"bmp"},
    "image/svg+xml": {"svg"},
}
```

「已知图片 MIME -> 可服务扩展名」的封闭映射。仅收录图片类型：即便 `storage.allowed_mime_types` 被误配进 `application/pdf` 等非图片 MIME，也不会为其生成可服务扩展名。与 [`service.extFromMime`](../service/image.md) 的已知集合保持一致--新增可上传图片类型时两处都要加。

### `allowedImageExtensions(allowedMimeTypes []string) map[string]struct{}`

把 `storage.allowed_mime_types` 折算成扩展名集合（小写、不含点）。仅命中 `imageMimeExt` 的 MIME 才计入；未知 MIME 一律不服务。返回空集合时 `serveImages` 拒绝一切请求（对应「显式空白名单 = 禁止上传，也就无图可服务」）。

### `isAllowedExt(p string, allowed map[string]struct{}) bool`

判断 URL 末段扩展名是否在白名单内。拒绝：空路径、目录（末尾斜杠导致末段为空）、无扩展名、以点开头的隐藏文件（如 `.gitignore`）。扩展名大小写不敏感。**只做扩展名判断，路径逃逸（`../`）交由 `http.FileServer` 兜底**。

### `serveImages(rootDir string, allowedExts map[string]struct{}) gin.HandlerFunc`

返回 `/imgs` 的 `gin.HandlerFunc`：

1. 取 catch-all 参数 `filepath`，先过 `isAllowedExt`；未命中回 **404**（而非 403，避免向未认证访客泄露文件是否存在）。
2. 命中则交给 `http.StripPrefix("/imgs", http.FileServer(http.Dir(rootDir)))` 服务。

## 两层防护

| 层 | 职责 | 归属 |
| --- | --- | --- |
| 扩展名白名单 | 只放行图片扩展名，拒绝 `.yaml`/`.db`/`.go` 等 | 本文件（`serveImages` + `isAllowedExt`） |
| 路径逃逸防护 | `../` 折叠回 root 之下 | `http.FileServer`（Go net/http 内建，`path.Clean` + `http.Dir.Open`） |
| 启动期 fail-fast | 拒绝 `root_dir` 指向 cwd 本身/祖先 | [`storage.NewSaver`](../pkg/storage.md) 的 `guardRootDir` |

目录列表（`GET /imgs/`、`GET /imgs/2026/06/`）因末段无扩展名被 `isAllowedExt` 挡回 404，比原 `r.Static`（默认开启目录列表）更严格。

## 注册

由 [`router.New`](./router.md) 注册，同时挂 GET 与 HEAD（与原 `r.Static` 行为对齐）：

```go
imgServe := serveImages(cfg.Storage.RootDir, allowedImageExtensions(cfg.Storage.AllowedMimeTypes))
r.GET("/imgs/*filepath", imgServe)
r.HEAD("/imgs/*filepath", imgServe)
```

## 与其它包的关系

- 消费 `cfg.Storage.RootDir` 与 `cfg.Storage.AllowedMimeTypes`（由 [`router.New`](./router.md) 注入）。
- 与 [`storage.Saver`](../pkg/storage.md) 共享 root 目录与 `/imgs` URL 前缀约定：`Saver` 写盘、`serveImages` 读盘。
- 生产环境由 Nginx 反代 `/imgs/`，本 handler 仅开发期兜底（见 [`IMAGE.md`](../../IMAGE.md)）；`deploy/nginx/*.example` 模板的 `/imgs/` 块已内置同样的扩展名白名单（`if ($uri !~* \.(png|jpg|jpeg|gif|webp)$) return 404`），前后端双重纵深防御。扩展 `allowed_mime_types` 新增图片类型时，Nginx 正则与本文 `imageMimeExt` 都要同步。

## 注意

- 仅兜底「文件内容泄露」；不替代 [`storage.guardRootDir`](../pkg/storage.md) 的启动期误配拦截，二者互补。
- `http.FileServer` 的 `..` 防护是 Go 标准库长期保证的行为，本文件不重复实现。
