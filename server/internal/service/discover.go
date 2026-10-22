package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"hunjie/internal/db"
	"hunjie/internal/dto"
	"hunjie/internal/model"
)

// 列表查询用的扁平结构，一次 join 取完卡片需要的字段
type cardRow struct {
	ID            uint64
	Nickname      string
	Avatar        string
	Gender        int8
	LastActiveAt  *time.Time
	Age           int8
	Height        int16
	Weight        int16
	Education     int8
	School        string
	Occupation    string
	IncomeLevel   int8
	MaritalStatus int8
	HasChildren   int8
	HouseStatus   int8
	CarStatus     int8
	Province      string
	City          string
	CityCode      string
	HomeProvince  string
	HomeCity      string
	SelfIntro     string
	Industry      int8
}

const cardFields = `u.id, u.nickname, u.avatar, u.gender, u.last_active_at,
	p.age, p.height, p.weight, p.education, p.school, p.occupation, p.income_level,
	p.marital_status, p.has_children, p.house_status, p.car_status,
	p.province, p.city, p.city_code, p.home_province, p.home_city, p.self_intro, p.industry`

// baseQuery 可见用户的公共过滤条件：
// 异性、状态正常、资料审核通过、非我本人、排除黑名单双向
func baseQuery(me *model.User) *gorm.DB {
	q := db.DB.Table("user u").
		Joins("JOIN user_profile p ON p.user_id = u.id AND p.deleted_at IS NULL").
		Where("u.deleted_at IS NULL").
		Where("u.status = ?", model.StatusNormal).
		Where("p.audit_status = ?", model.AuditPassed).
		Where("u.id <> ?", me.ID)

	// 一期只推异性
	if me.Gender != 0 {
		q = q.Where("u.gender = ?", model.OppositeGender(me.Gender))
	}

	// 我拉黑的人 + 拉黑我的人，双向不可见
	q = q.Where("u.id NOT IN (?)",
		db.DB.Table("blacklist").Select("target_user_id").Where("user_id = ?", me.ID))
	q = q.Where("u.id NOT IN (?)",
		db.DB.Table("blacklist").Select("user_id").Where("target_user_id = ?", me.ID))

	return q
}

// Recommend 推荐流。用我的择偶要求做「硬门槛 + 软容差」过滤（见 applyRequirement 注释），
// 排序按活跃时间倒序 —— 刻意不用匹配度排序，原因见同一处注释。
func Recommend(me *model.User, page dto.PageQuery) ([]dto.UserCard, int64, error) {
	q := baseQuery(me)

	// 推荐流排除已经表态过的人（喜欢/不喜欢），搜索页则不排除
	q = q.Where("u.id NOT IN (?)",
		db.DB.Table("interaction").Select("to_user_id").
			Where("from_user_id = ? AND type IN ?", me.ID,
				[]int{model.InteractionLike, model.InteractionPass}))

	myReq := loadRequirement(me.ID)
	q = applyRequirement(q, myReq)

	return runCardQuery(me, q, page)
}

// Search 条件筛选。用请求参数过滤，与我的择偶要求无关
func Search(me *model.User, f dto.SearchQuery) ([]dto.UserCard, int64, error) {
	q := baseQuery(me)

	if f.AgeMin > 0 {
		q = q.Where("p.age >= ?", f.AgeMin)
	}
	if f.AgeMax > 0 {
		q = q.Where("p.age <= ?", f.AgeMax)
	}
	if f.HeightMin > 0 {
		q = q.Where("p.height >= ?", f.HeightMin)
	}
	if f.HeightMax > 0 {
		q = q.Where("p.height <= ?", f.HeightMax)
	}
	if f.EducationMin > 0 {
		q = q.Where("p.education >= ?", f.EducationMin)
	}
	if f.IncomeMin > 0 {
		q = q.Where("p.income_level >= ?", f.IncomeMin)
	}
	if vals := parseIntList(f.MaritalStatus); len(vals) > 0 {
		q = q.Where("p.marital_status IN ?", vals)
	}
	if f.HasChildren != "" {
		if v, err := strconv.Atoi(f.HasChildren); err == nil {
			q = q.Where("p.has_children = ?", v)
		}
	}
	if f.CityCode != "" {
		q = q.Where("p.city_code = ?", f.CityCode)
	}
	if f.HouseStatus > 0 {
		q = q.Where("p.house_status = ?", f.HouseStatus)
	}
	if f.CarStatus > 0 {
		q = q.Where("p.car_status = ?", f.CarStatus)
	}
	// 行业标签筛选：「同系统」归属感，粗粒度不暴露具体单位
	if f.Industry > 0 {
		q = q.Where("p.industry = ?", f.Industry)
	}

	return runCardQuery(me, q, f.PageQuery)
}

