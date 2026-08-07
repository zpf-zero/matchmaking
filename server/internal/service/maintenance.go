package service

import (
	"time"

	"hunjie/internal/db"
	"hunjie/internal/pkg/logger"
)

func StartMaintenance() {
	refreshProfileAges()
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			refreshProfileAges()
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
