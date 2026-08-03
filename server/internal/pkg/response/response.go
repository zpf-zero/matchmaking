package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 业务状态码
const (
	CodeSuccess      = 0
	CodeParamError   = 400
	CodeUnauthorized = 401
	CodeForbidden    = 403
	CodeNotFound     = 404
	CodeServerError  = 500

	// 用户状态相关，前端据此跳转
	CodeProfileIncomplete = 4001 // 资料未完善 → 跳资料填写向导
	CodeProfileAuditing   = 4002 // 资料审核中
	CodeProfileRejected   = 4003 // 资料审核未通过
	CodeBanned            = 4004 // 账号已封禁
)

type Body struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// PageData 分页响应结构
type PageData struct {
	List     any   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Body{Code: CodeSuccess, Message: "ok", Data: data})
}

func Page(c *gin.Context, list any, total int64, page, pageSize int) {
	OK(c, PageData{List: list, Total: total, Page: page, PageSize: pageSize})
}

// Fail 返回业务错误。HTTP 状态码统一 200，靠 code 区分
func Fail(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Body{Code: code, Message: message, Data: nil})
}

// Abort 返回业务错误并中断后续 handler，中间件里用
func Abort(c *gin.Context, code int, message string) {
	c.AbortWithStatusJSON(http.StatusOK, Body{Code: code, Message: message, Data: nil})
}

func ParamError(c *gin.Context, message string) {
	if message == "" {
		message = "参数错误"
	}
	Fail(c, CodeParamError, message)
}

func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "请先登录"
	}
	Abort(c, CodeUnauthorized, message)
}

func Forbidden(c *gin.Context, message string) {
	if message == "" {
		message = "没有操作权限"
	}
	Abort(c, CodeForbidden, message)
}

func NotFound(c *gin.Context, message string) {
	if message == "" {
		message = "资源不存在"
	}
	Fail(c, CodeNotFound, message)
}

func ServerError(c *gin.Context, message string) {
	if message == "" {
		message = "服务器开小差了，请稍后再试"
	}
	Fail(c, CodeServerError, message)
}
