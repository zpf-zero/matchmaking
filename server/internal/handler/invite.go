package handler

import (
	"github.com/gin-gonic/gin"

	"hunjie/internal/dto"
	"hunjie/internal/middleware"
	"hunjie/internal/pkg/response"
	"hunjie/internal/service"
)

// MyInvite 我的邀请码与邀请进度
// GET /api/v1/invite/my-code
func MyInvite(c *gin.Context) {
	info, err := service.MyInvite(middleware.CurrentUser(c))
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, info)
}

// BindInvite 提交邀请码，绑定邀请关系
// POST /api/v1/invite/bind
func BindInvite(c *gin.Context) {
	var req dto.BindInviteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "请填写邀请码")
		return
	}
	if err := service.BindInvite(middleware.CurrentUser(c), req.Code); err != nil {
		fail(c, err)
		return
	}
	response.OK(c, nil)
}

// MarkInviteSeen 邀请人查看结算结果，清除红点
// POST /api/v1/invite/seen
func MarkInviteSeen(c *gin.Context) {
	if err := service.MarkInviteSeen(middleware.CurrentUser(c)); err != nil {
		fail(c, err)
		return
	}
	response.OK(c, nil)
}
