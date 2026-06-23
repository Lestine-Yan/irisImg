package api

import (
	"errors"
	"io"
	"net/http"

	"github.com/Lestine-Yan/irisImg/backend/internal/middleware"
	"github.com/Lestine-Yan/irisImg/backend/internal/model"
	"github.com/Lestine-Yan/irisImg/backend/internal/pkg/response"
	"github.com/Lestine-Yan/irisImg/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// ImageAPI 是图片相关接口的控制器。
//
// 当前已落地 POST /images（添加图片），由 API Key 鉴权中间件保护
// （只读密钥访问 POST 会被中间件返回 403）。
// GET /images 暂为占位，待接入前端时再实现。
type ImageAPI struct {
	svc *service.ImageService
}

// NewImageAPI 通过依赖注入构造控制器。
func NewImageAPI(svc *service.ImageService) *ImageAPI {
	return &ImageAPI{svc: svc}
}

// uploadFormField 是上传字段名，固定为 "file"。
const uploadFormField = "file"

// List 处理 GET /images（占位）。任意有效密钥均可访问。
// 真实实现待前端列表页接入时再补，当前显式返回 501 与其它鉴权失败状态码区分开。
func (h *ImageAPI) List(c *gin.Context) {
	response.Fail(c, http.StatusNotImplemented, response.CodeServerError, "图片列表接口尚未实现（占位）")
}

// Create 处理 POST /images，落地一张图片：
//
//	请求：multipart/form-data，字段名 "file"，需 readwrite 密钥（由中间件保证）。
//	响应：200 + 完整的 model.Image（含对外 URL、宽高、hash、key_id 等）。
//
// 错误：
//   - 400 缺少文件 / 内容为空 / MIME 不在白名单
//   - 413 超过 storage.max_upload_size_mb
//   - 500 其它内部错误
func (h *ImageAPI) Create(c *gin.Context) {
	// 用 MaxBytesReader 把超大请求体拦在更早阶段，避免分配大内存。
	// 上限取 service 已经算好的 maxBytes。
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.svc.MaxBytes())

	fileHeader, err := c.FormFile(uploadFormField)
	if err != nil {
		// MaxBytesReader 触发时 FormFile 会返回 "http: request body too large"
		// 之类的错误；用 errors.As 抓 *http.MaxBytesError，更可靠。
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			response.PayloadTooLarge(c, "上传文件过大")
			return
		}
		response.BadRequest(c, "缺少上传文件字段 "+uploadFormField+"："+err.Error())
		return
	}

	src, err := fileHeader.Open()
	if err != nil {
		response.ServerError(c, "打开上传文件失败："+err.Error())
		return
	}
	defer src.Close()

	content, err := io.ReadAll(src)
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			response.PayloadTooLarge(c, "上传文件过大")
			return
		}
		response.ServerError(c, "读取上传文件失败："+err.Error())
		return
	}

	// 中间件保证此处 key_id 必然存在且 >0。
	keyID := c.GetInt(middleware.ContextKeyAPIKeyID)

	img, err := h.svc.Upload(c.Request.Context(), &model.UploadImageInput{
		Filename: fileHeader.Filename,
		Content:  content,
		KeyID:    &keyID,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmptyFile):
			response.BadRequest(c, "上传内容为空")
		case errors.Is(err, service.ErrFileTooLarge):
			response.PayloadTooLarge(c, "上传文件过大")
		case errors.Is(err, service.ErrUnsupportedMime):
			response.BadRequest(c, "不支持的图片类型")
		default:
			response.ServerError(c, "图片上传失败："+err.Error())
		}
		return
	}

	response.Success(c, img)
}
