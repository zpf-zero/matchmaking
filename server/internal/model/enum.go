package model

// ============ 用户状态 ============
const (
	StatusIncomplete = 0 // 待完善资料
	StatusAuditing   = 1 // 待审核
	StatusNormal     = 2 // 正常
	StatusRejected   = 3 // 审核未通过
	StatusBanned     = 4 // 已封禁
)

// ============ 角色（向上包含）============
const (
	RoleUser       = 1 // 普通用户
	RoleAuditor    = 2 // 审核员
	RoleAdmin      = 3 // 管理员
	RoleSuperAdmin = 4 // 超级管理员
)

// ============ 性别 ============
const (
	GenderMale   = 1
	GenderFemale = 2
)

// OppositeGender 返回异性性别值。一期只推异性
func OppositeGender(g int8) int8 {
	if g == GenderMale {
		return GenderFemale
	}
	return GenderMale
}

// ============ 审核状态 ============
const (
	AuditPending  = 0
	AuditPassed   = 1
	AuditRejected = 2
)

// ============ 互动类型 ============
const (
	InteractionView = 1 // 浏览
	InteractionLike = 2 // 喜欢
	InteractionPass = 3 // 不喜欢
)

// ============ 会话状态 ============
const (
	ConvNormal = 1
	ConvClosed = 2
)

// ============ 消息类型 ============
const (
	MsgText   = 1
	MsgImage  = 2
	MsgSystem = 3
)

// ============ 举报状态 ============
const (
	ReportPending  = 0
	ReportHandled  = 1
	ReportRejected = 2
)

// ============ 字典 ============
// 这些枚举一年都不会改一次，放代码里即可，不建字典表。
// 通过 GET /api/v1/common/dict 一次性返回给前端缓存。

type DictItem struct {
	Value int    `json:"value"`
	Label string `json:"label"`
}

var Dict = map[string][]DictItem{
	"education": {
		{1, "高中及以下"}, {2, "大专"}, {3, "本科"}, {4, "硕士"}, {5, "博士"},
	},
	"income_level": {
		{1, "5千以下"}, {2, "5千-1万"}, {3, "1万-2万"},
		{4, "2万-3万"}, {5, "3万-5万"}, {6, "5万以上"},
	},
	"marital_status": {
		{1, "未婚"}, {2, "离异"}, {3, "丧偶"},
	},
	"has_children": {
		{0, "无"}, {1, "有（同住）"}, {2, "有（不同住）"},
	},
	"house_status": {
		{1, "已购房"}, {2, "与父母同住"}, {3, "租房"}, {4, "计划购房"},
	},
	"car_status": {
		{1, "已购车"}, {2, "暂未购车"},
	},
	"gender": {
		{1, "男"}, {2, "女"},
	},
	"report_reason": {
		{1, "虚假资料"}, {2, "骚扰"}, {3, "广告"}, {4, "违法信息"}, {5, "其他"},
	},
}

// DictLabel 取字典项的显示名
func DictLabel(dictType string, value int) string {
	for _, item := range Dict[dictType] {
		if item.Value == value {
			return item.Label
		}
	}
	return ""
}
