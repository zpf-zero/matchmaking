package sensitive

import (
	"bufio"
	"os"
	"regexp"
	"strings"
	"sync"
)

// 一期用本地词库做兜底过滤。拿到 AppSecret 后，文本走微信 msgSecCheck，
// 这里退化为第一道快速筛子（微信接口有频率限制，本地先挡一遍更稳）。

var (
	mu    sync.RWMutex
	words []string
)

// 内置基础词库。婚恋场景主要防三类：色情、诈骗引流、辱骂
var defaultWords = []string{
	// 色情低俗
	"约炮", "一夜情", "包养", "援交", "裸聊", "色情", "黄片", "卖淫", "嫖娼",
	// 诈骗引流
	"杀猪盘", "刷单", "返利", "博彩", "赌博", "澳门赌场", "网贷", "套现",
	"投资理财群", "带你赚钱", "内幕消息", "稳赚不赔",
	// 辱骂
	"傻逼", "煞笔", "傻B", "垃圾玩意", "贱人", "婊子", "去死",
	// 违禁
	"办证", "代开发票", "枪支", "毒品", "冰毒", "大麻",
}

// 联系方式正则。婚恋产品里，资料页留联系方式通常是诈骗或引流
var (
	phoneRe  = regexp.MustCompile(`1[3-9]\d{9}`)
	qqRe     = regexp.MustCompile(`(?i)(qq|扣扣|企鹅)\s*[:：]?\s*\d{5,12}`)
	wechatRe = regexp.MustCompile(`(?i)(微信|weixin|wx|vx|v信)\s*[:：]?\s*[a-zA-Z][-_a-zA-Z0-9]{5,19}`)
)

// Init 加载词库。path 为空或文件不存在时用内置词库
func Init(path string) {
	mu.Lock()
	defer mu.Unlock()

	words = append([]string{}, defaultWords...)
	if path == "" {
		return
	}
	f, err := os.Open(path)
	if err != nil {
		return // 文件不存在就只用内置词库，不算错误
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		w := strings.TrimSpace(sc.Text())
		if w != "" && !strings.HasPrefix(w, "#") {
			words = append(words, w)
		}
	}
}

// Check 检查文本是否合规。返回是否通过、命中的词
func Check(text string) (bool, string) {
	if text == "" {
		return true, ""
	}
	lower := strings.ToLower(text)

	mu.RLock()
	defer mu.RUnlock()
	for _, w := range words {
		if strings.Contains(lower, strings.ToLower(w)) {
			return false, w
		}
	}
	return true, ""
}

// CheckContact 检查是否含联系方式
func CheckContact(text string) (bool, string) {
	switch {
	case phoneRe.MatchString(text):
		return false, "手机号"
	case qqRe.MatchString(text):
		return false, "QQ号"
	case wechatRe.MatchString(text):
		return false, "微信号"
	}
	return true, ""
}

// CheckAll 敏感词 + 联系方式一起检查，返回不通过的原因
func CheckAll(text string) (bool, string) {
	if ok, hit := Check(text); !ok {
		return false, "内容包含敏感词「" + hit + "」"
	}
	if ok, kind := CheckContact(text); !ok {
		return false, "为保障双方安全，请勿填写" + kind + "，配对成功后可在站内聊天"
	}
	return true, ""
}