// runCardQuery 执行分页查询并组装卡片（含匹配度）
func runCardQuery(me *model.User, q *gorm.DB, page dto.PageQuery) ([]dto.UserCard, int64, error) {
	page.Normalize()

	var total int64
	if err := q.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []dto.UserCard{}, 0, nil
	}

	var rows []cardRow
	if err := q.Session(&gorm.Session{}).
		Select(cardFields).
		// 权益用户的曝光加成：把「最近活跃」的主排序让出一档给会员，
		// 但不做纯会员优先 —— 婚恋场景里对方不再上线，再高的曝光也没有意义。
		// 用 BOOLEAN 表达式做第一排序键，等价于「会员组内部仍按活跃度排」。
		//
		// 权益判定已从 vip_expire_at（到期时间）改为 vip_level（永久等级），
		// 见 docs/08-权益体系重构方案.md 4.4 与 migrations/003_benefit.sql。
		Order("(u.vip_level >= 1) DESC, u.last_active_at DESC, u.id DESC").
		Limit(page.PageSize).Offset(page.Offset()).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	cards := buildCards(me, rows)
	return cards, total, nil
}

// buildCards 批量组装卡片。匹配度需要对方的资料与择偶要求，一次性批量取，避免 N+1
func buildCards(me *model.User, rows []cardRow) []dto.UserCard {
	myProfile := loadProfile(me.ID)
	myReq := loadRequirement(me.ID)

	ids := make([]uint64, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.ID)
	}
	otherReqs := loadRequirements(ids)

	cards := make([]dto.UserCard, 0, len(rows))
	for _, r := range rows {
		cards = append(cards, dto.UserCard{
			ID:            r.ID,
			Nickname:      r.Nickname,
			Avatar:        r.Avatar,
			Gender:        r.Gender,
			Age:           r.Age,
			Height:        r.Height,
			Education:     r.Education,
			EducationText: model.DictLabel("education", int(r.Education)),
			IncomeLevel:   r.IncomeLevel,
			IncomeText:    model.DictLabel("income_level", int(r.IncomeLevel)),
			MaritalStatus: r.MaritalStatus,
			MaritalText:   model.DictLabel("marital_status", int(r.MaritalStatus)),
			City:          r.City,
			Occupation:    r.Occupation,
			Industry:      r.Industry,
			IndustryText:  model.DictLabel("industry", int(r.Industry)),
			MatchRate:     matchRate(myReq, myProfile, otherReqs[r.ID], rowToProfile(r)),
			ActiveText:    activeText(r.LastActiveAt),
		})
	}
	return cards
}

