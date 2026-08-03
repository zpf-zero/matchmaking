package service

import (
	"errors"

	"gorm.io/gorm"

	"hunjie/internal/db"
	"hunjie/internal/dto"
	"hunjie/internal/model"
)

// CreateReport 举报
func CreateReport(me *model.User, req dto.ReportReq) error {
	if req.TargetUserID == me.ID {
		return Biz("不能举报自己")
	}
	if err := db.DB.First(&model.User{}, req.TargetUserID).Error; err != nil {
		return Biz("被举报的用户不存在")
	}
	if model.DictLabel("report_reason", int(req.Reason)) == "" {
		return Biz("请选择举报原因")
	}

	// 同一个人重复举报同一对象且未处理时，不重复建单
	var exists int64
	db.DB.Model(&model.Report{}).
		Where("from_user_id = ? AND target_user_id = ? AND status = ?",
			me.ID, req.TargetUserID, model.ReportPending).Count(&exists)
	if exists > 0 {
		return Biz("你已举报过该用户，我们正在处理")
	}

	return db.DB.Create(&model.Report{
		FromUserID:   me.ID,
		TargetUserID: req.TargetUserID,
		Reason:       req.Reason,
		Description:  req.Description,
		Images:       req.Images,
		Status:       model.ReportPending,
	}).Error
}

// AddBlacklist 拉黑。双向不可见，已有的会话一并关闭
func AddBlacklist(me *model.User, targetID uint64) error {
	if targetID == me.ID {
		return Biz("不能拉黑自己")
	}
	if err := db.DB.First(&model.User{}, targetID).Error; err != nil {
		return Biz("该用户不存在")
	}

	var exists model.Blacklist
	err := db.DB.Where("user_id = ? AND target_user_id = ?", me.ID, targetID).First(&exists).Error
	if err == nil {
		return nil // 已拉黑，幂等
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&model.Blacklist{
			UserID: me.ID, TargetUserID: targetID,
		}).Error; err != nil {
			return err
		}
		// 互相的喜欢一并清掉，否则对方还留在「谁喜欢我」里
		if err := tx.Where(
			"(from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?)",
			me.ID, targetID, targetID, me.ID).
			Delete(&model.Interaction{}).Error; err != nil {
			return err
		}
		lo, hi := orderPair(me.ID, targetID)
		return tx.Model(&model.Conversation{}).
			Where("user_a_id = ? AND user_b_id = ?", lo, hi).
			Update("status", model.ConvClosed).Error
	})
}

// RemoveBlacklist 取消拉黑
func RemoveBlacklist(me *model.User, targetID uint64) error {
	return db.DB.Where("user_id = ? AND target_user_id = ?", me.ID, targetID).
		Delete(&model.Blacklist{}).Error
}

// ListBlacklist 黑名单列表
func ListBlacklist(me *model.User) ([]map[string]any, error) {
	var list []model.Blacklist
	if err := db.DB.Where("user_id = ?", me.ID).Order("id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return []map[string]any{}, nil
	}

	ids := make([]uint64, 0, len(list))
	for _, b := range list {
		ids = append(ids, b.TargetUserID)
	}
	var users []model.User
	db.DB.Where("id IN ?", ids).Find(&users)
	userMap := make(map[uint64]model.User, len(users))
	for _, u := range users {
		userMap[u.ID] = u
	}

	out := make([]map[string]any, 0, len(list))
	for _, b := range list {
		u := userMap[b.TargetUserID]
		out = append(out, map[string]any{
			"user_id":  b.TargetUserID,
			"nickname": u.Nickname,
			"avatar":   u.Avatar,
		})
	}
	return out, nil
}
