package service

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"

	"hunjie/internal/db"
	"hunjie/internal/model"
	"hunjie/internal/pkg/logger"
	"hunjie/internal/pkg/wechat"
	"hunjie/internal/ws"
)

// InitWS 注入 WebSocket 业务回调。ws 包不 import service，避免循环依赖
func InitWS() {
	ws.OnMessage = handleWSMessage
}

// ConversationItem 会话列表项
type ConversationItem struct {
	ID            uint64     `json:"id"`
	PeerID        uint64     `json:"peer_id"`
	PeerNickname  string     `json:"peer_nickname"`
	PeerAvatar    string     `json:"peer_avatar"`
	LastMessage   string     `json:"last_message"`
	LastMessageAt *time.Time `json:"last_message_at"`
	Unread        int        `json:"unread"`
	MatchedAt     *time.Time `json:"matched_at"`
	PeerOnline    bool       `json:"peer_online"`
}

// ListConversations 会话列表（= 配对列表）
func ListConversations(me *model.User) ([]ConversationItem, error) {
	var convs []model.Conversation
	if err := db.DB.
		Where("(user_a_id = ? OR user_b_id = ?) AND status = ?", me.ID, me.ID, model.ConvNormal).
		Order("last_message_at DESC, id DESC").
		Find(&convs).Error; err != nil {
		return nil, err
	}
	if len(convs) == 0 {
		return []ConversationItem{}, nil
	}

	peerIDs := make([]uint64, 0, len(convs))
	for _, c := range convs {
		peerIDs = append(peerIDs, c.PeerID(me.ID))
	}
	var users []model.User
	db.DB.Where("id IN ?", peerIDs).Find(&users)
	userMap := make(map[uint64]model.User, len(users))
	for _, u := range users {
		userMap[u.ID] = u
	}

	items := make([]ConversationItem, 0, len(convs))
	for _, c := range convs {
		peerID := c.PeerID(me.ID)
		peer := userMap[peerID]
		items = append(items, ConversationItem{
			ID:            c.ID,
			PeerID:        peerID,
			PeerNickname:  peer.Nickname,
			PeerAvatar:    peer.Avatar,
			LastMessage:   c.LastMessage,
			LastMessageAt: c.LastMessageAt,
			Unread:        c.UnreadFor(me.ID),
			MatchedAt:     c.MatchedAt,
			PeerOnline:    ws.IsOnline(peerID),
		})
	}
	return items, nil
}

// ListMessages 历史消息。beforeID 用于向上翻页
func ListMessages(me *model.User, convID, beforeID uint64, limit int) ([]model.Message, error) {
	if _, err := checkConversation(me, convID); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	q := db.DB.Where("conversation_id = ?", convID)
	if beforeID > 0 {
		q = q.Where("id < ?", beforeID)
	}

	var msgs []model.Message
	if err := q.Order("id DESC").Limit(limit).Find(&msgs).Error; err != nil {
		return nil, err
	}
	// 查出来是倒序，反转成正序方便前端直接渲染
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, nil
}

// SendMessage 发送消息。WebSocket 与 HTTP 兜底走的是同一个方法
func SendMessage(me *model.User, convID uint64, msgType int8, content string) (*model.Message, error) {
	conv, err := checkConversation(me, convID)
	if err != nil {
		return nil, err
	}
	if content == "" {
		return nil, Biz("消息不能为空")
	}
	if len([]rune(content)) > 500 {
		return nil, Biz("单条消息不能超过 500 字")
	}
	if msgType != model.MsgText && msgType != model.MsgImage {
		msgType = model.MsgText
	}

	if msgType == model.MsgText {
		if ok, reason := wechat.C.MsgSecCheck(me.OpenID, content); !ok {
			return nil, Biz("%s", reason)
		}
	}

	peerID := conv.PeerID(me.ID)
	msg := model.Message{
		ConversationID: convID,
		FromUserID:     me.ID,
		Type:           msgType,
		Content:        content,
	}

	summary := content
	if msgType == model.MsgImage {
		summary = "[图片]"
	}
	if len([]rune(summary)) > 50 {
		summary = string([]rune(summary)[:50])
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&msg).Error; err != nil {
			return err
		}
		unreadCol := "b_unread"
		if conv.UserAID == peerID {
			unreadCol = "a_unread"
		}
		return tx.Model(&model.Conversation{}).Where("id = ?", convID).
			Updates(map[string]any{
				"last_message":    summary,
				"last_message_at": time.Now(),
				unreadCol:         gorm.Expr(unreadCol + " + 1"),
			}).Error
	})
	if err != nil {
		return nil, err
	}

	// 实时推给对方，离线则等 ta 下次拉历史消息
	ws.Push(peerID, "message", msg)
	return &msg, nil
}

