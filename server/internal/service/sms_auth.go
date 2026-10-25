package service

/*
 * H5 手机号验证码登录。
 *
 * 与小程序微信登录并列的第二种登录方式，适用于没有 openid 的 H5 场景。
 * 安全边界：
 *   · 发送接口未登录可调，必须限频（单手机号 60s + 24h 上限，单 IP 小时上限）；
 *   · 验证码只存 HMAC 哈希，5 分钟过期、5 次错误上限、用后即焚；
 *   · 登录按手机号找账号，找不到才建号，已有微信账号绑定过同一手机号时直接复用；
 *   · 生产环境不回传 dev_code（见 SendLoginCode）。
 */

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"

	"gorm.io/gorm"

	"hunjie/internal/config"
	"hunjie/internal/db"
	"hunjie/internal/dto"
	"hunjie/internal/model"
	"hunjie/internal/pkg/jwt"
	"hunjie/internal/pkg/logger"
	"hunjie/internal/pkg/sms"
)

const (
	smsSceneLogin  = "login"
	smsMaxAttempts = 5
)

// SendLoginCode 发送登录验证码
func SendLoginCode(rawPhone, ip string) (*dto.SendSmsCodeResp, error) {
	phone := normalizePhone(rawPhone)
	if !phonePattern.MatchString(phone) {
		return nil, Biz("手机号格式不正确")
	}

	now := time.Now()
	ttl := config.C.SMS.CodeTTL()
	resend := config.C.SMS.ResendInterval()

	var last model.SmsCode
	err := db.DB.Where("phone = ? AND scene = ?", phone, smsSceneLogin).
		Order("id DESC").First(&last).Error
	if err == nil {
		if wait := resend - now.Sub(last.CreatedAt); wait > 0 {
			return nil, Biz("发送太频繁，请 %d 秒后再试", int(math.Ceil(wait.Seconds())))
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	var daily int64
	if err := db.DB.Model(&model.SmsCode{}).
		Where("phone = ? AND scene = ? AND created_at >= ?", phone, smsSceneLogin, now.Add(-24*time.Hour)).
		Count(&daily).Error; err != nil {
		return nil, err
	}
	if daily >= int64(config.C.SMS.DailyLimitOrDefault()) {
		return nil, Biz("该手机号今日获取验证码次数已达上限，请明天再试")
	}

	if ip != "" {
		var hourly int64
		if err := db.DB.Model(&model.SmsCode{}).
			Where("ip = ? AND created_at >= ?", ip, now.Add(-time.Hour)).
			Count(&hourly).Error; err != nil {
			return nil, err
		}
		if hourly >= int64(config.C.SMS.IPHourlyLimitOrDefault()) {
			return nil, Biz("当前网络请求过于频繁，请稍后再试")
		}
	}

	code, err := randomSMSCode()
	if err != nil {
		return nil, err
	}

	rec := model.SmsCode{
		Phone:     phone,
		CodeHash:  hashSMSCode(phone, code),
		Scene:     smsSceneLogin,
		IP:        ip,
		ExpiresAt: now.Add(ttl),
	}
	if err := db.DB.Create(&rec).Error; err != nil {
		return nil, fmt.Errorf("保存验证码失败: %w", err)
	}

	if err := sms.C.Send(phone, code); err != nil {
		// 发送失败要立即删掉记录，否则 60 秒冷却会把用户挡在重试之外
		db.DB.Delete(&rec)
		logger.Errorf("短信发送失败 phone=%s: %v", maskPhone(phone), err)
		return nil, Biz("验证码发送失败，请稍后重试")
	}

	logger.Infof("登录验证码已发送 phone=%s ip=%s", maskPhone(phone), ip)

	resp := &dto.SendSmsCodeResp{
		ExpiresIn: int(ttl.Seconds()),
		ResendIn:  int(resend.Seconds()),
	}
	// 只在开发环境回传验证码。生产即使 sms.mock=true（凭证未到位）也不能回传，
	// 否则任何调用方都能拿到任意手机号的验证码，等于开放「任意手机号登录」。
	if config.C.SMS.Mock && config.C.Server.Mode != "release" {
		resp.DevCode = code
	}
	return resp, nil
}

// PhoneLogin 手机号验证码登录。手机号不存在时自动建号
func PhoneLogin(rawPhone, code, ip string) (*dto.WxLoginResp, error) {
	phone := normalizePhone(rawPhone)
	code = strings.TrimSpace(code)
	if !phonePattern.MatchString(phone) {
		return nil, Biz("手机号格式不正确")
	}
	if len(code) != 6 {
		return nil, Biz("请输入 6 位验证码")
	}

	now := time.Now()
	var rec model.SmsCode
	err := db.DB.Where("phone = ? AND scene = ? AND used_at IS NULL AND expires_at > ?",
		phone, smsSceneLogin, now).Order("id DESC").First(&rec).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, Biz("验证码已过期，请重新获取")
		}
		return nil, err
	}
	if rec.Attempts >= smsMaxAttempts {
		return nil, Biz("验证码错误次数过多，请重新获取")
	}
	if subtle.ConstantTimeCompare([]byte(rec.CodeHash), []byte(hashSMSCode(phone, code))) != 1 {
		db.DB.Model(&model.SmsCode{}).Where("id = ?", rec.ID).
			UpdateColumn("attempts", gorm.Expr("attempts + 1"))
		return nil, Biz("验证码不正确")
	}

	// 原子消费：同一个码并发提交时只有一个请求能成功
	res := db.DB.Model(&model.SmsCode{}).
		Where("id = ? AND used_at IS NULL", rec.ID).
		Update("used_at", now)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, Biz("验证码已使用，请重新获取")
	}

	u, err := findOrCreatePhoneUser(phone)
	if err != nil {
		return nil, err
	}

	db.DB.Model(u).UpdateColumn("last_active_at", now)
	u.LastActiveAt = &now

	token, err := jwt.Generate(u.ID, u.Role)
	if err != nil {
		return nil, fmt.Errorf("签发 token 失败: %w", err)
	}

	logger.Infof("手机号登录: uid=%d phone=%s ip=%s", u.ID, maskPhone(phone), ip)
	return &dto.WxLoginResp{
		Token:       token,
		User:        u,
		Status:      u.Status,
		Role:        u.Role,
		NeedProfile: u.Role == model.RoleUser && u.Status == model.StatusIncomplete,
	}, nil
}

