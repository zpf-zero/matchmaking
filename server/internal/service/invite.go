package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"

	"hunjie/internal/db"
	"hunjie/internal/model"
	"hunjie/internal/pkg/logger"
)

/*
 * 邀请裂变：邀请码 + 深度互动结算
 *
 * 合规要点（详见 docs/06-微信登录与分享裂变调研.md 第四、六章）：
 *
 * 《微信小程序平台运营规范》5.1.1 把「分享后才能解锁功能」明文列为强制分享违规，
 * 5.1.4 禁止「分享后无深度互动即得利益」。所以本模块刻意不做下面这些事：
 *   - 不检测分享是否成功（微信 2024-07-05 起已不返回该结果，做了也不可信）
 *   - 不要求用户先分享才能用功能（无任何功能门禁）
 *   - 不做「分享给好友立得奖励」这类无互动获利
 *   - 不在分享按钮上加红点或做动态形变
 *
 * 唯一形态是：邀请新用户 → 完成注册 + 完善资料 + 通过人工审核 → 双方各得权益。
 * 被邀请人走完这段真实业务流程，正是规范定义的「深度互动」。
 *
 * ============================ 本模块涉及的待申请事项 ============================
 * 【待申请】活动规则页必须在微信提审时一并提交，且内容需与代码实际逻辑一致
 *   —— docs/07-待申请事项清单.md 第 1 项（类目资质）的配套材料
 *   规则页：miniapp/src/pages/invite/rules.vue
 *   注意：规则页里写的「上限 20 人」与常量 InviteMaxSettle 必须始终一致，
 *   改常量时务必同步改规则页，否则构成「规则与实现不符」（5.1.5 误导用户）
 * ============================================================================
 */

// ==================== 我的邀请概览 ====================

// MyInviteInfo 我的邀请页数据
//
// 注意：这里刻意不出现任何「剩余天数」「到期时间」字段。
// 权益自 2026-10 起改为一次性永久解锁（VipLevel），前端展示的是
// 「是否已解锁完整资料」，不是一个会递减的倒计时（见 docs/08-权益体系重构方案.md 4.4）。
type MyInviteInfo struct {
	Code          string         `json:"code"`
	InviteURL     string         `json:"invite_url"`     // 前端拼分享路径用
	SettledCount  int64          `json:"settled_count"`  // 已成功邀请人数
	PendingCount  int64          `json:"pending_count"`  // 已绑定待达成人数
	MaxSettle     int            `json:"max_settle"`     // 结算上限（防刷）
	Unlocked      bool           `json:"unlocked"`       // 我是否已解锁完整资料
	UnseenSettled int64          `json:"unseen_settled"` // 已结算但我尚未查看，红点/弹层用
	Records       []InviteRecord `json:"records"`
}