// GetUserDetail 他人资料详情。顺带记录一条浏览
//
// 分层可见性（L0/L1/L2）在这里统一收口，见 docs/08-权益体系重构方案.md 4.1。
// 门禁刻意做在「数据量」上而不是「接口」上：
// 免费用户依然能正常浏览资料、前 3 张照片、喜欢与聊天，
// 只是看不到更深入的内容（更多照片 / 三观答案 / 原生家庭）。
// 这与「谁看过我」的现有设计是同一套原则 —— 锁「更多」，不锁「基础」。
func GetUserDetail(me *model.User, targetID uint64) (*dto.UserDetailResp, error) {
	if targetID == me.ID {
		return nil, Biz("看自己的资料请去「我的」页面")
	}

	var rows []cardRow
	if err := baseQuery(me).Where("u.id = ?", targetID).
		Select(cardFields).Scan(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, Biz("该用户不存在或资料已下架")
	}
	r := rows[0]

	unlocked := me.HasVip()

	card := buildCards(me, rows)[0]
	resp := &dto.UserDetailResp{
		UserCard:     card,
		Weight:       r.Weight,
		HasChildren:  r.HasChildren,
		School:       r.School,
		HouseStatus:  r.HouseStatus,
		CarStatus:    r.CarStatus,
		Province:     r.Province,
		HomeProvince: r.HomeProvince,
		HomeCity:     r.HomeCity,
		SelfIntro:    r.SelfIntro,
	}

	// 相册只展示审核通过的。免费用户只取前 N 张，
	// 并单独统计总数与被锁张数，前端据此渲染「还有 N 张」的解锁提示。
	var photos []model.UserPhoto
	db.DB.Where("user_id = ? AND audit_status = ?", targetID, model.AuditPassed).
		Order("sort asc, id asc").Find(&photos)

	resp.PhotoTotal = len(photos)
	resp.Photos = make([]string, 0, len(photos))
	shown := 0
	for _, p := range photos {
		// 未解锁且已超出免费额度：不返回图片地址，只计数
		if !unlocked && shown >= model.FreePhotoLimit {
			resp.LockedPhotoCount++
			continue
		}
		resp.Photos = append(resp.Photos, p.URL)
		shown++
	}
	resp.PhotosLocked = resp.LockedPhotoCount > 0

	// ---- 原生家庭（L2，双重条件：对方授权 + 我已解锁）----
	// 必须由对方主动开启 FamilyVisible 才展示，这是最敏感的资料块。
	// 默认公开会劝退大量用户，所以「没填」和「没授权」都返回空。
	var peerProfile model.UserProfile
	if err := db.DB.Where("user_id = ?", targetID).First(&peerProfile).Error; err == nil {
		if unlocked && peerProfile.FamilyVisible == 1 {
			resp.ViewedFamily = true
			resp.Family = &dto.FamilyInfo{
				ParentOcc:     peerProfile.FamilyParentOcc,
				ParentOccText: model.DictLabel("family_parent_occ", int(peerProfile.FamilyParentOcc)),
				Atmosphere:    peerProfile.FamilyAtmosphere,
				AtmosText:     model.DictLabel("family_atmosphere", int(peerProfile.FamilyAtmosphere)),
				OnlyChild:     peerProfile.FamilyOnlyChild,
				OnlyChildText: model.DictLabel("family_only_child", int(peerProfile.FamilyOnlyChild)),
				City:          peerProfile.FamilyCity,
				Note:          peerProfile.FamilyNote,
			}
		}
	}

	// ---- 三观问答（L2：答免费、看答案需权益）----
	peerAnswers, _ := loadAnswers(targetID)
	resp.ThreeViewAnswered = len(peerAnswers) > 0
	resp.ThreeViewUnlocked = unlocked
	resp.ThreeView = buildThreeView(me, targetID, unlocked)

	// 与我的关系
	var acts []model.Interaction
	db.DB.Where("(from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?)",
		me.ID, targetID, targetID, me.ID).Find(&acts)
	for _, a := range acts {
		if a.Type != model.InteractionLike {
			continue
		}
		if a.FromUserID == me.ID {
			resp.IsLiked = true
		} else {
			resp.LikedMe = true
		}
	}

	if conv := findConversation(me.ID, targetID); conv != nil && conv.Status == model.ConvNormal {
		resp.IsMatched = true
		resp.ConversationID = conv.ID
	}

	recordView(me.ID, targetID)
	return resp, nil
}

// ==================== 匹配度 ====================

// matchRate 双向匹配度百分比。
//
// 只做展示，不参与排序。单向匹配（只算"对方符合我的要求"）会让用户
// 大量喜欢那些根本看不上自己的人，配对率低、挫伤积极性。
func matchRate(myReq *model.UserRequirement, myProfile *model.UserProfile,
	otherReq *model.UserRequirement, otherProfile *model.UserProfile) int {

	mine := conditionScore(myReq, otherProfile)   // 对方满足我的要求
	theirs := conditionScore(otherReq, myProfile) // 我满足对方的要求
	return int((mine*0.5 + theirs*0.5) * 100)
}

