package service

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"

	"hunjie/internal/config"
	"hunjie/internal/db"
	"hunjie/internal/dto"
	"hunjie/internal/model"
	"hunjie/internal/pkg/jwt"
	"hunjie/internal/pkg/logger"
	"hunjie/internal/pkg/wechat"
)

var phonePattern = regexp.MustCompile(`^1[3-9]\d{9}$`)

// WxLogin 微信登录。首次登录自动建号
func WxLogin(code string) (*dto.WxLoginResp, error) {
	openID, err := wechat.C.Code2Session(code)
	if err != nil {
		return nil, err
	}

	var u model.User
	err = db.DB.Where("openid = ?", openID).First(&u).Error

	switch {
	case err == nil:
		// 已有账号
	case errors.Is(err, gorm.ErrRecordNotFound):
		u = model.User{
			OpenID: openID,
			Role:   model.RoleUser,
			Status: model.StatusIncomplete,
		}
		if err := db.DB.Create(&u).Error; err != nil {
			return nil, fmt.Errorf("创建账号失败: %w", err)
		}
		logger.Infof("新用户注册: id=%d openid=%s", u.ID, openID)
	default:
		return nil, fmt.Errorf("查询账号失败: %w", err)
	}

	// 超管白名单自举：解决「没有管理员就无法任命管理员」的死锁
	if config.C.Admin.IsSuperAdmin(openID) && u.Role != model.RoleSuperAdmin {
		if err := db.DB.Model(&u).Update("role", model.RoleSuperAdmin).Error; err != nil {
			logger.Errorf("提升超管失败 uid=%d: %v", u.ID, err)
		} else {
			u.Role = model.RoleSuperAdmin
			logger.Infof("openid 命中超管白名单，已提升为超级管理员: uid=%d", u.ID)
		}
	}

	// 白名单账号首次登录时是刚建的空白号，没有昵称头像。
	// 如果放任不管，工作台和「进入用户端」都会显示空名字，很难看也不好认。
	// 只在昵称确实为空时补一个，不覆盖用户自己改过的名字。
	if u.Role > model.RoleUser && strings.TrimSpace(u.Nickname) == "" {
		nick := "管理员"
		if err := db.DB.Model(&u).Update("nickname", nick).Error; err != nil {
			logger.Errorf("运营号补默认昵称失败 uid=%d: %v", u.ID, err)
		} else {
			u.Nickname = nick
		}
	}

	now := time.Now()
	db.DB.Model(&u).UpdateColumn("last_active_at", now)
	u.LastActiveAt = &now

	// 管理员/审核员不需要走「填资料 → 审核」这条普通用户链路：
	// 他们登录只为进后台，若仍按 status==0 要求填择偶资料，会把人卡在向导里。
	// 这里顺带把被白名单提权、但状态还没扶正的运营号扶正，避免：
	//   ① 前端登录后被 profile/setup 守卫拦截；
	//   ② 出现在待审队列里被自己审；
	//   ③ 被 SetRole 提权时停在「待审核」，而队列又排除运营号 —— 永远出不来。
	// 正常(2) 和 已封禁(4) 不动：前者本就正确，后者要管理员手动解封。
	if u.Role > model.RoleUser && u.Status != model.StatusNormal && u.Status != model.StatusBanned {
		if err := db.DB.Model(&u).Update("status", model.StatusNormal).Error; err != nil {
			logger.Errorf("管理员账号状态扶正失败 uid=%d: %v", u.ID, err)
		} else {
			logger.Infof("管理员账号状态扶正为正常: uid=%d role=%d 原status=%d", u.ID, u.Role, u.Status)
			u.Status = model.StatusNormal
		}
	}

	token, err := jwt.Generate(u.ID, u.Role)
	if err != nil {
		return nil, fmt.Errorf("签发 token 失败: %w", err)
	}

	return &dto.WxLoginResp{
		Token:       token,
		User:        &u,
		Status:      u.Status,
		Role:        u.Role,
		NeedProfile: u.Role == model.RoleUser && u.Status == model.StatusIncomplete,
	}, nil
}

// BindPhone 绑定手机号。mock 模式下前端直接传 phone，真实模式走微信解密
func BindPhone(u *model.User, req dto.BindPhoneReq) (string, error) {
	var phone string

	if wechat.C.IsMock() {
		phone = req.Phone
		if phone == "" {
			return "", Biz("请填写手机号")
		}
	} else {
		if req.Code == "" {
			return "", Biz("缺少微信授权 code")
		}
		p, err := wechat.C.GetPhoneNumber(req.Code)
		if err != nil {
			return "", err
		}
		phone = p
	}

	if !phonePattern.MatchString(phone) {
		return "", Biz("手机号格式不正确")
	}
	// 手机号是防「多账号互拉刷邀请」的底层身份锚点，先做应用层查重。
	// 不直接加 UNIQUE 索引：MySQL 5.6 没有部分索引，phone 又是 NOT NULL DEFAULT ''，
	// 多个空串会互相冲突；等历史数据清洗干净后再考虑改为可空列 + 唯一索引
	var dup int64
	if err := db.DB.Model(&model.User{}).
		Where("phone = ? AND id <> ?", phone, u.ID).
		Count(&dup).Error; err != nil {
		return "", err
	}
	if dup > 0 {
		return "", Biz("该手机号已绑定其他账号，请更换手机号")
	}
	if err := db.DB.Model(u).Update("phone", phone).Error; err != nil {
		return "", fmt.Errorf("保存手机号失败: %w", err)
	}
	u.Phone = phone
	return phone, nil
}

// DeleteAccount 注销账号。软删 + 匿名化
//
// openid 上有唯一索引，软删后记录仍在表里，直接软删会导致该微信号无法重新注册。
// 这里给 openid 打上前缀释放占用，同时清空昵称头像等个人信息。
func DeleteAccount(u *model.User) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(u).Updates(map[string]any{
			"openid":     fmt.Sprintf("deleted_%d_%d", u.ID, time.Now().Unix()),
			"phone":      "",
			"nickname":   "已注销用户",
			"avatar":     "",
			"ban_reason": "",
		}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&model.User{}, u.ID).Error; err != nil {
			return err
		}
		// 资料与相册一并软删，避免注销后仍出现在他人列表里
		if err := tx.Where("user_id = ?", u.ID).Delete(&model.UserProfile{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", u.ID).Delete(&model.UserPhoto{}).Error; err != nil {
			return err
		}
		// 关闭所有会话
		return tx.Model(&model.Conversation{}).
			Where("user_a_id = ? OR user_b_id = ?", u.ID, u.ID).
			Update("status", model.ConvClosed).Error
	})
}
