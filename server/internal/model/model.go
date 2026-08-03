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
	LastActiveAt *time.Time     `json:"last_active_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"-"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
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
	CompleteRate  int8           `json:"complete_rate"`
	AuditStatus   int8           `json:"audit_status"`
	AuditRemark   string         `gorm:"size:255" json:"audit_remark"`
	AuditedAt     *time.Time     `json:"audited_at"`
	CreatedAt     time.Time      `json:"-"`
	UpdatedAt     time.Time      `json:"-"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
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
