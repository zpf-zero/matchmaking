package service

import (
	"errors"
	"sync"
	"time"

	"gorm.io/gorm"

	"hunjie/internal/db"
	"hunjie/internal/dto"
	"hunjie/internal/model"
)

// LikeResult 喜欢的结果
type LikeResult struct {
	IsMatch        bool   `json:"is_match"`
	ConversationID uint64 `json:"conversation_id"`
	Peer           any    `json:"peer,omitempty"` // 配对成功时带上对方简介，前端直接弹窗
}

// Like 喜欢某人。互相喜欢即配对，直接创建会话
// 操作者自身的状态门槛由路由层的 RequireActive 统一把关
func Like(me *model.User, toID uint64) (*LikeResult, error) {
	target, err := checkTarget(me, toID)
	if err != nil {
		return nil, err
	}

	if err := upsertInteraction(me.ID, toID, model.InteractionLike); err != nil {
		return nil, err
	}

	// 对方是否也喜欢我
	var mutual int64
	db.DB.Model(&model.Interaction{}).
		Where("from_user_id = ? AND to_user_id = ? AND type = ?", toID, me.ID, model.InteractionLike).
		Count(&mutual)

	if mutual == 0 {
		PushLike(toID, me)
		return &LikeResult{IsMatch: false}, nil
	}

	conv, err := ensureConversation(me.ID, toID)
	if err != nil {
		return nil, err
	}
	// 只推给被动方：发起方已经从 HTTP 响应里拿到 is_match 并弹窗，
	// 再推一次会导致前端全局通知重复弹
	PushMatch(toID, conv, me)

	return &LikeResult{
		IsMatch:        true,
		ConversationID: conv.ID,
		Peer: map[string]any{
			"id": target.ID, "nickname": target.Nickname, "avatar": target.Avatar,
		},
	}, nil
}

// Pass 不喜欢。之后不再出现在推荐流里
// 操作者自身的状态门槛由路由层的 RequireActive 统一把关
func Pass(me *model.User, toID uint64) error {
	if toID == me.ID {
		return Biz("不能对自己操作")
	}
	// 先撤掉可能存在的喜欢，避免同时存在 like 和 pass
	db.DB.Where("from_user_id = ? AND to_user_id = ? AND type = ?",
		me.ID, toID, model.InteractionLike).Delete(&model.Interaction{})
	return upsertInteraction(me.ID, toID, model.InteractionPass)
}

// Unlike 取消喜欢。已配对的会话一并关闭
func Unlike(me *model.User, toID uint64) error {
	if err := db.DB.Where("from_user_id = ? AND to_user_id = ? AND type = ?",
		me.ID, toID, model.InteractionLike).Delete(&model.Interaction{}).Error; err != nil {
		return err
	}
	if conv := findConversation(me.ID, toID); conv != nil && conv.Status == model.ConvNormal {
		return db.DB.Model(conv).Update("status", model.ConvClosed).Error
	}
	return nil
}

// ListLikedMe 谁喜欢我
func ListLikedMe(me *model.User, page dto.PageQuery) ([]dto.UserCard, int64, error) {
	return listByInteraction(me, page, model.InteractionLike, "to_user_id", "from_user_id")
}

// ListMyLikes 我喜欢谁
func ListMyLikes(me *model.User, page dto.PageQuery) ([]dto.UserCard, int64, error) {
	return listByInteraction(me, page, model.InteractionLike, "from_user_id", "to_user_id")
}

// ViewedMeLimit 免费用户可见的「谁看过我」条数。
// 权益点设计沿用竞品规律：把「别人对我的行为」做成可售卖的信息可见性
// （珍爱/世纪佳缘的「谁看过我」都是会员权益），成本为零、用户感知强烈。
// 这里刻意只锁「看更多」，不锁任何功能入口 —— 免费用户依然能正常浏览、喜欢、聊天。
const ViewedMeLimit = 3

// ViewedMeResp 「谁看过我」列表。非权益用户只返回前若干条
type ViewedMeResp struct {
	List      []dto.UserCard `json:"list"`
	Total     int64          `json:"total"`    // 实际看过我的人数
	Locked    int64          `json:"locked"`   // 被锁住、需要权益才能看的人数
	Unlocked  bool           `json:"unlocked"` // 我是否已解锁完整资料
	FreeLimit int            `json:"free_limit"`
}

