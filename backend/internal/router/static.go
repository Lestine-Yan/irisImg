package router

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// imageMimeExt 是「已知图片 MIME -> 可服务扩展名」的封闭映射，用于把
// storage.allowed_mime_types 折算成 /imgs 静态服务的扩展名白名单。
//
// 仅收录图片类型：即便上传白名单被误配进 application/pdf 等非图片 MIME，
// 也不会为其生成可服务扩展名，构成「即使 root_dir 配错也只暴露图片」的纵深防御。
// 与 service.extFromMime 的已知集合保持一致：新增可上传图片类型时两处都要加。
var imageMimeExt = map[string][]string{
	"image/png":     {"png"},
	"image/jpeg":    {"jpg", "jpeg"},
	"image/gif":     {"gif"},
	"image/webp":    {"webp"},
	"image/bmp":     {"bmp"},
	"image/svg+xml": {"svg"},
}

// allowedImageExtensions 把 storage.allowed_mime_types 折算成扩展名集合（小写、不含点）。
// 仅命中 imageMimeExt 的 MIME 才计入；未知 MIME（如 application/pdf）一律不服务，
// 避免非图片文件经 /imgs 静态暴露。返回空集合时 serveImages 会拒绝一切请求。
func allowedImageExtensions(allowedMimeTypes []string) map[string]struct{} {
	exts := make(map[string]struct{})
	for _, m := range allowedMimeTypes {
		for _, e := range imageMimeExt[strings.ToLower(strings.TrimSpace(m))] {
			exts[e] = struct{}{}
		}
	}
	return exts
}

// isAllowedExt 判断 URL 末段扩展名是否在白名单内。
//
// 拒绝：空路径、目录（末尾斜杠导致末段为空）、无扩展名、以点开头的隐藏文件（如 .gitignore）。
// 扩展名大小写不敏感。注意此处只做扩展名判断，路径逃逸（../）交由 http.FileServer 兜底。
func isAllowedExt(p string, allowed map[string]struct{}) bool {
	if p == "" {
		return false
	}
	name := p
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	if name == "" {
		return false
	}
	dot := strings.LastIndex(name, ".")
	if dot <= 0 { // 无扩展名，或以点开头（隐藏文件）
		return false
	}
	_, ok := allowed[strings.ToLower(name[dot+1:])]
	return ok
}

// serveImages 返回 /imgs 静态文件 handler，带「图片扩展名白名单」前置过滤：
//
// 仅放行 allowedExts 中的扩展名（.png/.jpg/.jpeg/.gif/.webp 等），拒绝 .yaml/.db/.go
// 等。即使 storage.root_dir 被误配成工作目录或 backend/ 本身，未认证访客也无法经
// /imgs 下载 config.yaml / irisImg.db / 源码——这是即便用户误配也成立的纵深防御。
//
// 路径清洗与 .. 逃逸防护复用 http.FileServer（Go net/http 内建，path.Clean + http.Dir
// 会把 .. 折叠回 root 之下），此处不重复造轮子。未命中白名单一律回 404（而非 403），
// 避免向未认证访客泄露文件是否存在。目录列表同样因末段无扩展名被挡回 404。
//
// 注意：handler 同时挂 GET 与 HEAD（见 router.go），与原 r.Static 行为对齐。
func serveImages(rootDir string, allowedExts map[string]struct{}) gin.HandlerFunc {
	handler := http.StripPrefix("/imgs", http.FileServer(http.Dir(rootDir)))
	return func(c *gin.Context) {
		if !isAllowedExt(c.Param("filepath"), allowedExts) {
			c.Status(http.StatusNotFound)
			return
		}
		handler.ServeHTTP(c.Writer, c.Request)
	}
}
