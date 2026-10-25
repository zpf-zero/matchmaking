package sms

/*
 * 短信通道。H5 手机号验证码登录使用。
 *
 * 设计要点：
 *   · Mock 与真实通道是同一 Sender 接口的两个实现，切换只改配置，业务代码不动；
 *   · Mock 模式下验证码只写日志；开发环境还会由 /auth/sms/send 回传 dev_code，
 *     生产（server.mode=release）绝不回传，否则等于开放「任意手机号登录」；
 *   · 真实通道当前实现阿里云 Dysmsapi（RPC 签名 v1），
 *     需申请签名与模板（docs/07-待申请事项清单.md 第 7.1 节）。
 */

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"hunjie/internal/config"
	"hunjie/internal/pkg/logger"
)

// Sender 短信发送器
type Sender interface {
	Send(phone, code string) error
}

var C Sender

func Init() {
	cfg := config.C.SMS

	if cfg.Mock {
		C = mockSender{}
		if config.C.Server.Mode == "release" {
			logger.Warnf("短信通道为 mock 但当前是 release：验证码不会真实下发，H5 登录仅能从服务端日志取码，上线前必须配置真实通道")
		} else {
			logger.Info("短信通道：Mock 模式（验证码经 dev_code 回传）")
		}
		return
	}

	provider := cfg.Provider
	if provider == "" {
		provider = "aliyun"
	}
	if provider != "aliyun" {
		logger.Errorf("不支持的短信通道 provider=%s，验证码将无法发送", provider)
		C = unconfiguredSender{}
		return
	}
	if cfg.AccessKeyID == "" || cfg.AccessKeySecret == "" ||
		cfg.SignName == "" || cfg.TemplateCode == "" {
		logger.Errorf("短信通道配置不完整（access_key_id / access_key_secret / sign_name / template_code），验证码将无法发送")
		C = unconfiguredSender{}
		return
	}

	region := cfg.RegionID
	if region == "" {
		region = "cn-hangzhou"
	}
	C = &aliyunSender{
		keyID:      cfg.AccessKeyID,
		keySecret:  cfg.AccessKeySecret,
		signName:   cfg.SignName,
		template:   cfg.TemplateCode,
		regionID:   region,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
	logger.Info("短信通道：阿里云短信")
}

// ==================== Mock ====================

type mockSender struct{}

func (mockSender) Send(phone, code string) error {
	logger.Infof("【Mock 短信】%s 的验证码是 %s（%s 内有效）",
		maskPhone(phone), code, config.C.SMS.CodeTTL())
	return nil
}

// ==================== 未配置 ====================

type unconfiguredSender struct{}

func (unconfiguredSender) Send(_, _ string) error {
	return fmt.Errorf("短信通道未配置")
}

// ==================== 阿里云 ====================

type aliyunSender struct {
	keyID      string
	keySecret  string
	signName   string
	template   string
	regionID   string
	httpClient *http.Client
}

func (s *aliyunSender) Send(phone, code string) error {
	params := map[string]string{
		"AccessKeyId":      s.keyID,
		"Action":           "SendSms",
		"Format":           "JSON",
		"PhoneNumbers":     phone,
		"RegionId":         s.regionID,
		"SignName":         s.signName,
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureNonce":   nonce(),
		"SignatureVersion": "1.0",
		"TemplateCode":     s.template,
		"TemplateParam":    fmt.Sprintf(`{"code":"%s"}`, code),
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"Version":          "2017-05-25",
	}

	canonical := canonicalizedQuery(params)
	stringToSign := "GET&%2F&" + percentEncode(canonical)
	mac := hmac.New(sha1.New, []byte(s.keySecret+"&"))
	mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	endpoint := "https://dysmsapi.aliyuncs.com/?Signature=" +
		percentEncode(signature) + "&" + canonical

	resp, err := s.httpClient.Get(endpoint)
	if err != nil {
		return fmt.Errorf("请求短信网关失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("读取短信网关响应失败: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("短信网关返回 HTTP %d: %s", resp.StatusCode, string(body))
	}

	var r struct {
		Code    string `json:"Code"`
		Message string `json:"Message"`
	}
	if err := json.Unmarshal(body, &r); err != nil {
		return fmt.Errorf("解析短信网关响应失败: %w", err)
	}
	if r.Code != "OK" {
		return fmt.Errorf("短信发送失败: %s (%s)", r.Message, r.Code)
	}
	return nil
}

// canonicalizedQuery 按参数名排序并做 RFC3986 编码，供签名使用
func canonicalizedQuery(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, percentEncode(k)+"="+percentEncode(params[k]))
	}
	return strings.Join(parts, "&")
}

// percentEncode 阿里云要求的 RFC3986 编码
func percentEncode(s string) string {
	e := url.QueryEscape(s)
	e = strings.ReplaceAll(e, "+", "%20")
	e = strings.ReplaceAll(e, "*", "%2A")
	e = strings.ReplaceAll(e, "%7E", "~")
	return e
}

func nonce() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}

func maskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}