// ListViewedMe 谁看过我。非权益用户只返回最近若干条，其余需通过邀请获得权益后解锁
func ListViewedMe(me *model.User, page dto.PageQuery) (*ViewedMeResp, error) {
	page.Normalize()

	vip := me.HasVip()

	// 总人数与列表口径必须一致：都基于 baseQuery 的可见性规则
	// （对方已封禁/资料下架/拉黑后就不该再计入，否则会出现「锁了但解不开」的幻影）
	sub := db.DB.Table("interaction").Select("from_user_id").
		Where("to_user_id = ? AND type = ?", me.ID, model.InteractionView)

	base := baseQuery(me).Where("u.id IN (?)", sub)

	var total int64
	if err := base.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, err
	}
	limit := page.PageSize
	locked := int64(0)
	if !vip {
		// 免费用户：只看第一页的前 N 条。翻页一律返回空，避免绕过限制
		if page.Page > 1 {
			return &ViewedMeResp{
				List: []dto.UserCard{}, Total: total,
				Locked: maxInt64(total-ViewedMeLimit, 0), Unlocked: false, FreeLimit: ViewedMeLimit,
			}, nil
		}
		limit = ViewedMeLimit
		locked = maxInt64(total-ViewedMeLimit, 0)
	}

	// 总数上面已经算过，这里只取当前页，不再重复 COUNT
	cards, err := runCardQueryPage(me, base, dto.PageQuery{Page: page.Page, PageSize: limit})
	if err != nil {
		return nil, err
	}
	return &ViewedMeResp{
		List: cards, Total: total, Locked: locked,
		Unlocked: vip, FreeLimit: ViewedMeLimit,
	}, nil
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// listByInteraction 按互动记录反查用户列表。
// meCol 是我在互动记录里的位置，peerCol 是对方的位置
func listByInteraction(me *model.User, page dto.PageQuery,
	actType int, meCol, peerCol string) ([]dto.UserCard, int64, error) {

	sub := db.DB.Table("interaction").Select(peerCol).
		Where(meCol+" = ? AND type = ?", me.ID, actType)

	q := baseQuery(me).Where("u.id IN (?)", sub)
	return runCardQuery(me, q, page)
}

// MarkRead 标记某类互动为已读（清红点）
func MarkRead(me *model.User, actType int) error {
	return db.DB.Model(&model.Interaction{}).
		Where("to_user_id = ? AND type = ? AND is_read = 0", me.ID, actType).
		Update("is_read", 1).Error
}

// UnreadCount 各类未读数。
// 三个 COUNT 互不依赖，并发执行：这个接口每次切 tab/回前台都会调，
// 串行跑在远程库上要 150ms+，会直接变成「切换不跟手」的体感。
func UnreadCount(me *model.User) map[string]int64 {
	var likedMe, viewedMe, msg int64

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		db.DB.Model(&model.Interaction{}).
			Where("to_user_id = ? AND type = ? AND is_read = 0", me.ID, model.InteractionLike).
			Count(&likedMe)
	}()
	go func() {
		defer wg.Done()
		db.DB.Model(&model.Interaction{}).
			Where("to_user_id = ? AND type = ? AND is_read = 0", me.ID, model.InteractionView).
			Count(&viewedMe)
	}()
	go func() {
		defer wg.Done()
		db.DB.Model(&model.Conversation{}).
			Select("COALESCE(SUM(CASE WHEN user_a_id = ? THEN a_unread ELSE b_unread END), 0)", me.ID).
			Where("(user_a_id = ? OR user_b_id = ?) AND status = ?", me.ID, me.ID, model.ConvNormal).
			Scan(&msg)
	}()
	wg.Wait()

	return map[string]int64{
		"liked_me":  likedMe,
		"viewed_me": viewedMe,
		"message":   msg,
		"total":     likedMe + viewedMe + msg,
	}
}

// ==================== 内部 ====================

// checkTarget 校验互动对象是否可操作
func checkTarget(me *model.User, toID uint64) (*model.User, error) {
	if toID == me.ID {
		return nil, Biz("不能对自己操作")
	}

	var target model.User
	if err := db.DB.First(&target, toID).Error; err != nil {
		return nil, Biz("该用户不存在")
	}
	if target.Status != model.StatusNormal {
		return nil, Biz("该用户暂时无法互动")
	}
	if target.Gender == me.Gender {
		return nil, Biz("暂不支持同性匹配")
	}

	var blocked int64
	db.DB.Model(&model.Blacklist{}).
		Where("(user_id = ? AND target_user_id = ?) OR (user_id = ? AND target_user_id = ?)",
			me.ID, toID, toID, me.ID).Count(&blocked)
	if blocked > 0 {
		return nil, Biz("你与该用户之间已拉黑，无法互动")
	}

	return &target, nil
}

// upsertInteraction 写入互动记录。同类型只保留一条，重复操作只更新时间
func upsertInteraction(fromID, toID uint64, actType int) error {
	var it model.Interaction
	err := db.DB.Where("from_user_id = ? AND to_user_id = ? AND type = ?",
		fromID, toID, actType).First(&it).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return db.DB.Create(&model.Interaction{
			FromUserID: fromID, ToUserID: toID, Type: int8(actType),
		}).Error
	}
	if err != nil {
		return err
	}
	return db.DB.Model(&it).UpdateColumn("updated_at", time.Now()).Error
}

// ensureConversation 取或建会话。已解除的会话重新配对时会被激活
func ensureConversation(a, b uint64) (*model.Conversation, error) {
	lo, hi := orderPair(a, b)

	var conv model.Conversation
	err := db.DB.Where("user_a_id = ? AND user_b_id = ?", lo, hi).First(&conv).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		now := time.Now()
		conv = model.Conversation{
			UserAID: lo, UserBID: hi,
			Status:    model.ConvNormal,
			MatchedAt: &now,
		}
		if err := db.DB.Create(&conv).Error; err != nil {
			return nil, err
		}
		return &conv, nil
	}
	if err != nil {
		return nil, err
	}

	if conv.Status != model.ConvNormal {
		now := time.Now()
		if err := db.DB.Model(&conv).Updates(map[string]any{
			"status": model.ConvNormal, "matched_at": now,
		}).Error; err != nil {
			return nil, err
		}
		conv.Status = model.ConvNormal
	}
	return &conv, nil
}
