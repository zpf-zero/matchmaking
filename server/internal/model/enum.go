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

// ============ 权益等级 ============
const (
	VipLevelNone     = 0 // 无权益
	VipLevelUnlocked = 1 // 已解锁完整资料（永久）
)

// ============ 资料可见层级 ============
// 见 docs/08-权益体系重构方案.md 4.1。
// 门禁做在「数据量」上而不是「接口」上：免费用户依然能正常浏览、喜欢、聊天，
// 只是看不到更深入的内容 —— 与「谁看过我」现有设计保持一致。
const (
	// FreePhotoLimit 免费可见的照片张数（按 sort 顺序）。第 4 张起为权益内容。
	// 锁的是「更多」而不是「基础」：免费用户仍能看到前 3 张并正常喜欢对方。
	FreePhotoLimit = 3
)

// ============ 邀请关系状态 ============
const (
	InvitePending  = 0 // 待达成：已绑定，被邀请人尚未通过审核
	InviteSettled  = 1 // 已结算：被邀请人通过审核，双方权益已发放
	InviteCanceled = 2 // 已作废：被邀请人始终未通过审核或被封禁
)

// 邀请权益上限。改这里即可调整运营节奏。
//
// 注意：不再有 InviteRewardDays —— 权益已改为「永久解锁」，
// 没有天数可发。防刷靠「被邀请人须通过人工审核」这一成本闸门，
// 而不是靠时间衰减（见 docs/08-权益体系重构方案.md 4.4）。
const (
	InviteMaxSettle   = 20 // 单账号最多可结算的邀请人数，防红娘化刷量
	InviteCodeLength  = 6  // 邀请码长度。小程序码 scene 上限 32 字符，留足余量
	InviteCodeRetries = 8  // 生成邀请码时的唯一性重试次数
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

	// ---- 行业标签（粗粒度）----
	// 刻意只到「系统」级别，不到具体单位：体制内用户对可被识别的信息极敏感，
	// 写「教育系统」是归属感，写「XX市第三中学」就是可反查到人。
	// 见 docs/08-权益体系重构方案.md 3.3 方案 B。
	"industry": {
		{1, "机关事业单位"}, {2, "教育系统"}, {3, "医疗卫生"}, {4, "公检法"},
		{5, "国有企业"}, {6, "金融银行"}, {7, "互联网/IT"}, {8, "其他企业"},
		{9, "自由职业"}, {10, "其他"},
	},

	// ---- 原生家庭 ----
	"family_parent_occ": {
		{1, "均为公职"}, {2, "一方公职"}, {3, "均为企业职工"},
		{4, "个体经营"}, {5, "务农"}, {6, "已退休"}, {7, "不便透露"},
	},
	"family_atmosphere": {
		{1, "氛围融洽"}, {2, "较为和睦"}, {3, "比较独立"}, {4, "不便透露"},
	},
	"family_only_child": {
		{1, "是"}, {2, "否"},
	},
}

// ==================== 三观问答题库 ====================

// ThreeViewQuestion 一道三观题
type ThreeViewQuestion struct {
	ID      int16    `json:"id"`
	Title   string   `json:"title"`
	Options []string `json:"options"`
}

// ThreeViewQuestions 题库。
//
// 为什么不建题目表：题目一年不会改一次，建表等于给自己加一套 CRUD 界面
// （与「枚举放代码不放数据库」的既有约定一致）。
// 答题记录落 three_view_answer 表，通过 QID 引用这里的题目。
//
// 题目设计原则：
//  1. 都是「无对错」的偏好题 —— 没有标准答案才不会让人为了讨好而撒谎
//  2. 覆盖金钱观、生活节奏、家庭观念、沟通方式四个高频冲突维度
//  3. 选项控制在 3~4 个，答题耗时不超过 1 分钟
//
// ⚠️ 改动选项数量/顺序会让历史答案的 OptionIdx 语义漂移。
// 若要改，应新增题目（新 ID）而不是修改已有题目的选项。
var ThreeViewQuestions = []ThreeViewQuestion{
	{
		ID: 1, Title: "关于消费与储蓄，你更认同",
		Options: []string{"及时行乐，该花就花", "有规划地花，也留余地", "能省则省，储蓄优先", "看具体情况，没有固定偏好"},
	},
	{
		ID: 2, Title: "理想的周末通常是什么样",
		Options: []string{"在家休息，谁也不见", "两个人安静地待着", "约朋友一起出门", "出去走走，换换环境"},
	},
	{
		ID: 3, Title: "婚后你更希望的居住安排是",
		Options: []string{"与父母同住，方便照应", "单独住，但离父母近", "单独住，距离不是问题", "还没想好，可以一起商量"},
	},
	{
		ID: 4, Title: "两个人意见不合时，你倾向于",
		Options: []string{"当场说清楚，不留隔夜气", "先各自冷静，再坐下来谈", "多让一步，不必争对错", "找第三方（朋友/长辈）聊聊"},
	},
	{
		ID: 5, Title: "对生育的想法，你更接近",
		Options: []string{"希望要孩子", "希望要，但可以晚几年", "倾向于不要", "尊重对方意愿，可以商量"},
	},
	{
		ID: 6, Title: "关于工作和生活的比重",
		Options: []string{"事业为重，愿意为工作投入更多", "工作生活平衡，两者都要", "生活为重，工作够用就好", "看阶段，视情况调整"},
	},
	{
		ID: 7, Title: "逢年过节，你更倾向于",
		Options: []string{"双方父母轮流过", "一起过，两边都顾上", "各回各家，各自陪父母", "按实际情况灵活安排"},
	},
	{
		ID: 8, Title: "对另一半的社交圈，你的态度是",
		Options: []string{"希望彼此生活高度重合", "各有一半独立空间更好", "完全互不干涉", "顺其自然，不刻意要求"},
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
