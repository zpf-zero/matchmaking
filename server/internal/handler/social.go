package handler

import (
	"github.com/gin-gonic/gin"

	"hunjie/internal/dto"
	"hunjie/internal/middleware"
	"hunjie/internal/pkg/response"
	"hunjie/internal/service"
)

// CreateReport 举报
// POST /api/v1/common/report
func CreateReport(c *gin.Context) {
	var req dto.ReportReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "请选择举报原因")
		return
	}
	if err := service.CreateReport(middleware.CurrentUser(c), req); err != nil {
		fail(c, err)
		return
	}
	response.OK(c, nil)
}

// AddBlacklist 拉黑
// POST /api/v1/common/blacklist
func AddBlacklist(c *gin.Context) {
	var req dto.TargetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数错误")
		return
	}
	if err := service.AddBlacklist(middleware.CurrentUser(c), req.ToUserID); err != nil {
		fail(c, err)
		return
	}
	response.OK(c, nil)
}

// RemoveBlacklist 取消拉黑
// DELETE /api/v1/common/blacklist/:id
func RemoveBlacklist(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	if err := service.RemoveBlacklist(middleware.CurrentUser(c), id); err != nil {
		fail(c, err)
		return
	}
	response.OK(c, nil)
}

// ListBlacklist 黑名单列表
// GET /api/v1/common/blacklist
func ListBlacklist(c *gin.Context) {
	list, err := service.ListBlacklist(middleware.CurrentUser(c))
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, list)
}
