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

// ==================== 邀请 ====================

type BindInviteReq struct {
	Code string `json:"code" binding:"required"`
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

	// 行业标签（粗粒度）。选填，免费内容，用于「同系统」归属感，
	// 刻意只到「教育系统」这一级，不给具体单位（防反查，见方案 3.3）
	Industry int8 `json:"industry"`

	// ---- 原生家庭（选填，默认私有）----
	// 用户必须主动把 FamilyVisible 置 1 才会对权益用户展示。
	// 这是最敏感的资料块，默认公开会劝退大量用户（见方案 6-2）。
	FamilyParentOcc  int8   `json:"family_parent_occ"`
	FamilyAtmosphere int8   `json:"family_atmosphere"`
	FamilyOnlyChild  int8   `json:"family_only_child"`
	FamilyCity       string `json:"family_city"`
	FamilyNote       string `json:"family_note"`
	FamilyVisible    int8   `json:"family_visible"`
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
	User         *model.User            `json:"user"`
	Phone        string                 `json:"phone"` // 仅本人接口返回
	Profile      *model.UserProfile     `json:"profile"`
	Requirement  *model.UserRequirement `json:"requirement"`
	Photos       []model.UserPhoto      `json:"photos"`
	InviteUnseen int64                  `json:"invite_unseen"` // 已结算未查看的邀请数，驱动「我的」页红点
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
	Industry      int8   `json:"industry"`      // 行业标签（粗粒度），免费内容
	IndustryText  string `json:"industry_text"` // 熟人背书：「同系统」归属感，不给具体单位
	MatchRate     int    `json:"match_rate"`    // 双向匹配度百分比，仅展示不参与排序
	ActiveText    string `json:"active_text"`   // "刚刚活跃" / "3天前活跃"
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

	// ---- 照片分层 ----
	// 免费用户只能看前 FreePhotoLimit 张，其余以 locked_photo_count 计数返回。
	// 锁的是「更多」而不是「基础」——与「谁看过我」逻辑一致：
	// 免费用户依然能正常看到前 3 张并喜欢对方（见 docs/08 4.2③）。
	PhotoTotal       int  `json:"photo_total"`        // 审核通过的照片总数
	LockedPhotoCount int  `json:"locked_photo_count"` // 被权益锁住、需解锁才能看的张数
	PhotosLocked     bool `json:"photos_locked"`      // 是否存在被锁的照片

	// ---- 权益内容（L2）----
	// Family 只有「对方主动授权（family_visible=1）」且「我已解锁」时才有值。
	// 三观与原生家庭分开控制：三观默认公开（是加分项），原生家庭默认私有。
	Family *FamilyInfo `json:"family"`
	// ViewedFamily 我是否能看到对方的原生家庭。false 时 Family 必为 nil，
	// 前端据此区分「对方没填」和「我没权限看」两种空状态。
	ViewedFamily bool `json:"viewed_family"`
	// ThreeView 对方的三观答案。答题免费、看答案需权益（见方案 6-3）。
	// 未解锁时返回 nil，但 ThreeViewAnswered 仍会告知对方是否答过题。
	ThreeView         *ThreeViewInfo `json:"three_view"`
	ThreeViewAnswered bool           `json:"three_view_answered"`
	ThreeViewUnlocked bool           `json:"three_view_unlocked"`

	// 与我的关系
	IsLiked        bool   `json:"is_liked"` // 我是否已喜欢 ta
	LikedMe        bool   `json:"liked_me"` // ta 是否已喜欢我
	IsMatched      bool   `json:"is_matched"`
	ConversationID uint64 `json:"conversation_id"`
	IsBlocked      bool   `json:"is_blocked"`
}

// FamilyInfo 原生家庭（L2 权益内容，需对方主动授权公开）
type FamilyInfo struct {
	ParentOcc     int8   `json:"parent_occ"`
	ParentOccText string `json:"parent_occ_text"`
	Atmosphere    int8   `json:"atmosphere"`
	AtmosText     string `json:"atmosphere_text"`
	OnlyChild     int8   `json:"only_child"`
	OnlyChildText string `json:"only_child_text"`
	City          string `json:"city"`
	Note          string `json:"note"`
}

// ThreeViewInfo 三观问答的展示结构。
//
// Answers 与 MyAnswers 并列返回，前端可直接渲染「对方的选择 vs 我的选择」，
// 不需要再拉题库（题库通过 /common/three-view/questions 单独取）。
type ThreeViewInfo struct {
	Answers    []int8 `json:"answers"`      // 对方按题库顺序的选项下标，未答为 -1
	Answered   int    `json:"answered"`     // 对方已答题数
	Total      int    `json:"total"`        // 题库总题数
	MatchRate  int    `json:"match_rate"`   // 与我答案的一致率（0~100），仅用于展示
	MyAnswers  []int8 `json:"my_answers"`   // 我的答案，便于前端并列展示
	MyAnswered bool   `json:"my_answered"`  // 我是否答过题
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
	// Industry 行业标签筛选（「同系统」）。粗粒度，不暴露具体单位 ——
	// 这是「熟人背书」的合规落点：给归属感，不给可识别到人的信息。
	Industry int8 `form:"industry"`
}

// ==================== 三观问答 ====================

// SaveThreeViewReq 提交我的三观答案（全量覆盖）。答题免费。
type SaveThreeViewReq struct {
	// Answers 与题库顺序一一对应，长度须等于题库题数；未答用 -1 占位。
	Answers []int8 `json:"answers" binding:"required"`
}

// MyThreeViewResp 我的三观页数据
type MyThreeViewResp struct {
	Questions []model.ThreeViewQuestion `json:"questions"`
	Answers   []int8                    `json:"answers"`  // 与 Questions 同序，未答为 -1
	Answered  int                       `json:"answered"` // 已答题数
	Total     int                       `json:"total"`
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

// SetRoleReq 任命/撤销管理角色。role: 1 普通用户 / 2 审核员 / 3 管理员
type SetRoleReq struct {
	Role int8 `json:"role" binding:"required"`
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
