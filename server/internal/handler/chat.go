package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"hunjie/internal/middleware"
	"hunjie/internal/pkg/response"
	"hunjie/internal/service"
	"hunjie/internal/ws"
)

// ListConversations 会话列表（= 配对列表）
// GET /api/v1/conversation/list
func ListConversations(c *gin.Context) {
	list, err := service.ListConversations(middleware.CurrentUser(c))
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, list)
}

// CloseConversation 解除配对
// DELETE /api/v1/conversation/:id
func CloseConversation(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	if err := service.CloseConversation(middleware.CurrentUser(c), id); err != nil {
		fail(c, err)
		return
	}
	response.OK(c, nil)
}

// ListMessages 历史消息
// GET /api/v1/message/list?conversation_id=1&before_id=0&limit=20
func ListMessages(c *gin.Context) {
	convID, _ := strconv.ParseUint(c.Query("conversation_id"), 10, 64)
	if convID == 0 {
		response.ParamError(c, "缺少 conversation_id")
		return
	}
	beforeID, _ := strconv.ParseUint(c.Query("before_id"), 10, 64)
	limit, _ := strconv.Atoi(c.Query("limit"))

	msgs, err := service.ListMessages(middleware.CurrentUser(c), convID, beforeID, limit)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, msgs)
}

// SendMessage HTTP 兜底发送。WebSocket 断开时前端走这里
// POST /api/v1/message/send
func SendMessage(c *gin.Context) {
	var req struct {
		ConversationID uint64 `json:"conversation_id" binding:"required"`
		MsgType        int8   `json:"msg_type"`
		Content        string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数错误")
		return
	}

	msg, err := service.SendMessage(middleware.CurrentUser(c), req.ConversationID, req.MsgType, req.Content)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, msg)
}

// RecallMessage 撤回消息（2 分钟内）
// PUT /api/v1/message/recall
func RecallMessage(c *gin.Context) {
	var req struct {
		MessageID uint64 `json:"message_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数错误")
		return
	}
	if err := service.RecallMessage(middleware.CurrentUser(c), req.MessageID); err != nil {
		fail(c, err)
		return
	}
	response.OK(c, nil)
}

// ReadMessages 标记会话已读
// PUT /api/v1/message/read
func ReadMessages(c *gin.Context) {
	var req struct {
		ConversationID uint64 `json:"conversation_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数错误")
		return
	}
	if err := service.MarkConvRead(middleware.CurrentUser(c), req.ConversationID); err != nil {
		fail(c, err)
		return
	}
	response.OK(c, nil)
}

// WebSocket 长连接
// GET /ws?token=<jwt>
func WebSocket(c *gin.Context) {
	ws.Serve(c, middleware.CurrentUserID(c))
}