// InviteRecord 邀请记录（被邀请人昵称脱敏）
type InviteRecord struct {
	Nickname  string     `json:"nickname"`
	Status    int8       `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	SettledAt *time.Time `json:"settled_at"`
}

// MyInvite 取（或在首次访问时惰性生成）我的邀请码，并返回邀请进度。
// 邀请码、三个计数、关系列表、权益状态互不依赖，并发查：
// 串行在远程库上要 400ms+，邀请页会白等。
func MyInvite(me *model.User) (*MyInviteInfo, error) {
	var (
		code             string
		settled, pending int64
		unseen           int64
		relations        []model.InviteRelation
		u                model.User
		codeErr          error
	)
	var wg sync.WaitGroup
	wg.Add(6)
	go func() { defer wg.Done(); code, codeErr = ensureInviteCode(me.ID) }()
	go func() {
		defer wg.Done()
		db.DB.Model(&model.InviteRelation{}).
			Where("inviter_id = ? AND status = ?", me.ID, model.InviteSettled).Count(&settled)
	}()
	go func() {
		defer wg.Done()
		db.DB.Model(&model.InviteRelation{}).
			Where("inviter_id = ? AND status = ?", me.ID, model.InvitePending).Count(&pending)
	}()
	go func() {
		defer wg.Done()
		// 结算发生在审核通过的那一刻，邀请人多半不在场。
		// unseen 是「有新结算要看」的信号，驱动「我的」页红点与邀请页一次性提示
		db.DB.Model(&model.InviteRelation{}).
			Where("inviter_id = ? AND status = ? AND inviter_seen = ?", me.ID, model.InviteSettled, false).
			Count(&unseen)
	}()
	go func() {
		defer wg.Done()
		db.DB.Where("inviter_id = ?", me.ID).Order("id DESC").Limit(50).Find(&relations)
	}()
	go func() {
		defer wg.Done()
		// 权益状态取最新实况：可能刚被结算发放
		db.DB.Select("vip_level").First(&u, me.ID)
	}()
	wg.Wait()

	if codeErr != nil {
		return nil, codeErr
	}

	infos := make([]InviteRecord, 0, len(relations))
	if len(relations) > 0 {
		ids := make([]uint64, 0, len(relations))
		for _, r := range relations {
			ids = append(ids, r.InviteeID)
		}
		var users []model.User
		db.DB.Unscoped().Where("id IN ?", ids).Find(&users)
		nameMap := make(map[uint64]string, len(users))
		for _, u := range users {
			nameMap[u.ID] = maskNickname(u.Nickname)
		}
		for _, r := range relations {
			infos = append(infos, InviteRecord{
				Nickname:  nameMap[r.InviteeID],
				Status:    r.Status,
				CreatedAt: r.CreatedAt,
				SettledAt: r.SettledAt,
			})
		}
	}

	return &MyInviteInfo{
		Code:          code,
		InviteURL:     "/pages/login/index?invite=" + code,
		SettledCount:  settled,
		PendingCount:  pending,
		MaxSettle:     model.InviteMaxSettle,
		Unlocked:      u.HasVip(),
		UnseenSettled: unseen,
		Records:       infos,
	}, nil
}

// ==================== 绑定邀请关系 ====================

// BindInvite 提交邀请码，绑定邀请关系。
//
// 时机刻意放在「注册后、提交审核前」：此时只建立关系不发放任何权益，
// 被邀请人仍需走完资料填写与人工审核。审核通过时由 AuditUser 触发结算。
func BindInvite(me *model.User, rawCode string) error {
	code := normalizeCode(rawCode)
	if code == "" {
		return Biz("请填写邀请码")
	}

	// 已有绑定关系就不重复处理（幂等）：用户可能在向导里多次进来
	var exist model.InviteRelation
	err := db.DB.Where("invitee_id = ?", me.ID).First(&exist).Error
	if err == nil {
		return Biz("你已经使用过邀请码了")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// 邀请码只属于「新用户首登 → 填资料 → 过审」这条链路。
	// 已是正常状态的老用户绑了也永远不会结算（结算只挂在审核通过上），
	// 只会留下一条永远 pending 的关系，让邀请人误以为对方在审核中
	if me.Status != model.StatusIncomplete && me.Status != model.StatusAuditing &&
		me.Status != model.StatusRejected {
		return Biz("邀请码仅限首次提交资料前填写")
	}

	var owner model.InviteCode
	if err := db.DB.Where("code = ?", code).First(&owner).Error; err != nil {
		return Biz("邀请码不存在，请核对后重试")
	}
	if owner.UserID == me.ID {
		return Biz("不能使用自己的邀请码")
	}

	// 邀请人必须是「已通过人工审核」的正常账号（与规则页、docs/06 口径一致）。
	// 未完善/审核中/被拒/封禁的账号不参与 —— 否则批量注册空白小号就能互相拉
	var inviter model.User
	if err := db.DB.First(&inviter, owner.UserID).Error; err != nil {
		return Biz("邀请人账号不存在")
	}
	if inviter.Status != model.StatusNormal {
		return Biz("该邀请码暂不可用，邀请人账号尚未通过审核或已失效")
	}

	var cnt int64
	db.DB.Model(&model.InviteRelation{}).
		Where("inviter_id = ? AND status = ?", owner.UserID, model.InviteSettled).Count(&cnt)
	if cnt >= model.InviteMaxSettle {
		return Biz("该邀请码已达邀请上限")
	}

	rel := model.InviteRelation{
		InviterID: owner.UserID,
		InviteeID: me.ID,
		Code:      code,
		Status:    model.InvitePending,
	}
	if err := db.DB.Create(&rel).Error; err != nil {
		// 唯一索引冲突：并发重复提交，按幂等处理
		if isDuplicateKey(err) {
			return Biz("你已经使用过邀请码了")
		}
		return fmt.Errorf("保存邀请关系失败: %w", err)
	}

	logger.Infof("邀请关系绑定: inviter=%d invitee=%d code=%s", owner.UserID, me.ID, code)
	return nil
}

// ==================== 结算 ====================

// SettleInvite 结算邀请权益。由「资料审核通过」这一动作触发。
//
// 为什么必须挂在审核通过而不是注册：
//  1. 这是「深度互动」的最强证明 —— 被邀请人走完了完整业务流程；
//  2. 空号、资料不全、审核被拒的号拿不到奖励，天然对抗刷量；
//  3. 结算时机越靠后，薅羊毛成本越高。
//
// 结算失败不阻断审核主流程：审核是业务主线，权益发放是附加动作，
// 不能因为发权益出错就让管理员审不了资料。故只记日志，返回 nil。
func SettleInvite(inviteeID uint64) {
	var rel model.InviteRelation
	err := db.DB.Where("invitee_id = ? AND status = ?", inviteeID, model.InvitePending).
		First(&rel).Error
	if err != nil {
		return // 没有邀请关系，或已结算过，都无需处理
	}

	// 邀请人此时可能已掉出正常状态（被拒/封禁），结算前再校验一次
	var inviter model.User
	if err := db.DB.First(&inviter, rel.InviterID).Error; err != nil {
		logger.Warnf("邀请结算跳过：邀请人不存在 inviter=%d", rel.InviterID)
		return
	}
	if inviter.Status != model.StatusNormal {
		db.DB.Model(&rel).Update("status", model.InviteCanceled)
		logger.Infof("邀请作废：邀请人状态非正常(%d) inviter=%d invitee=%d",
			inviter.Status, rel.InviterID, inviteeID)
		return
	}

	var settled int64
	db.DB.Model(&model.InviteRelation{}).
		Where("inviter_id = ? AND status = ?", rel.InviterID, model.InviteSettled).Count(&settled)
	if settled >= model.InviteMaxSettle {
		db.DB.Model(&rel).Update("status", model.InviteCanceled)
		logger.Infof("邀请作废：邀请人已达上限 inviter=%d invitee=%d", rel.InviterID, inviteeID)
		return
	}

	now := time.Now()
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&rel).Updates(map[string]any{
			"status": model.InviteSettled, "settled_at": now,
		}).Error; err != nil {
			return err
		}
		if err := grantVip(tx, rel.InviterID); err != nil {
			return err
		}
		return grantVip(tx, rel.InviteeID)
	})
	if err != nil {
		logger.Errorf("邀请结算失败 inviter=%d invitee=%d: %v", rel.InviterID, inviteeID, err)
		return
	}

	logger.Infof("邀请结算成功: inviter=%d invitee=%d 双方权益已永久解锁",
		rel.InviterID, inviteeID)
}

// MarkInviteSeen 邀请人查看过结算结果后清除红点。
//
// 一次性把所有「已结算未查看」置为已看：红点与弹层表达的是
// 「有新结算」，是开关不是逐条状态，无需按条标记。
func MarkInviteSeen(me *model.User) error {
	return db.DB.Model(&model.InviteRelation{}).
		Where("inviter_id = ? AND status = ? AND inviter_seen = ?",
			me.ID, model.InviteSettled, false).
		Update("inviter_seen", true).Error
}

// grantVip 发放权益：把用户标记为「已解锁完整资料」，永久有效。
//
// 为什么是永久而不是发天数（见 docs/08-权益体系重构方案.md 4.4）：
//  1. 赠送权益用天数逻辑没有心理锚点 —— 用户看到「剩 7 天」不知道能换来什么，
//     付费产品能用天数是因为有价格对照（30 元/月 → 每天 1 元），我们没有；
//  2. 倒计时会加速犹豫型用户流失，与婚恋场景的低压决策相悖；
//  3. 「永久」用户不需要理解任何规则，也不会因为到期产生咨询工单。
//
// 防刷不靠时间衰减，而靠「被邀请人必须通过人工审核」这一成本闸门。
//
// 幂等：已经是 1 的不重复写库，避免把 UpdatedAt 刷得没有意义。
func grantVip(tx *gorm.DB, userID uint64) error {
	return tx.Model(&model.User{}).
		Where("id = ? AND vip_level < ?", userID, model.VipLevelUnlocked).
		Update("vip_level", model.VipLevelUnlocked).Error
}

// ==================== 内部工具 ====================

// ensureInviteCode 取我的邀请码，没有则惰性生成一个
func ensureInviteCode(userID uint64) (string, error) {
	var ic model.InviteCode
	err := db.DB.Where("user_id = ?", userID).First(&ic).Error
	if err == nil {
		return ic.Code, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}

	// 碰撞重试：短码空间有限，靠唯一索引兜底 + 有限次重试
	for i := 0; i < model.InviteCodeRetries; i++ {
		code := randomCode(model.InviteCodeLength)
		ic = model.InviteCode{UserID: userID, Code: code}
		if err := db.DB.Create(&ic).Error; err != nil {
			if isDuplicateKey(err) {
				// 可能是码撞了，也可能是同一用户并发建码；先查一次用户
				var exist model.InviteCode
				if e := db.DB.Where("user_id = ?", userID).First(&exist).Error; e == nil {
					return exist.Code, nil
				}
				continue
			}
			return "", err
		}
		return code, nil
	}
	return "", errors.New("生成邀请码失败，请稍后重试")
}

// 去掉易混淆字符（0/O、1/I/L），方便用户手动输入与口头转述
const codeAlphabet = "23456789ABCDEFGHJKMNPQRSTUVWXYZ"

func randomCode(n int) string {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		// crypto/rand 失败极为罕见，退化为时间戳派生，避免整个功能不可用
		seed := time.Now().UnixNano()
		for i := range buf {
			buf[i] = byte(seed >> (i % 8))
		}
	}
	out := make([]byte, n)
	for i := range buf {
		out[i] = codeAlphabet[int(buf[i])%len(codeAlphabet)]
	}
	return string(out)
}

// normalizeCode 邀请码统一转大写并去空白，用户小写输入也能通过
func normalizeCode(s string) string {
	return strings.ToUpper(strings.TrimSpace(s))
}

// maskNickname 昵称脱敏。邀请页只给邀请人看个大概，不全量暴露被邀请人昵称
func maskNickname(name string) string {
	r := []rune(strings.TrimSpace(name))
	if len(r) == 0 {
		return "新用户"
	}
	if len(r) <= 1 {
		return string(r) + "*"
	}
	return string(r[:1]) + "**"
}

// isDuplicateKey 判断是否唯一索引冲突（MySQL 1062）
func isDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "1062") ||
		strings.Contains(err.Error(), "Duplicate entry")
}
