package service

import (
	"errors"
	"fmt"
	"regexp"
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

	now := time.Now()
	db.DB.Model(&u).UpdateColumn("last_active_at", now)
	u.LastActiveAt = &now

	token, err := jwt.Generate(u.ID, u.Role)
	if err != nil {
		return nil, fmt.Errorf("签发 token 失败: %w", err)
	}

	return &dto.WxLoginResp{
		Token:       token,
		User:        &u,
		Status:      u.Status,
		Role:        u.Role,
		NeedProfile: u.Status == model.StatusIncomplete,
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
