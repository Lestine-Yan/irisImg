package api

import (
	"net/http"

	"github.com/Lestine-Yan/irisImg/backend/internal/dao"
	"github.com/Lestine-Yan/irisImg/backend/internal/middleware"
	"github.com/Lestine-Yan/irisImg/backend/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

// ImageAPI 是图片相关接口的控制器。
//
// 当前为占位实现：真实的「申请图片(GET) / 添加图片(POST)」业务逻辑后续单独实现。
// 这里保留占位处理函数，主要用于挂载并演示 API 密钥鉴权中间件：
// 只读密钥可访问 GET，读写密钥才能 POST，超出限流会被拒。
type ImageAPI struct {
	imageDAO dao.ImageDAO
}

// NewImageAPI 构造控制器。
func NewImageAPI(imageDAO dao.ImageDAO) *ImageAPI {
	return &ImageAPI{imageDAO: imageDAO}
}

// List 处理 GET /images（占位）。任意有效密钥均可访问。
func (h *ImageAPI) List(c *gin.Context) {
	keyID := c.GetInt(middleware.ContextKeyAPIKeyID)
	response.Fail(c, http.StatusNotImplemented, response.CodeServerError, "图片列表接口尚未实现（占位）")
	_ = keyID
}

// Create 处理 POST /images（占位）。需读写密钥。
// 真实实现时应将 middleware.ContextKeyAPIKeyID 写入 model.Image.KeyID 落库，
// 以记录图片由哪个密钥添加。
func (h *ImageAPI) Create(c *gin.Context) {
	keyID := c.GetInt(middleware.ContextKeyAPIKeyID)
	response.Fail(c, http.StatusNotImplemented, response.CodeServerError, "图片添加接口尚未实现（占位）")
	_ = keyID
}