// conditionScore 一份择偶要求对一份资料的命中率。未设置的条件视为不限，算命中
func conditionScore(req *model.UserRequirement, p *model.UserProfile) float64 {
	if req == nil || p == nil {
		return 1
	}

	total, hit := 0, 0
	check := func(hasRule, ok bool) {
		if !hasRule {
			return
		}
		total++
		if ok {
			hit++
		}
	}

	check(req.AgeMin > 0 || req.AgeMax > 0,
		(req.AgeMin == 0 || p.Age >= req.AgeMin) && (req.AgeMax == 0 || p.Age <= req.AgeMax))
	check(req.HeightMin > 0 || req.HeightMax > 0,
		(req.HeightMin == 0 || p.Height >= req.HeightMin) && (req.HeightMax == 0 || p.Height <= req.HeightMax))
	check(req.EducationMin > 0, p.Education >= req.EducationMin)
	check(req.IncomeMin > 0, p.IncomeLevel >= req.IncomeMin)
	check(req.MaritalStatus != "", containsInt(parseIntList(req.MaritalStatus), int(p.MaritalStatus)))
	check(req.AcceptChildren == 0, p.HasChildren == 0)
	check(req.CityCodes != "", strings.Contains(","+req.CityCodes+",", ","+p.CityCode+","))

	if total == 0 {
		return 1
	}
	return float64(hit) / float64(total)
}

// ==================== 推荐流的条件过滤 ====================

/*
 * 推荐流的过滤策略：硬门槛 + 软容差。
 *
 * 为什么不能把 7 项要求全部当硬条件：
 * 匹配度是「双向命中率」，如果推荐流仍用全部要求做 SQL 硬过滤，那么
 * 「对方满足我的要求」那一半必然 100%，整体匹配度被锁死在 50%~100%，
 * 只剩「我满足对方」在起作用 —— 卡片上那个环就几乎不携带信息量了
 * （实测 7 个候选全打 100%）。
 *
 * 反过来也不能完全不过滤、纯按匹配度排序：
 * 匹配度分辨率很低（7 项条件只产生 4 个档位，7/11 人并列 100%），
 * 拿它当唯一排序键等于没排序；更糟的是会把「昨天还在线」的人
 * 排到「70 天没上线」的人后面 —— 婚恋场景里对方不再上线，再高的
 * 匹配度也不会有人回应。所以排序仍以活跃度为主，这里只负责把
 * 「明显不合适」的人挡掉、把「有点偏差」的人放进来。
 *
 * 于是分两类：
 *   硬门槛 —— 婚史、城市。这两项在婚恋场景里多是「不可谈」
 *             （不接受某种婚史、不接受异地），不是能折中的偏好。
 *   软容差 —— 年龄、身高、学历、收入。用户实际能接受邻近值，
 *             放宽后既扩大了池子，又让匹配度重新产生区分度。
 *
 * 子女（accept_children）也从硬条件改为软容差：它不设时表示「不限」，
 * 设了 0 表示「希望对方无孩」，但这更像偏好而非底线，交给匹配度去扣分。
 */

// 软容差常量。改动这些值会直接改变推荐池大小与匹配度分布。
const (
	ageTolerance    = 3  // 年龄上下各放宽 3 岁
	heightTolerance = 5  // 身高上下各放宽 5 cm
	levelTolerance  = 1  // 学历/收入档位允许低 1 档
)