// RecallMessage 撤回消息。仅限本人、2 分钟内
func RecallMessage(me *model.User, msgID uint64) error {
	var msg model.Message
	if err := db.DB.First(&msg, msgID).Error; err != nil {
		return Biz("消息不存在")
	}
	if msg.FromUserID != me.ID {
		return Biz("只能撤回自己发的消息")
	}
	if msg.Type == model.MsgSystem {
		return Biz("该消息不能撤回")
	}
	if time.Since(msg.CreatedAt) > 2*time.Minute {
		return Biz("发出超过 2 分钟的消息不能撤回")
	}

	conv, err := checkConversation(me, msg.ConversationID)
	if err != nil {
		return err
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		// 转成系统消息占位，前端按 from_user_id 显示「你/对方 撤回了一条消息」
		if err := tx.Model(&msg).Updates(map[string]any{
			"type":    model.MsgSystem,
			"content": "撤回了一条消息",
		}).Error; err != nil {
			return err
		}
		// 撤回的恰好是最后一条时，同步刷新会话摘要
		var latest model.Message
		if err := tx.Where("conversation_id = ?", conv.ID).
			Order("id DESC").First(&latest).Error; err == nil && latest.ID == msg.ID {
			return tx.Model(&model.Conversation{}).Where("id = ?", conv.ID).
				Update("last_message", "[消息已撤回]").Error
		}
		return nil
	})
	if err != nil {
		return err
	}

	ws.Push(conv.PeerID(me.ID), "recall", map[string]any{
		"conversation_id": conv.ID,
		"message_id":      msg.ID,
		"from_user_id":    me.ID,
	})
	return nil
}

// MarkConvRead 会话已读
func MarkConvRead(me *model.User, convID uint64) error {
	conv, err := checkConversation(me, convID)
	if err != nil {
		return err
	}

	unreadCol := "a_unread"
	if conv.UserBID == me.ID {
		unreadCol = "b_unread"
	}

	return db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Conversation{}).Where("id = ?", convID).
			Update(unreadCol, 0).Error; err != nil {
			return err
		}
		return tx.Model(&model.Message{}).
			Where("conversation_id = ? AND from_user_id <> ? AND is_read = 0", convID, me.ID).
			Update("is_read", 1).Error
	})
}

// CloseConversation 解除配对
func CloseConversation(me *model.User, convID uint64) error {
	conv, err := checkConversation(me, convID)
	if err != nil {
		return err
	}
	peerID := conv.PeerID(me.ID)

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(conv).Update("status", model.ConvClosed).Error; err != nil {
			return err
		}
		// 双方的喜欢一并清掉，否则会立刻重新配对
		return tx.Where(
			"(from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?)",
			me.ID, peerID, peerID, me.ID).
			Where("type = ?", model.InteractionLike).
			Delete(&model.Interaction{}).Error
	})
	if err != nil {
		return err
	}

	ws.Push(peerID, "unmatch", map[string]any{"conversation_id": convID})
	return nil
}

// checkConversation 校验会话存在、我是参与方、且未关闭
func checkConversation(me *model.User, convID uint64) (*model.Conversation, error) {
	var conv model.Conversation
	if err := db.DB.First(&conv, convID).Error; err != nil {
		return nil, Biz("会话不存在")
	}
	if conv.UserAID != me.ID && conv.UserBID != me.ID {
		return nil, Biz("无权访问该会话")
	}
	if conv.Status != model.ConvNormal {
		return nil, Biz("该会话已结束")
	}
	return &conv, nil
}

// ==================== WebSocket 消息处理 ====================

func handleWSMessage(userID uint64, msgType string, data json.RawMessage) {
	var u model.User
	if err := db.DB.First(&u, userID).Error; err != nil {
		return
	}

	switch msgType {
	case "send":
		var req struct {
			ConversationID uint64 `json:"conversation_id"`
			MsgType        int8   `json:"msg_type"`
			Content        string `json:"content"`
			TempID         string `json:"temp_id"`
		}
		if err := json.Unmarshal(data, &req); err != nil {
			ws.Push(userID, "error", map[string]string{"message": "参数格式错误"})
			return
		}

		msg, err := SendMessage(&u, req.ConversationID, req.MsgType, req.Content)
		if err != nil {
			ws.Push(userID, "error", map[string]any{
				"message": err.Error(), "temp_id": req.TempID,
			})
			return
		}
		// 回执让前端把"发送中"的气泡换成真实消息
		ws.Push(userID, "ack", map[string]any{
			"temp_id": req.TempID, "message": msg,
		})

	case "read":
		var req struct {
			ConversationID uint64 `json:"conversation_id"`
		}
		if err := json.Unmarshal(data, &req); err != nil {
			return
		}
		if err := MarkConvRead(&u, req.ConversationID); err != nil {
			logger.Debugf("标记已读失败 uid=%d conv=%d: %v", userID, req.ConversationID, err)
		}

	case "typing":
		// 正在输入提示：只做转发，不落库
		var req struct {
			ConversationID uint64 `json:"conversation_id"`
		}
		if err := json.Unmarshal(data, &req); err != nil {
			return
		}
		conv, err := checkConversation(&u, req.ConversationID)
		if err != nil {
			return
		}
		ws.Push(conv.PeerID(u.ID), "typing", map[string]any{
			"conversation_id": conv.ID,
		})

	default:
		ws.Push(userID, "error", map[string]string{"message": "不支持的消息类型: " + msgType})
	}
}

// PushMatch 配对成功通知
func PushMatch(userID uint64, conv *model.Conversation, peer *model.User) {
	ws.Push(userID, "match", map[string]any{
		"conversation_id": conv.ID,
		"peer": map[string]any{
			"id": peer.ID, "nickname": peer.Nickname, "avatar": peer.Avatar,
		},
	})
}

// PushLike 有人喜欢我的通知
func PushLike(userID uint64, from *model.User) {
	ws.Push(userID, "like", map[string]any{
		"id": from.ID, "nickname": from.Nickname, "avatar": from.Avatar,
	})
}

