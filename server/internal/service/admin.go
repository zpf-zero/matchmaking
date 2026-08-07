package service

import (
	"time"

	"gorm.io/gorm"

	"hunjie/internal/db"
	"hunjie/internal/dto"
	"hunjie/internal/model"
	"hunjie/internal/pkg/logger"
)

// Overview 数据概览。小项目直接实时 count，不做统计表和定时任务
func Overview() map[string]int64 {
	var totalUser, todayNew, pendingAudit, pendingReport, matched, activeWeek int64
	today := time.Now().Format("2006-01-02")

	db.DB.Model(&model.User{}).Where("status <> ?", model.StatusIncomplete).Count(&totalUser)
	db.DB.Model(&model.User{}).Where("created_at >= ?", today).Count(&todayNew)
	db.DB.Table("user_profile p").
		Joins("JOIN user u ON u.id = p.user_id AND u.deleted_at IS NULL").
		Where("p.deleted_at IS NULL AND p.audit_status = ?", model.AuditPending).
		Where("u.status IN ?", []int{model.StatusAuditing, model.StatusNormal}).
		Count(&pendingAudit)
	db.DB.Model(&model.Report{}).Where("status = ?", model.ReportPending).Count(&pendingReport)
	db.DB.Model(&model.Conversation{}).Where("status = ?", model.ConvNormal).Count(&matched)
	db.DB.Model(&model.User{}).
		Where("last_active_at >= ?", time.Now().AddDate(0, 0, -7)).Count(&activeWeek)

	return map[string]int64{
		"total_user":     totalUser,
		"today_new":      todayNew,
		"pending_audit":  pendingAudit,
		"pending_report": pendingReport,
		"matched":        matched,
		"active_week":    activeWeek,
	}
}

// AuditItem 待审条目。资料 + 头像 + 相册一次返回，审核员一屏审完
type AuditItem struct {
	UserID    uint64             `json:"user_id"`
	Nickname  string             `json:"nickname"`
	Avatar    string             `json:"avatar"`
	Gender    int8               `json:"gender"`
	Phone     string             `json:"phone"`
	Status    int8               `json:"status"`
	CreatedAt time.Time          `json:"created_at"`
	Profile   *model.UserProfile `json:"profile"`
	Photos    []model.UserPhoto  `json:"photos"`
}

// AuditList 待审队列
func AuditList(page dto.PageQuery) ([]AuditItem, int64, error) {
	page.Normalize()

	q := db.DB.Table("user_profile p").
		Joins("JOIN user u ON u.id = p.user_id AND u.deleted_at IS NULL").
		Where("p.deleted_at IS NULL AND p.audit_status = ?", model.AuditPending).
		Where("u.status IN ?", []int{model.StatusAuditing, model.StatusNormal})

	var total int64
	if err := q.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []AuditItem{}, 0, nil
	}

	var profiles []model.UserProfile
	if err := q.Session(&gorm.Session{}).
		Select("p.*").
		Order("p.updated_at asc"). // 先提交的先审
		Limit(page.PageSize).Offset(page.Offset()).
		Scan(&profiles).Error; err != nil {
		return nil, 0, err
	}

	userIDs := make([]uint64, 0, len(profiles))
	for _, p := range profiles {
		userIDs = append(userIDs, p.UserID)
	}

	var users []model.User
	db.DB.Where("id IN ?", userIDs).Find(&users)
	userMap := make(map[uint64]model.User, len(users))
	for _, u := range users {
		userMap[u.ID] = u
	}

	var photos []model.UserPhoto
	db.DB.Where("user_id IN ?", userIDs).Order("sort asc, id asc").Find(&photos)
	photoMap := make(map[uint64][]model.UserPhoto)
	for _, p := range photos {
		photoMap[p.UserID] = append(photoMap[p.UserID], p)
	}

	items := make([]AuditItem, 0, len(profiles))
	for i := range profiles {
		p := profiles[i]
		u := userMap[p.UserID]
		items = append(items, AuditItem{
			UserID:    p.UserID,
			Nickname:  u.Nickname,
			Avatar:    u.Avatar,
			Gender:    u.Gender,
			Phone:     u.Phone,
			Status:    u.Status,
			CreatedAt: u.CreatedAt,
			Profile:   &p,
			Photos:    photoMap[p.UserID],
		})
	}
	return items, total, nil
}

