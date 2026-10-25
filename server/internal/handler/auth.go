package handler

import (
	"github.com/gin-gonic/gin"

	"hunjie/internal/dto"
	"hunjie/internal/middleware"
	"hunjie/internal/pkg/response"
	"hunjie/internal/service"
)

// WxLogin 微信登录
// POST /api/v1/auth/wx-login
func WxLogin(c *gin.Context) {
	var req dto.WxLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "缺少 code")
		return
	}

	resp, err := service.WxLogin(req.Code)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, resp)
}

// SendSmsCode 发送 H5 登录验证码
// POST /api/v1/auth/sms/send
func SendSmsCode(c *gin.Context) {
	var req dto.SendSmsCodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "请填写手机号")
		return
	}

	resp, err := service.SendLoginCode(req.Phone, c.ClientIP())
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, resp)
}

// PhoneLogin 手机号验证码登录，手机号不存在时自动建号
// POST /api/v1/auth/sms/login
func PhoneLogin(c *gin.Context) {
	var req dto.PhoneLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "请填写手机号和验证码")
		return
	}

	resp, err := service.PhoneLogin(req.Phone, req.Code, c.ClientIP())
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, resp)
}

// BindPhone 绑定手机号
// POST /api/v1/auth/bind-phone
func BindPhone(c *gin.Context) {
	var req dto.BindPhoneReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数错误")
		return
	}

	phone, err := service.BindPhone(middleware.CurrentUser(c), req)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, gin.H{"phone": phone})
}

// Logout 退出登录。JWT 无状态，服务端不需要做什么，前端清 token 即可
// POST /api/v1/auth/logout
func Logout(c *gin.Context) {
	response.OK(c, nil)
}

// DeleteAccount 注销账号
// DELETE /api/v1/auth/account
func DeleteAccount(c *gin.Context) {
	if err := service.DeleteAccount(middleware.CurrentUser(c)); err != nil {
		fail(c, err)
		return
	}
	response.OK(c, nil)
}
