package service

import (
	"time"

	"hunjie/internal/db"
	"hunjie/internal/model"
	"hunjie/internal/pkg/logger"
)

func StartMaintenance() {
	refreshProfileAges()
	cleanupSmsCodes()
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			refreshProfileAges()
			cleanupSmsCodes()
		}
	}()
}

func refreshProfileAges() {
	result := db.DB.Exec(`
		UPDATE user_profile
		SET age = TIMESTAMPDIFF(YEAR, birthday, CURDATE())
		WHERE deleted_at IS NULL
		  AND birthday IS NOT NULL
		  AND age <> TIMESTAMPDIFF(YEAR, birthday, CURDATE())
	`)
	if result.Error != nil {
		logger.Errorf("刷新用户年龄失败: %v", result.Error)
		return
	}
	if result.RowsAffected > 0 {
		logger.Infof("已刷新用户年龄: %d", result.RowsAffected)
	}
}

// cleanupSmsCodes 清理 7 天前的验证码记录。
// 发送频次校验只看最近 24 小时，历史记录留着只会让表无限增长
func cleanupSmsCodes() {
	cutoff := time.Now().AddDate(0, 0, -7)
	result := db.DB.Where("created_at < ?", cutoff).Delete(&model.SmsCode{})
	if result.Error != nil {
		logger.Errorf("清理过期验证码失败: %v", result.Error)
		return
	}
	if result.RowsAffected > 0 {
		logger.Infof("已清理过期验证码: %d", result.RowsAffected)
	}
}