// findOrCreatePhoneUser 按手机号找账号，不存在则建号。
// 已有微信账号绑定过同一手机号时，H5 与小程序共用同一个账号。
func findOrCreatePhoneUser(phone string) (*model.User, error) {
	var users []model.User
	if err := db.DB.Where("phone = ?", phone).Order("id ASC").Limit(2).Find(&users).Error; err != nil {
		return nil, err
	}
	switch len(users) {
	case 1:
		return &users[0], nil
	case 0:
	default:
		// 历史数据可能存在重复手机号（唯一性校验是后加的），先让人工处理
		return nil, Biz("该手机号关联了多个账号，请联系客服处理")
	}

	// H5 没有 openid。用 h5_<手机号> 作为稳定登录标识占住 openid 唯一索引，
	// 后续若做微信账号合并，再演进为独立的绑定关系表。
	u := model.User{
		OpenID: "h5_" + phone,
		Phone:  phone,
		Role:   model.RoleUser,
		Status: model.StatusIncomplete,
	}
	if err := db.DB.Create(&u).Error; err != nil {
		if isDuplicateKey(err) {
			// 并发首次登录：另一个请求已经建号，取回即可
			if e := db.DB.Where("phone = ?", phone).First(&u).Error; e == nil {
				return &u, nil
			}
		}
		return nil, fmt.Errorf("创建账号失败: %w", err)
	}
	logger.Infof("手机号注册: uid=%d phone=%s", u.ID, maskPhone(phone))
	return &u, nil
}

func normalizePhone(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.TrimPrefix(s, "+86")
	return s
}

func randomSMSCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", fmt.Errorf("生成验证码失败: %w", err)
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// hashSMSCode 6 位数字空间很小，用 JWT secret 作 pepper，
// 避免库文件泄露后被直接读出可用验证码
func hashSMSCode(phone, code string) string {
	mac := hmac.New(sha256.New, []byte(config.C.JWT.Secret))
	mac.Write([]byte(phone + ":" + code))
	return hex.EncodeToString(mac.Sum(nil))
}

func maskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}
