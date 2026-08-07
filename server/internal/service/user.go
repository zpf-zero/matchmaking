package service

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"hunjie/internal/db"
	"hunjie/internal/dto"
	"hunjie/internal/model"
	"hunjie/internal/pkg/upload"
	"hunjie/internal/pkg/wechat"
)

const maxPhotoNum = 9

// GetMe 我的完整信息
func GetMe(u *model.User) (*dto.MeResp, error) {
	resp := &dto.MeResp{User: u, Phone: u.Phone}

	var p model.UserProfile
	if err := db.DB.Where("user_id = ?", u.ID).First(&p).Error; err == nil {
		resp.Profile = &p
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	var r model.UserRequirement
	if err := db.DB.Where("user_id = ?", u.ID).First(&r).Error; err == nil {
		resp.Requirement = &r
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	var photos []model.UserPhoto
	if err := db.DB.Where("user_id = ?", u.ID).Order("sort asc, id asc").Find(&photos).Error; err != nil {
		return nil, err
	}
	resp.Photos = photos

	return resp, nil
}

// SaveProfile 全量保存资料。前端在本地暂存向导数据，最后一次性提交
func SaveProfile(u *model.User, req dto.SaveProfileReq) error {
	birthday, err := time.ParseInLocation("2006-01-02", req.Birthday, time.Local)
	if err != nil {
		return Biz("生日格式不正确，应为 YYYY-MM-DD")
	}

	age := calcAge(birthday)
	if age < 18 {
		return Biz("未满 18 周岁不能使用本服务")
	}
	if age > 80 {
		return Biz("请填写正确的出生日期")
	}
	if req.Height < 140 || req.Height > 220 {
		return Biz("身高请填写 140-220cm 之间")
	}
	if len([]rune(req.Nickname)) < 2 || len([]rune(req.Nickname)) > 12 {
		return Biz("昵称请填写 2-12 个字")
	}
	if len([]rune(req.SelfIntro)) < 30 {
		return Biz("自我介绍至少 30 字，写详细些更容易被关注")
	}
	if len([]rune(req.SelfIntro)) > 500 {
		return Biz("自我介绍不能超过 500 字")
	}

	// 内容安全：昵称与自我介绍都要过
	if ok, reason := wechat.C.MsgSecCheck(u.OpenID, req.Nickname); !ok {
		return Biz("昵称不合规：%s", reason)
	}
	if ok, reason := wechat.C.MsgSecCheck(u.OpenID, req.SelfIntro); !ok {
		return Biz("自我介绍不合规：%s", reason)
	}

	// 性别注册后锁死，防止改性别刷池子
	gender := u.Gender
	if gender == 0 {
		if req.Gender != model.GenderMale && req.Gender != model.GenderFemale {
			return Biz("请选择性别")
		}
		gender = req.Gender
	}

	var photoCount int64
	db.DB.Model(&model.UserPhoto{}).Where("user_id = ?", u.ID).Count(&photoCount)

	return db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(u).Updates(map[string]any{
			"nickname": req.Nickname,
			"avatar":   req.Avatar,
			"gender":   gender,
		}).Error; err != nil {
			return err
		}
		u.Nickname, u.Avatar, u.Gender = req.Nickname, req.Avatar, gender

		var p model.UserProfile
		err := tx.Where("user_id = ?", u.ID).First(&p).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		isNew := p.ID == 0

		p.UserID = u.ID
		p.Birthday = &birthday
		p.Age = age
		p.Height = req.Height
		p.Weight = req.Weight
		p.MaritalStatus = req.MaritalStatus
		p.HasChildren = req.HasChildren
		p.Education = req.Education
		p.School = req.School
		p.Occupation = req.Occupation
		p.IncomeLevel = req.IncomeLevel
		p.HouseStatus = req.HouseStatus
		p.CarStatus = req.CarStatus
		p.Province = req.Province
		p.City = req.City
		p.CityCode = req.CityCode
		p.HomeProvince = req.HomeProvince
		p.HomeCity = req.HomeCity
		p.SelfIntro = req.SelfIntro
		p.CompleteRate = calcCompleteRate(u, &p, int(photoCount))

		// 已通过审核的用户改了资料，重新进复审队列。
		// 但 user.status 保持 2 —— 让用户无感继续使用，审核员照常复审，
		// 真发现问题再封禁。否则改一句自我介绍就要被打回，体验太差
		if p.AuditStatus == model.AuditPassed {
			p.AuditStatus = model.AuditPending
			p.AuditRemark = ""
		}

		if isNew {
			return tx.Create(&p).Error
		}
		return tx.Save(&p).Error
	})
}

