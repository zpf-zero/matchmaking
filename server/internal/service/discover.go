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
}

const cardFields = `u.id, u.nickname, u.avatar, u.gender, u.last_active_at,
	p.age, p.height, p.weight, p.education, p.school, p.occupation, p.income_level,
	p.marital_status, p.has_children, p.house_status, p.car_status,
	p.province, p.city, p.city_code, p.home_province, p.home_city, p.self_intro`

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

// Recommend 推荐流。用我的择偶要求做过滤，按活跃时间倒序
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
		Order("u.last_active_at DESC, u.id DESC").
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
			MatchRate:     matchRate(myReq, myProfile, otherReqs[r.ID], rowToProfile(r)),
			ActiveText:    activeText(r.LastActiveAt),
		})
	}
	return cards
}

// GetUserDetail 他人资料详情。顺带记录一条浏览
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

	// 相册只展示审核通过的
	var photos []model.UserPhoto
	db.DB.Where("user_id = ? AND audit_status = ?", targetID, model.AuditPassed).
		Order("sort asc, id asc").Find(&photos)
	resp.Photos = make([]string, 0, len(photos))
	for _, p := range photos {
		resp.Photos = append(resp.Photos, p.URL)
	}

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

// applyRequirement 把择偶要求翻译成 SQL 条件
func applyRequirement(q *gorm.DB, req *model.UserRequirement) *gorm.DB {
	if req == nil {
		return q
	}
	if req.AgeMin > 0 {
		q = q.Where("p.age >= ?", req.AgeMin)
	}
	if req.AgeMax > 0 {
		q = q.Where("p.age <= ?", req.AgeMax)
	}
	if req.HeightMin > 0 {
		q = q.Where("p.height >= ?", req.HeightMin)
	}
	if req.HeightMax > 0 {
		q = q.Where("p.height <= ?", req.HeightMax)
	}
	if req.EducationMin > 0 {
		q = q.Where("p.education >= ?", req.EducationMin)
	}
	if req.IncomeMin > 0 {
		q = q.Where("p.income_level >= ?", req.IncomeMin)
	}
	if vals := parseIntList(req.MaritalStatus); len(vals) > 0 {
		q = q.Where("p.marital_status IN ?", vals)
	}
	if req.AcceptChildren == 0 {
		q = q.Where("p.has_children = 0")
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
