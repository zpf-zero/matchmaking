package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// C 全局配置
var C Config

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	MySQL  MySQLConfig  `mapstructure:"mysql"`
	JWT    JWTConfig    `mapstructure:"jwt"`
	WeChat WeChatConfig `mapstructure:"wechat"`
	Upload UploadConfig `mapstructure:"upload"`
	Admin  AdminConfig  `mapstructure:"admin"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type MySQLConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
}

type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	ExpireDays int    `mapstructure:"expire_days"`
}

type WeChatConfig struct {
	Mock      bool   `mapstructure:"mock"`
	AppID     string `mapstructure:"app_id"`
	AppSecret string `mapstructure:"app_secret"`
}

type UploadConfig struct {
	Dir       string `mapstructure:"dir"`
	BaseURL   string `mapstructure:"base_url"`
	MaxSizeMB int    `mapstructure:"max_size_mb"`
}

type AdminConfig struct {
	SuperAdminOpenIDs []string `mapstructure:"super_admin_openids"`
}

// Load 从指定 YAML 文件加载配置
func Load(path string) error {
	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}
	if err := viper.Unmarshal(&C); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}
	return nil
}

// DSN MySQL 连接串
func (m MySQLConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		m.User, m.Password, m.Host, m.Port, m.Database)
}

// IsSuperAdmin 判断 openid 是否在超管白名单内
func (a AdminConfig) IsSuperAdmin(openID string) bool {
	for _, id := range a.SuperAdminOpenIDs {
		if id == openID {
			return true
		}
	}
	return false
}
