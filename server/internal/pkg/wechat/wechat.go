package wechat

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"hunjie/internal/config"
	"hunjie/internal/pkg/logger"
	"hunjie/internal/pkg/sensitive"
)

// Client 微信接口。Mock 与真实是同一接口的两个实现，
// 拿到 AppID/AppSecret 后只需把当前环境 YAML 里的 wechat.mock 改成 false，业务代码不动。
type Client interface {
	// Code2Session 用 wx.login 的 code 换 openid
	Code2Session(code string) (string, error)
	// GetPhoneNumber 用 getPhoneNumber 回调的 code 换手机号
	GetPhoneNumber(code string) (string, error)
	// MsgSecCheck 文本内容安全检测，返回是否通过、不通过原因
	MsgSecCheck(openID, text string) (bool, string)
	// IsMock 当前是否 mock 模式
	IsMock() bool
}

var C Client

func Init() {
	if config.C.WeChat.Mock {
		C = &mockClient{}
		logger.Info("微信客户端：Mock 模式")
		return
	}
	if config.C.WeChat.AppID == "" || config.C.WeChat.AppSecret == "" {
		logger.Fatalf("wechat.mock=false 但未配置 app_id / app_secret")
	}
	C = &realClient{}
	logger.Info("微信客户端：真实模式")
}

// ==================== Mock 实现 ====================

type mockClient struct{}

func (m *mockClient) IsMock() bool { return true }

// Code2Session mock 下把 code 直接当 openid 用。
// 开发者工具里换不同的 code 就能模拟多个账号；
// 开发环境 YAML 的超管白名单里填哪个 code，用哪个登录就是超管。
func (m *mockClient) Code2Session(code string) (string, error) {
	if code == "" {
		return "", errors.New("code 不能为空")
	}
	return code, nil
}

// GetPhoneNumber mock 下不走微信解密，由 handler 直接取前端手填的 phone
func (m *mockClient) GetPhoneNumber(code string) (string, error) {
	return "", errors.New("mock 模式请直接传 phone 参数")
}

// MsgSecCheck mock 下走本地敏感词库
func (m *mockClient) MsgSecCheck(_, text string) (bool, string) {
	return sensitive.CheckAll(text)
}

// ==================== 真实实现 ====================

type realClient struct {
	mu       sync.Mutex
	token    string
	expireAt time.Time
}

func (r *realClient) IsMock() bool { return false }

type wxError struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (r *realClient) Code2Session(code string) (string, error) {
	if code == "" {
		return "", errors.New("code 不能为空")
	}
	u := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		url.QueryEscape(config.C.WeChat.AppID),
		url.QueryEscape(config.C.WeChat.AppSecret),
		url.QueryEscape(code),
	)

	var resp struct {
		wxError
		OpenID string `json:"openid"`
	}
	if err := getJSON(u, &resp); err != nil {
		return "", err
	}
	if resp.ErrCode != 0 {
		return "", fmt.Errorf("微信登录失败(%d): %s", resp.ErrCode, resp.ErrMsg)
	}
	if resp.OpenID == "" {
		return "", errors.New("微信未返回 openid")
	}
	return resp.OpenID, nil
}

func (r *realClient) GetPhoneNumber(code string) (string, error) {
	token, err := r.accessToken()
	if err != nil {
		return "", err
	}
	u := "https://api.weixin.qq.com/wxa/business/getuserphonenumber?access_token=" + url.QueryEscape(token)

	var resp struct {
		wxError
		PhoneInfo struct {
			PurePhoneNumber string `json:"purePhoneNumber"`
		} `json:"phone_info"`
	}
	if err := postJSON(u, map[string]string{"code": code}, &resp); err != nil {
		return "", err
	}
	if resp.ErrCode != 0 {
		return "", fmt.Errorf("获取手机号失败(%d): %s", resp.ErrCode, resp.ErrMsg)
	}
	return resp.PhoneInfo.PurePhoneNumber, nil
}

// MsgSecCheck 微信内容安全。本地词库先挡一遍，再走微信接口。
// 微信接口异常时放行但记日志 —— 不能因为第三方抖动就让用户发不出消息
func (r *realClient) MsgSecCheck(openID, text string) (bool, string) {
	if ok, reason := sensitive.CheckAll(text); !ok {
		return false, reason
	}

	token, err := r.accessToken()
	if err != nil {
		logger.Errorf("内容安全检测取 access_token 失败，本次放行: %v", err)
		return true, ""
	}
	u := "https://api.weixin.qq.com/wxa/msg_sec_check?access_token=" + url.QueryEscape(token)

	body := map[string]any{
		"openid":  openID,
		"scene":   2, // 2=评论，用户资料与聊天内容都归这类
		"version": 2,
		"content": text,
	}
	var resp struct {
		wxError
		Result struct {
			Suggest string `json:"suggest"` // pass / review / risky
			Label   int    `json:"label"`
		} `json:"result"`
	}
	if err := postJSON(u, body, &resp); err != nil {
		logger.Errorf("内容安全检测请求失败，本次放行: %v", err)
		return true, ""
	}
	if resp.ErrCode != 0 {
		logger.Errorf("内容安全检测返回错误(%d: %s)，本次放行", resp.ErrCode, resp.ErrMsg)
		return true, ""
	}
	if resp.Result.Suggest == "risky" {
		return false, "内容涉嫌违规，请修改后重试"
	}
	return true, ""
}

// accessToken 带内存缓存的接口调用凭证（微信侧有效期 7200 秒）
func (r *realClient) accessToken() (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.token != "" && time.Now().Before(r.expireAt) {
		return r.token, nil
	}

	u := fmt.Sprintf(
		"https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		url.QueryEscape(config.C.WeChat.AppID),
		url.QueryEscape(config.C.WeChat.AppSecret),
	)
	var resp struct {
		wxError
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := getJSON(u, &resp); err != nil {
		return "", err
	}
	if resp.ErrCode != 0 {
		return "", fmt.Errorf("获取 access_token 失败(%d): %s", resp.ErrCode, resp.ErrMsg)
	}

	r.token = resp.AccessToken
	// 提前 5 分钟过期，避免临界点用到失效 token
	r.expireAt = time.Now().Add(time.Duration(resp.ExpiresIn-300) * time.Second)
	return r.token, nil
}

// ==================== HTTP helper ====================

var httpClient = &http.Client{Timeout: 10 * time.Second}

func getJSON(u string, out any) error {
	resp, err := httpClient.Get(u)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(out)
}

func postJSON(u string, body any, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	resp, err := httpClient.Post(u, "application/json", strings.NewReader(string(b)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(out)
}
