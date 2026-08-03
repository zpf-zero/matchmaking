package service

import "fmt"

// BizError 业务错误，Msg 是可以直接展示给用户看的文案。
//
// 与之相对，service 里用 fmt.Errorf 包装的是系统错误（数据库、网络等），
// handler 只会记日志并回一句通用提示，不把内部细节暴露给前端。
type BizError struct {
	Msg string
}

func (e *BizError) Error() string { return e.Msg }

// Biz 构造业务错误
func Biz(format string, args ...any) error {
	if len(args) == 0 {
		return &BizError{Msg: format}
	}
	return &BizError{Msg: fmt.Sprintf(format, args...)}
}
