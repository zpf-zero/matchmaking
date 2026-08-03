package dto

import (
	"time"

	"hunjie/internal/model"
)

// ==================== 认证 ====================

type WxLoginReq struct {
	Code string `json:"code" binding:"required"`
}

type WxLoginResp struct {
	Token       string      `json:"token"`
	User        *model.User `json:"user"`
	Status      int8        `json:"status"`
	Role        int8        `json:"role"`
	NeedProfile bool        `json:"need_profile"`
}

type BindPhoneReq struct {
	Code  string `json:"code"`  // 真实模式：getPhoneNumber 回调的 code
	Phone string `json:"phone"` // mock 模式：前端手填
}

// ==================== 我的资料 ====================

// SaveProfileReq 全量保存资料。前端在本地暂存 4 步向导的数据，最后一次性提交
type SaveProfileReq struct {
	// user 表字段
	Nickname string `json:"nickname" binding:"required"`
	Avatar   string `json:"avatar" binding:"required"`
	Gender   int8   `json:"gender"` // 首次提交后锁死，后续修改无效

	// user_profile 表字段
	Birthday      string `json:"birthday" binding:"required"` // YYYY-MM-DD
	Height        int16  `json:"height" binding:"required"`
	Weight        int16  `json:"weight"`
	MaritalStatus int8   `json:"marital_status" binding:"required"`
	HasChildren   int8   `json:"has_children"`
	Education     int8   `json:"education" binding:"required"`
	School        string `json:"school"`
	Occupation    string `json:"occupation" binding:"required"`
	IncomeLevel   int8   `json:"income_level" binding:"required"`
	HouseStatus   int8   `json:"house_status" binding:"required"`
	CarStatus     int8   `json:"car_status" binding:"required"`
	Province      string `json:"province" binding:"required"`
	City          string `json:"city" binding:"required"`
	CityCode      string `json:"city_code"`
	HomeProvince  string `json:"home_province"`
	HomeCity      string `json:"home_city"`
	SelfIntro     string `json:"self_intro" binding:"required"`
}

type SaveRequirementReq struct {
	AgeMin         int8   `json:"age_min"`
	AgeMax         int8   `json:"age_max"`
	HeightMin      int16  `json:"height_min"`
	HeightMax      int16  `json:"height_max"`
	EducationMin   int8   `json:"education_min"`
	IncomeMin      int8   `json:"income_min"`
	MaritalStatus  string `json:"marital_status"` // "1,2"
	AcceptChildren int8   `json:"accept_children"`
	CityCodes      string `json:"city_codes"`
	Other          string `json:"other"`
}

type MeResp struct {
	User        *model.User            `json:"user"`
	Phone       string                 `json:"phone"` // 仅本人接口返回
	Profile     *model.UserProfile     `json:"profile"`
	Requirement *model.UserRequirement `json:"requirement"`
	Photos      []model.UserPhoto      `json:"photos"`
}

type AddPhotoReq struct {
	URL string `json:"url" binding:"required"`
}

type SortPhotoReq struct {
	IDs []uint64 `json:"ids" binding:"required"` // 按顺序排列的照片 id
}

// ==================== 他人资料 ====================

// UserCard 列表卡片。只带筛选相关的硬条件，详情才返回全部
type UserCard struct {
	ID            uint64 `json:"id"`
	Nickname      string `json:"nickname"`
	Avatar        string `json:"avatar"`
	Gender        int8   `json:"gender"`
	Age           int8   `json:"age"`
	Height        int16  `json:"height"`
	Education     int8   `json:"education"`
	EducationText string `json:"education_text"`
	IncomeLevel   int8   `json:"income_level"`
	IncomeText    string `json:"income_text"`
	MaritalStatus int8   `json:"marital_status"`
	MaritalText   string `json:"marital_text"`
	City          string `json:"city"`
	Occupation    string `json:"occupation"`
	MatchRate     int    `json:"match_rate"`  // 双向匹配度百分比，仅展示不参与排序
	ActiveText    string `json:"active_text"` // "刚刚活跃" / "3天前活跃"
}

type UserDetailResp struct {
	UserCard
	Weight       int16    `json:"weight"`
	HasChildren  int8     `json:"has_children"`
	School       string   `json:"school"`
	HouseStatus  int8     `json:"house_status"`
	CarStatus    int8     `json:"car_status"`
	Province     string   `json:"province"`
	HomeProvince string   `json:"home_province"`
	HomeCity     string   `json:"home_city"`
	SelfIntro    string   `json:"self_intro"`
	Photos       []string `json:"photos"`

	// 与我的关系
	IsLiked        bool   `json:"is_liked"`  // 我是否已喜欢 ta
	LikedMe        bool   `json:"liked_me"`  // ta 是否已喜欢我
	IsMatched      bool   `json:"is_matched"`
	ConversationID uint64 `json:"conversation_id"`
	IsBlocked      bool   `json:"is_blocked"`
}

// ==================== 通用 ====================

type PageQuery struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

// SearchQuery 筛选条件。0 / 空字符串一律表示"不限"
type SearchQuery struct {
	PageQuery
	AgeMin        int8   `form:"age_min"`
	AgeMax        int8   `form:"age_max"`
	HeightMin     int16  `form:"height_min"`
	HeightMax     int16  `form:"height_max"`
	EducationMin  int8   `form:"education_min"`
	IncomeMin     int8   `form:"income_min"`
	MaritalStatus string `form:"marital_status"` // "1,2"
	HasChildren   string `form:"has_children"`   // 用字符串是因为 0 是有效值（无子女）
	CityCode      string `form:"city_code"`
	HouseStatus   int8   `form:"house_status"`
	CarStatus     int8   `form:"car_status"`
}

type TargetReq struct {
	ToUserID uint64 `json:"to_user_id" binding:"required"`
}

// ==================== 举报 / 拉黑 ====================

type ReportReq struct {
	TargetUserID uint64 `json:"target_user_id" binding:"required"`
	Reason       int8   `json:"reason" binding:"required"`
	Description  string `json:"description"`
	Images       string `json:"images"` // 逗号分隔
}

// ==================== 管理端 ====================

type AdminUserItem struct {
	ID           uint64     `json:"id"`
	Nickname     string     `json:"nickname"`
	Avatar       string     `json:"avatar"`
	Phone        string     `json:"phone"`
	Gender       int8       `json:"gender"`
	Role         int8       `json:"role"`
	Status       int8       `json:"status"`
	BanReason    string     `json:"ban_reason"`
	CreatedAt    *time.Time `json:"created_at"`
	LastActiveAt *time.Time `json:"last_active_at"`
	Age          int8       `json:"age"`
	City         string     `json:"city"`
	Education    int8       `json:"education"`
	IncomeLevel  int8       `json:"income_level"`
	AuditStatus  int8       `json:"audit_status"`
}

type AuditReq struct {
	Pass   bool   `json:"pass"`
	Remark string `json:"remark"`
}

type BanReq struct {
	Reason string `json:"reason"`
}

type HandleReportReq struct {
	Action string `json:"action" binding:"required"` // ignore / ban
	Remark string `json:"remark"`
}

// Normalize 兜底分页参数
func (p *PageQuery) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 || p.PageSize > 50 {
		p.PageSize = 20
	}
}

func (p PageQuery) Offset() int { return (p.Page - 1) * p.PageSize }
