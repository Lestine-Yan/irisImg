package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Body 是统一响应体结构。
type Body struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// 业务状态码，与 HTTP 状态码区分开。
const (
	CodeOK           = 0
	CodeBadRequest   = 40000
	CodeUnauthorized = 40100
	// CodeAPIKeyMissing 表示请求缺少 API 密钥（header 缺失或为空）。
	CodeAPIKeyMissing = 40110
	// CodeAPIKeyInvalid 表示 API 密钥格式非法、不存在或已吊销。
	CodeAPIKeyInvalid = 40120
	CodeForbidden     = 40300
	CodeNotFound      = 40400
	// CodePayloadTooLarge 表示上传内容超过限制（如 max_upload_size_mb）。
	CodePayloadTooLarge = 41300
	// CodeTooManyRequests 表示触发限流。
	CodeTooManyRequests = 42900
	CodeServerError     = 50000
)

// Success 返回成功响应。
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Body{
		Code:    CodeOK,
		Message: "success",
		Data:    data,
	})
}

// Fail 返回失败响应。httpStatus 是 HTTP 层状态码，code 是业务层状态码。
func Fail(c *gin.Context, httpStatus, code int, msg string) {
	c.JSON(httpStatus, Body{
		Code:    code,
		Message: msg,
	})
}

// BadRequest 是 400 的快捷方法。
func BadRequest(c *gin.Context, msg string) {
	Fail(c, http.StatusBadRequest, CodeBadRequest, msg)
}

// NotFound 是 404 的快捷方法。
func NotFound(c *gin.Context, msg string) {
	Fail(c, http.StatusNotFound, CodeNotFound, msg)
}

// Unauthorized 是 401 的快捷方法。
func Unauthorized(c *gin.Context, msg string) {
	Fail(c, http.StatusUnauthorized, CodeUnauthorized, msg)
}

// Forbidden 是 403 的快捷方法（已认证但无权访问）。
func Forbidden(c *gin.Context, msg string) {
	Fail(c, http.StatusForbidden, CodeForbidden, msg)
}

// TooManyRequests 是 429 的快捷方法（触发限流）。
func TooManyRequests(c *gin.Context, msg string) {
	Fail(c, http.StatusTooManyRequests, CodeTooManyRequests, msg)
}

// PayloadTooLarge 是 413 的快捷方法（上传超限）。
func PayloadTooLarge(c *gin.Context, msg string) {
	Fail(c, http.StatusRequestEntityTooLarge, CodePayloadTooLarge, msg)
}

// ServerError 是 500 的快捷方法。
func ServerError(c *gin.Context, msg string) {
	Fail(c, http.StatusInternalServerError, CodeServerError, msg)
}
