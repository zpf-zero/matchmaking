package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户账号。管理员与普通用户共用此表，通过 Role 区分
type User struct {
	ID           uint64         `gorm:"primaryKey" json:"id"`
	OpenID       string         `gorm:"column:openid;size:64" json:"-"`
	Phone        string         `gorm:"size:20" json:"-"` // 手机号不对外暴露，仅本人接口单独返回
	Nickname     string         `gorm:"size:32" json:"nickname"`
	Avatar       string         `gorm:"size:255" json:"avatar"`
	Gender       int8           `json:"gender"` // 1男 2女，注册后不可改
	Role         int8           `gorm:"default:1" json:"role"`
	Status       int8           `json:"status"`
	BanReason    string         `gorm:"size:255" json:"ban_reason,omitempty"`
	VipLevel     int8           `gorm:"default:0" json:"vip_level"` // 权益等级 0无 1已解锁完整资料（永久）
	VipExpireAt  *time.Time     `json:"-"`                          // 旧字段，已由 VipLevel 取代，仅为兼容保留，不对外返回
	LastActiveAt *time.Time     `json:"last_active_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"-"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// HasVip 当前是否拥有权益。
//
// 判定只看 VipLevel，不再看到期时间 —— 权益是「永久解锁」而非「限时」。
// 详见 docs/08-权益体系重构方案.md 4.4：赠送权益用天数逻辑没有心理锚点
// （用户看到「剩 7 天」不知道能换来什么），且倒计时会加速犹豫型用户流失。
func (u *User) HasVip() bool {
	return u.VipLevel >= VipLevelUnlocked
}

// UserProfile 详细资料
type UserProfile struct {
	ID            uint64         `gorm:"primaryKey" json:"id"`
	UserID        uint64         `json:"user_id"`
	Birthday      *time.Time     `gorm:"type:date" json:"birthday"`
	Age           int8           `json:"age"` // 冗余字段，保存资料时计算，筛选走索引
	Height        int16          `json:"height"`
	Weight        int16          `json:"weight"`
	MaritalStatus int8           `json:"marital_status"`
	HasChildren   int8           `json:"has_children"`
	Education     int8           `json:"education"`
	School        string         `gorm:"size:64" json:"school"`
	Occupation    string         `gorm:"size:64" json:"occupation"`
	IncomeLevel   int8           `json:"income_level"`
	HouseStatus   int8           `json:"house_status"`
	CarStatus     int8           `json:"car_status"`
	Province      string         `gorm:"size:32" json:"province"`
	City          string         `gorm:"size:32" json:"city"`
	CityCode      string         `gorm:"size:16" json:"city_code"`
	HomeProvince  string         `gorm:"size:32" json:"home_province"`
	HomeCity      string         `gorm:"size:32" json:"home_city"`
	SelfIntro     string         `gorm:"size:1000" json:"self_intro"`
	Industry      int8           `json:"industry"` // 行业标签（粗粒度），免费内容

	// ---- 原生家庭（权益内容，默认私有，需用户主动开启 FamilyVisible）----
	// 敏感度最高的资料块。默认不公开是刻意的：默认公开会劝退大量用户。
	// 父母职业只记「类型」不记具体单位 —— 写「公务员」可以，写「XX局」不行，
	// 后者结合年龄/城市可被反查到具体的人。
	FamilyParentOcc int8   `json:"family_parent_occ"`
	FamilyAtmosphere int8  `json:"family_atmosphere"`
	FamilyOnlyChild  int8  `json:"family_only_child"` // 1是 2否
	FamilyCity       string `gorm:"size:32" json:"family_city"`
	FamilyNote       string `gorm:"size:200" json:"family_note"`
	FamilyVisible    int8   `json:"family_visible"` // 1 表示同意对权益用户展示

	CompleteRate  int8           `json:"complete_rate"`
	AuditStatus   int8           `json:"audit_status"`
	AuditRemark   string         `gorm:"size:255" json:"audit_remark"`
	AuditedAt     *time.Time     `json:"audited_at"`
	CreatedAt     time.Time      `json:"-"`
	UpdatedAt     time.Time      `json:"-"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// ThreeViewAnswer 三观问答答案。
//
// 答题免费、「看到对方的答案」才需权益 —— 若答题也要权益就没人填了，
// 而题库的价值恰恰在于让所有人都填（见方案 4.2②）。
type ThreeViewAnswer struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	UserID    uint64    `json:"user_id"`
	QID       int16     `json:"q_id"`       // 对应 ThreeViewQuestions 的 ID
	OptionIdx int8      `json:"option_idx"` // 所选选项下标
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// UserRequirement 择偶要求
type UserRequirement struct {
	ID             uint64         `gorm:"primaryKey" json:"id"`
	UserID         uint64         `json:"user_id"`
	AgeMin         int8           `json:"age_min"`
	AgeMax         int8           `json:"age_max"`
	HeightMin      int16          `json:"height_min"`
	HeightMax      int16          `json:"height_max"`
	EducationMin   int8           `json:"education_min"`
	IncomeMin      int8           `json:"income_min"`
	MaritalStatus  string         `gorm:"size:32" json:"marital_status"` // 逗号分隔 "1,2"
	AcceptChildren int8           `gorm:"default:1" json:"accept_children"`
	CityCodes      string         `gorm:"size:255" json:"city_codes"` // 逗号分隔
	Other          string         `gorm:"size:500" json:"other"`
	CreatedAt      time.Time      `json:"-"`
	UpdatedAt      time.Time      `json:"-"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// UserPhoto 相册。头像单独存在 User.Avatar，不在此表
type UserPhoto struct {
	ID          uint64         `gorm:"primaryKey" json:"id"`
	UserID      uint64         `json:"user_id"`
	URL         string         `gorm:"column:url;size:255" json:"url"`
	Sort        int16          `json:"sort"`
	AuditStatus int8           `json:"audit_status"`
	CreatedAt   time.Time      `json:"-"`
	UpdatedAt   time.Time      `json:"-"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// Interaction 互动记录（浏览/喜欢/不喜欢）
type Interaction struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	FromUserID uint64    `json:"from_user_id"`
	ToUserID   uint64    `json:"to_user_id"`
	Type       int8      `json:"type"`
	IsRead     int8      `json:"is_read"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"-"`
}

// Conversation 会话（= 配对关系）。UserAID 恒为较小的 user_id
type Conversation struct {
	ID            uint64     `gorm:"primaryKey" json:"id"`
	UserAID       uint64     `json:"user_a_id"`
	UserBID       uint64     `json:"user_b_id"`
	LastMessage   string     `gorm:"size:255" json:"last_message"`
	LastMessageAt *time.Time `json:"last_message_at"`
	AUnread       int        `json:"-"`
	BUnread       int        `json:"-"`
	Status        int8       `json:"status"`
	MatchedAt     *time.Time `json:"matched_at"`
	CreatedAt     time.Time  `json:"-"`
	UpdatedAt     time.Time  `json:"-"`
}

// PeerID 返回会话中对方的 user_id
func (c *Conversation) PeerID(myID uint64) uint64 {
	if c.UserAID == myID {
		return c.UserBID
	}
	return c.UserAID
}

// UnreadFor 返回指定用户在该会话中的未读数
func (c *Conversation) UnreadFor(myID uint64) int {
	if c.UserAID == myID {
		return c.AUnread
	}
	return c.BUnread
}

// Message 消息
type Message struct {
	ID             uint64    `gorm:"primaryKey" json:"id"`
	ConversationID uint64    `json:"conversation_id"`
	FromUserID     uint64    `json:"from_user_id"`
	Type           int8      `json:"type"`
	Content        string    `gorm:"size:1000" json:"content"`
	IsRead         int8      `json:"is_read"`
	CreatedAt      time.Time `json:"created_at"`
}

// Blacklist 黑名单
type Blacklist struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	UserID       uint64    `json:"user_id"`
	TargetUserID uint64    `json:"target_user_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// InviteCode 邀请码。一个用户恒定一个短码
type InviteCode struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	UserID    uint64    `json:"user_id"`
	Code      string    `gorm:"size:12" json:"code"`
	CreatedAt time.Time `json:"created_at"`
}

// InviteRelation 邀请关系与结算。
//
// 绑定与结算刻意分离：注册/填码时写入 status=0，被邀请人通过人工审核后
// 才置为 1 并发放双方权益。这是本方案合规的基础 —— 被邀请人必须走完
// 「注册 → 完善资料 → 通过审核」这段真实业务流程，才构成微信规范认定的
// 「深度互动」，而空号/资料不全的号拿不到奖励，也天然对抗刷量。
type InviteRelation struct {
	ID          uint64     `gorm:"primaryKey" json:"id"`
	InviterID   uint64     `json:"inviter_id"`
	InviteeID   uint64     `json:"invitee_id"` // 唯一索引，一人只能被邀请一次
	Code        string     `gorm:"size:12" json:"code"`
	Status      int8       `json:"status"`
	SettledAt   *time.Time `json:"settled_at"`
	InviterSeen bool       `json:"-"` // 结算后邀请人是否已查看（红点/弹层用），不对外返回
	CreatedAt   time.Time  `json:"created_at"`
}

// SmsCode 短信验证码。只存哈希不存明文；同一手机号保留多条历史用于限频审计
type SmsCode struct {
	ID        uint64     `gorm:"primaryKey" json:"-"`
	Phone     string     `gorm:"size:20" json:"-"`
	CodeHash  string     `gorm:"size:64" json:"-"`
	Scene     string     `gorm:"size:16" json:"-"` // login
	IP        string     `gorm:"size:64" json:"-"`
	Attempts  int8       `json:"-"`
	UsedAt    *time.Time `json:"-"`
	ExpiresAt time.Time  `json:"-"`
	CreatedAt time.Time  `json:"-"`
}

// Report 举报
type Report struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	FromUserID   uint64    `json:"from_user_id"`
	TargetUserID uint64    `json:"target_user_id"`
	Reason       int8      `json:"reason"`
	Description  string    `gorm:"size:500" json:"description"`
	Images       string    `gorm:"size:1000" json:"images"` // 逗号分隔
	Status       int8      `json:"status"`
	HandleRemark string    `gorm:"size:255" json:"handle_remark"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"-"`
}