// applyRequirement 把择偶要求翻译成推荐流的 SQL 条件（硬门槛 + 软容差）
func applyRequirement(q *gorm.DB, req *model.UserRequirement) *gorm.DB {
	if req == nil {
		return q
	}

	// ---------- 硬门槛：婚史、城市 ----------
	if vals := parseIntList(req.MaritalStatus); len(vals) > 0 {
		q = q.Where("p.marital_status IN ?", vals)
	}
	if cityCodes := parseStringList(req.CityCodes); len(cityCodes) > 0 {
		q = q.Where("p.city_code IN ?", cityCodes)
	}

	// ---------- 软容差：年龄、身高 ----------
	if req.AgeMin > 0 {
		q = q.Where("p.age >= ?", req.AgeMin-ageTolerance)
	}
	if req.AgeMax > 0 {
		q = q.Where("p.age <= ?", req.AgeMax+ageTolerance)
	}
	if req.HeightMin > 0 {
		q = q.Where("p.height >= ?", req.HeightMin-heightTolerance)
	}
	if req.HeightMax > 0 {
		q = q.Where("p.height <= ?", req.HeightMax+heightTolerance)
	}

	// ---------- 软容差：学历、收入（低 1 档仍放进来，由匹配度扣分） ----------
	if req.EducationMin > 0 {
		if floor := int(req.EducationMin) - levelTolerance; floor > 0 {
			q = q.Where("p.education >= ?", floor)
		}
	}
	if req.IncomeMin > 0 {
		if floor := int(req.IncomeMin) - levelTolerance; floor > 0 {
			q = q.Where("p.income_level >= ?", floor)
		}
	}

	return q
}

// ==================== 内部工具 ====================

func loadProfile(userID uint64) *model.UserProfile {
	var p model.UserProfile
	if err := db.DB.Where("user_id = ?", userID).First(&p).Error; err != nil {
		return nil
	}
	return &p
}

func loadRequirement(userID uint64) *model.UserRequirement {
	var r model.UserRequirement
	if err := db.DB.Where("user_id = ?", userID).First(&r).Error; err != nil {
		return nil
	}
	return &r
}

func loadRequirements(userIDs []uint64) map[uint64]*model.UserRequirement {
	out := make(map[uint64]*model.UserRequirement, len(userIDs))
	if len(userIDs) == 0 {
		return out
	}
	var list []model.UserRequirement
	db.DB.Where("user_id IN ?", userIDs).Find(&list)
	for i := range list {
		out[list[i].UserID] = &list[i]
	}
	return out
}

func rowToProfile(r cardRow) *model.UserProfile {
	return &model.UserProfile{
		UserID:        r.ID,
		Age:           r.Age,
		Height:        r.Height,
		Education:     r.Education,
		IncomeLevel:   r.IncomeLevel,
		MaritalStatus: r.MaritalStatus,
		HasChildren:   r.HasChildren,
		CityCode:      r.CityCode,
	}
}

// recordView 记录浏览。同一个人重复看只更新时间，不新增记录
func recordView(fromID, toID uint64) {
	var v model.Interaction
	err := db.DB.Where("from_user_id = ? AND to_user_id = ? AND type = ?",
		fromID, toID, model.InteractionView).First(&v).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		db.DB.Create(&model.Interaction{
			FromUserID: fromID, ToUserID: toID, Type: model.InteractionView,
		})
		return
	}
	if err == nil {
		db.DB.Model(&v).UpdateColumn("updated_at", time.Now())
	}
}

func findConversation(a, b uint64) *model.Conversation {
	lo, hi := orderPair(a, b)
	var c model.Conversation
	if err := db.DB.Where("user_a_id = ? AND user_b_id = ?", lo, hi).First(&c).Error; err != nil {
		return nil
	}
	return &c
}

// orderPair 会话双方按 id 大小归一化，保证同一对只有一条记录
func orderPair(a, b uint64) (uint64, uint64) {
	if a < b {
		return a, b
	}
	return b, a
}

func parseIntList(s string) []int {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		if v, err := strconv.Atoi(strings.TrimSpace(p)); err == nil {
			out = append(out, v)
		}
	}
	return out
}

func parseStringList(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if value := strings.TrimSpace(part); value != "" {
			out = append(out, value)
		}
	}
	return out
}

func containsInt(list []int, v int) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func activeText(t *time.Time) string {
	if t == nil {
		return ""
	}
	d := time.Since(*t)
	switch {
	case d < 10*time.Minute:
		return "刚刚活跃"
	case d < time.Hour:
		return fmt.Sprintf("%d分钟前活跃", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%d小时前活跃", int(d.Hours()))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%d天前活跃", int(d.Hours()/24))
	default:
		return "很久未活跃"
	}
}