// AuditUser 审核用户资料。资料与照片一起处理
func AuditUser(admin *model.User, userID uint64, pass bool, remark string) error {
	if !pass && remark == "" {
		return Biz("拒绝时请填写原因，用户需要知道改哪里")
	}

	var u model.User
	if err := db.DB.First(&u, userID).Error; err != nil {
		return Biz("该用户不存在")
	}
	if u.Status == model.StatusIncomplete {
		return Biz("该用户尚未正式提交审核")
	}

	var p model.UserProfile
	if err := db.DB.Where("user_id = ?", userID).First(&p).Error; err != nil {
		return Biz("该用户尚未填写资料")
	}

	now := time.Now()
	auditStatus := model.AuditRejected
	userStatus := model.StatusRejected
	if pass {
		auditStatus = model.AuditPassed
		userStatus = model.StatusNormal
	}

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&p).Updates(map[string]any{
			"audit_status": auditStatus,
			"audit_remark": remark,
			"audited_at":   now,
		}).Error; err != nil {
			return err
		}
		// 相册跟随资料一起过审
		if err := tx.Model(&model.UserPhoto{}).
			Where("user_id = ? AND audit_status = ?", userID, model.AuditPending).
			Update("audit_status", auditStatus).Error; err != nil {
			return err
		}
		// 封禁用户的状态不受审核影响，需要先解封
		if u.Status != model.StatusBanned {
			return tx.Model(&u).Update("status", userStatus).Error
		}
		return nil
	})
	if err != nil {
		return err
	}

	logger.Infof("审核: admin=%d user=%d pass=%v remark=%s", admin.ID, userID, pass, remark)
	return nil
}

// ==================== 用户管理 ====================

// AdminUserList 用户列表。关键词匹配昵称/手机号/id
func AdminUserList(keyword string, status int, page dto.PageQuery) ([]dto.AdminUserItem, int64, error) {
	page.Normalize()

	q := db.DB.Table("user u").
		Joins("LEFT JOIN user_profile p ON p.user_id = u.id AND p.deleted_at IS NULL").
		Where("u.deleted_at IS NULL")

	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("u.nickname LIKE ? OR u.phone LIKE ? OR u.id = ?", like, like, keyword)
	}
	if status >= 0 {
		q = q.Where("u.status = ?", status)
	}

	var total int64
	if err := q.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []dto.AdminUserItem
	if err := q.Session(&gorm.Session{}).
		Select(`u.id, u.nickname, u.avatar, u.phone, u.gender, u.role, u.status,
			u.ban_reason, u.created_at, u.last_active_at,
			p.age, p.city, p.education, p.income_level, p.audit_status`).
		Order("u.id DESC").
		Limit(page.PageSize).Offset(page.Offset()).
		Scan(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// AdminUserDetail 用户详情。含资料、相册、举报记录
func AdminUserDetail(userID uint64) (map[string]any, error) {
	var u model.User
	if err := db.DB.First(&u, userID).Error; err != nil {
		return nil, Biz("该用户不存在")
	}

	var p model.UserProfile
	db.DB.Where("user_id = ?", userID).First(&p)

	var r model.UserRequirement
	db.DB.Where("user_id = ?", userID).First(&r)

	var photos []model.UserPhoto
	db.DB.Where("user_id = ?", userID).Order("sort asc").Find(&photos)

	var reports []model.Report
	db.DB.Where("target_user_id = ?", userID).Order("id DESC").Limit(20).Find(&reports)

	var likeCount, likedCount, matchCount int64
	db.DB.Model(&model.Interaction{}).
		Where("from_user_id = ? AND type = ?", userID, model.InteractionLike).Count(&likeCount)
	db.DB.Model(&model.Interaction{}).
		Where("to_user_id = ? AND type = ?", userID, model.InteractionLike).Count(&likedCount)
	db.DB.Model(&model.Conversation{}).
		Where("(user_a_id = ? OR user_b_id = ?) AND status = ?",
			userID, userID, model.ConvNormal).Count(&matchCount)

	return map[string]any{
		"user":        u,
		"phone":       u.Phone,
		"profile":     p,
		"requirement": r,
		"photos":      photos,
		"reports":     reports,
		"stats": map[string]int64{
			"like_count": likeCount, "liked_count": likedCount, "match_count": matchCount,
		},
	}, nil
}

// BanUser 封禁
func BanUser(admin *model.User, userID uint64, reason string) error {
	if userID == admin.ID {
		return Biz("不能封禁自己")
	}
	var u model.User
	if err := db.DB.First(&u, userID).Error; err != nil {
		return Biz("该用户不存在")
	}
	if u.Role >= admin.Role {
		return Biz("不能封禁同级或更高权限的账号")
	}
	if reason == "" {
		return Biz("请填写封禁原因")
	}

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&u).Updates(map[string]any{
			"status": model.StatusBanned, "ban_reason": reason,
		}).Error; err != nil {
			return err
		}
		// 关闭该用户的所有会话，避免继续骚扰
		return tx.Model(&model.Conversation{}).
			Where("user_a_id = ? OR user_b_id = ?", userID, userID).
			Update("status", model.ConvClosed).Error
	})
	if err != nil {
		return err
	}
	logger.Infof("封禁: admin=%d user=%d reason=%s", admin.ID, userID, reason)
	return nil
}

