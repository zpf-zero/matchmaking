package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// C 全局配置
var C Config

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	MySQL  MySQLConfig  `mapstructure:"mysql"`
	JWT    JWTConfig    `mapstructure:"jwt"`
	WeChat WeChatConfig `mapstructure:"wechat"`
	SMS    SMSConfig    `mapstructure:"sms"`
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

// SMSConfig 短信验证码通道。H5 手机号登录使用。
//
// mock 模式下验证码不真实下发：开发环境由 /auth/sms/send 回传 dev_code，
// 生产环境（server.mode=release）只写服务端日志，绝不回传 ——
// 否则等于开放「任意手机号免验证码登录」。
type SMSConfig struct {
	Mock            bool   `mapstructure:"mock"`
	Provider        string `mapstructure:"provider"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret"`
	SignName        string `mapstructure:"sign_name"`
	TemplateCode    string `mapstructure:"template_code"`
	RegionID        string `mapstructure:"region_id"`
	CodeTTLSeconds  int    `mapstructure:"code_ttl_seconds"`
	ResendSeconds   int    `mapstructure:"resend_seconds"`
	DailyLimit      int    `mapstructure:"daily_limit"`
	IPHourlyLimit   int    `mapstructure:"ip_hourly_limit"`
}

func (s SMSConfig) CodeTTL() time.Duration {
	if s.CodeTTLSeconds <= 0 {
		return 5 * time.Minute
	}
	return time.Duration(s.CodeTTLSeconds) * time.Second
}

func (s SMSConfig) ResendInterval() time.Duration {
	if s.ResendSeconds <= 0 {
		return 60 * time.Second
	}
	return time.Duration(s.ResendSeconds) * time.Second
}

func (s SMSConfig) DailyLimitOrDefault() int {
	if s.DailyLimit <= 0 {
		return 5
	}
	return s.DailyLimit
}

func (s SMSConfig) IPHourlyLimitOrDefault() int {
	if s.IPHourlyLimit <= 0 {
		return 20
	}
	return s.IPHourlyLimit
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
