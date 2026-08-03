package handler

import (
	"github.com/gin-gonic/gin"

	"hunjie/internal/config"
	"hunjie/internal/model"
	"hunjie/internal/pkg/response"
)

// Dict 全量枚举字典。前端启动时拉一次缓存到本地
// GET /api/v1/common/dict
func Dict(c *gin.Context) {
	response.OK(c, model.Dict)
}

// ClientConfig 客户端配置
// GET /api/v1/common/config
func ClientConfig(c *gin.Context) {
	response.OK(c, gin.H{
		// mock 模式下前端手机号走输入框，真实模式走微信授权按钮
		"mock_mode":     config.C.WeChat.Mock,
		"notice":        "",
		"max_photo_num": 9,
	})
}

// Health 健康检查
func Health(c *gin.Context) {
	response.OK(c, gin.H{"status": "ok"})
}