// UnbanUser 解封。恢复到审核通过前的状态，需重新过审
func UnbanUser(admin *model.User, userID uint64) error {
	var u model.User
	if err := db.DB.First(&u, userID).Error; err != nil {
		return Biz("该用户不存在")
	}
	if u.Status != model.StatusBanned {
		return Biz("该用户未被封禁")
	}

	var p model.UserProfile
	newStatus := model.StatusIncomplete
	if err := db.DB.Where("user_id = ?", userID).First(&p).Error; err == nil {
		if p.AuditStatus == model.AuditPassed {
			newStatus = model.StatusNormal
		} else {
			newStatus = model.StatusAuditing
		}
	}

	if err := db.DB.Model(&u).Updates(map[string]any{
		"status": newStatus, "ban_reason": "",
	}).Error; err != nil {
		return err
	}
	logger.Infof("解封: admin=%d user=%d", admin.ID, userID)
	return nil
}

// ==================== 举报处理 ====================

func AdminReportList(status int, page dto.PageQuery) ([]map[string]any, int64, error) {
	page.Normalize()

	q := db.DB.Model(&model.Report{})
	if status >= 0 {
		q = q.Where("status = ?", status)
	}

	var total int64
	if err := q.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var reports []model.Report
	if err := q.Session(&gorm.Session{}).
		Order("status asc, id desc"). // 待处理的排前面
		Limit(page.PageSize).Offset(page.Offset()).
		Find(&reports).Error; err != nil {
		return nil, 0, err
	}

	// 批量补上举报人与被举报人的昵称
	idSet := make(map[uint64]bool)
	for _, r := range reports {
		idSet[r.FromUserID] = true
		idSet[r.TargetUserID] = true
	}
	ids := make([]uint64, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}
	var users []model.User
	db.DB.Unscoped().Where("id IN ?", ids).Find(&users)
	nameMap := make(map[uint64]string, len(users))
	for _, u := range users {
		nameMap[u.ID] = u.Nickname
	}

	list := make([]map[string]any, 0, len(reports))
	for _, r := range reports {
		list = append(list, map[string]any{
			"report":          r,
			"reason_text":     model.DictLabel("report_reason", int(r.Reason)),
			"from_nickname":   nameMap[r.FromUserID],
			"target_nickname": nameMap[r.TargetUserID],
		})
	}
	return list, total, nil
}

// HandleReport 处理举报。action: ignore / ban
func HandleReport(admin *model.User, reportID uint64, action, remark string) error {
	var r model.Report
	if err := db.DB.First(&r, reportID).Error; err != nil {
		return Biz("举报记录不存在")
	}
	if r.Status != model.ReportPending {
		return Biz("该举报已处理过")
	}

	switch action {
	case "ignore":
		return db.DB.Model(&r).Updates(map[string]any{
			"status": model.ReportRejected, "handle_remark": remark,
		}).Error

	case "ban":
		// 封禁属于管理员权限，审核员只能忽略或上报
		if admin.Role < model.RoleAdmin {
			return Biz("封禁用户需要管理员权限")
		}
		if remark == "" {
			remark = "违反社区规范"
		}
		if err := BanUser(admin, r.TargetUserID, remark); err != nil {
			return err
		}
		return db.DB.Model(&r).Updates(map[string]any{
			"status": model.ReportHandled, "handle_remark": remark,
		}).Error

	default:
		return Biz("action 只能是 ignore 或 ban")
	}
}