// SaveRequirement 保存择偶要求
func SaveRequirement(u *model.User, req dto.SaveRequirementReq) error {
	if req.AgeMin > 0 && req.AgeMax > 0 && req.AgeMin > req.AgeMax {
		return Biz("年龄范围填反了")
	}
	if req.HeightMin > 0 && req.HeightMax > 0 && req.HeightMin > req.HeightMax {
		return Biz("身高范围填反了")
	}

	var r model.UserRequirement
	err := db.DB.Where("user_id = ?", u.ID).First(&r).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	isNew := r.ID == 0

	r.UserID = u.ID
	r.AgeMin = req.AgeMin
	r.AgeMax = req.AgeMax
	r.HeightMin = req.HeightMin
	r.HeightMax = req.HeightMax
	r.EducationMin = req.EducationMin
	r.IncomeMin = req.IncomeMin
	r.MaritalStatus = req.MaritalStatus
	r.AcceptChildren = req.AcceptChildren
	r.CityCodes = req.CityCodes
	r.Other = req.Other

	if isNew {
		return db.DB.Create(&r).Error
	}
	return db.DB.Save(&r).Error
}

// SubmitAudit 提交资料审核
func SubmitAudit(u *model.User) error {
	if u.Status == model.StatusAuditing {
		return Biz("资料正在审核中，请耐心等待")
	}
	if u.Phone == "" {
		return Biz("请先绑定手机号")
	}

	var p model.UserProfile
	if err := db.DB.Where("user_id = ?", u.ID).First(&p).Error; err != nil {
		return Biz("请先完善个人资料")
	}
	if missing := checkRequired(u, &p); missing != "" {
		return Biz("请先填写%s", missing)
	}

	var r model.UserRequirement
	if err := db.DB.Where("user_id = ?", u.ID).First(&r).Error; err != nil {
		return Biz("请先设置择偶要求")
	}

	return db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&p).Updates(map[string]any{
			"audit_status": model.AuditPending,
			"audit_remark": "",
		}).Error; err != nil {
			return err
		}
		if err := tx.Model(u).Update("status", model.StatusAuditing).Error; err != nil {
			return err
		}
		u.Status = model.StatusAuditing
		return nil
	})
}

// ==================== 相册 ====================

func AddPhoto(u *model.User, url string) (*model.UserPhoto, error) {
	var count int64
	db.DB.Model(&model.UserPhoto{}).Where("user_id = ?", u.ID).Count(&count)
	if count >= maxPhotoNum {
		return nil, Biz("最多上传 %d 张照片", maxPhotoNum)
	}

	p := model.UserPhoto{
		UserID:      u.ID,
		URL:         url,
		Sort:        int16(count),
		AuditStatus: model.AuditPending,
	}
	if err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&p).Error; err != nil {
			return err
		}
		if u.Status != model.StatusNormal {
			return nil
		}
		return tx.Model(&model.UserProfile{}).
			Where("user_id = ?", u.ID).
			Updates(map[string]any{
				"audit_status": model.AuditPending,
				"audit_remark": "",
			}).Error
	}); err != nil {
		return nil, err
	}
	return &p, nil
}

func DeletePhoto(u *model.User, photoID uint64) error {
	var p model.UserPhoto
	if err := db.DB.Where("id = ? AND user_id = ?", photoID, u.ID).First(&p).Error; err != nil {
		return Biz("照片不存在")
	}
	if err := db.DB.Delete(&p).Error; err != nil {
		return err
	}
	upload.Delete(p.URL)
	return nil
}

func SortPhotos(u *model.User, ids []uint64) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			if err := tx.Model(&model.UserPhoto{}).
				Where("id = ? AND user_id = ?", id, u.ID).
				Update("sort", i).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ==================== 内部工具 ====================

func calcAge(birthday time.Time) int8 {
	now := time.Now()
	age := now.Year() - birthday.Year()
	if now.Month() < birthday.Month() ||
		(now.Month() == birthday.Month() && now.Day() < birthday.Day()) {
		age--
	}
	return int8(age)
}

// calcCompleteRate 资料完整度。展示给用户看，也用于提示 ta 补全
func calcCompleteRate(u *model.User, p *model.UserProfile, photoCount int) int8 {
	items := []bool{
		u.Nickname != "", u.Avatar != "", u.Gender != 0,
		p.Birthday != nil, p.Height > 0, p.MaritalStatus > 0,
		p.Education > 0, p.Occupation != "", p.IncomeLevel > 0,
		p.HouseStatus > 0, p.CarStatus > 0, p.Province != "",
		p.SelfIntro != "",
		// 选填项，填了加分
		p.Weight > 0, p.School != "", p.HomeProvince != "",
		photoCount > 0, photoCount >= 3,
	}
	filled := 0
	for _, ok := range items {
		if ok {
			filled++
		}
	}
	return int8(filled * 100 / len(items))
}

// checkRequired 返回第一个缺失的必填项名称
func checkRequired(u *model.User, p *model.UserProfile) string {
	switch {
	case u.Nickname == "":
		return "昵称"
	case u.Avatar == "":
		return "头像"
	case u.Gender == 0:
		return "性别"
	case p.Birthday == nil:
		return "出生日期"
	case p.Height == 0:
		return "身高"
	case p.MaritalStatus == 0:
		return "婚姻状况"
	case p.Education == 0:
		return "学历"
	case p.Occupation == "":
		return "职业"
	case p.IncomeLevel == 0:
		return "月收入"
	case p.HouseStatus == 0:
		return "购房情况"
	case p.CarStatus == 0:
		return "购车情况"
	case p.Province == "" || p.City == "":
		return "现居城市"
	case p.SelfIntro == "":
		return "自我介绍"
	}
	return ""
}
