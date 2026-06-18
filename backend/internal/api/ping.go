package api

import (
	"github.com/Lestine-Yan/irisImg/backend/config"
	"github.com/Lestine-Yan/irisImg/backend/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

// Ping 是最简单的健康检查接口。
func Ping(c *gin.Context) {
	data := gin.H{"pong": true}
	if config.Global != nil {
		data["app"] = config.Global.App.Name
		data["version"] = config.Global.App.Version
	}
	response.Success(c, data)
}
