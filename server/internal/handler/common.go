package handler

import (
	"github.com/gin-gonic/gin"

	"hunjie/internal/config"
	"hunjie/internal/model"
	"hunjie/internal/pkg/response"
	"hunjie/internal/service"
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
		"max_photo_num": service.MaxPhotoNum,
		// 免费可见照片张数。前端用它渲染「还有 N 张」的解锁提示，
		// 与后端 model.FreePhotoLimit 是同一个口径来源
		"free_photo_limit": model.FreePhotoLimit,
		// 私聊聊天远程开关。false 时前端隐藏消息列表与所有聊天入口，
		// 后端同时拒绝会话/消息接口（见 middleware.RequirePrivateChat）
		"private_chat": config.C.Features.PrivateChat,
	})
}

// Health 健康检查
func Health(c *gin.Context) {
	response.OK(c, gin.H{"status": "ok"})
}
