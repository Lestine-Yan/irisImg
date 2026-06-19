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
	CodeNotFound     = 40400
	CodeServerError  = 50000
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

// ServerError 是 500 的快捷方法。
func ServerError(c *gin.Context, msg string) {
	Fail(c, http.StatusInternalServerError, CodeServerError, msg)
}
