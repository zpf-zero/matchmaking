package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"hunjie/internal/dto"
	"hunjie/internal/middleware"
	"hunjie/internal/pkg/response"
	"hunjie/internal/service"
)

// AdminOverview 数据概览
// GET /api/v1/admin/overview
func AdminOverview(c *gin.Context) {
	response.OK(c, service.Overview())
}

// AdminAuditList 待审队列
// GET /api/v1/admin/audit/list
func AdminAuditList(c *gin.Context) {
	var page dto.PageQuery
	_ = c.ShouldBindQuery(&page)

	list, total, err := service.AuditList(page)
	if err != nil {
		fail(c, err)
		return
	}
	page.Normalize()
	response.Page(c, list, total, page.Page, page.PageSize)
}

// AdminAudit 审核用户资料
// POST /api/v1/admin/audit/:id
func AdminAudit(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var req dto.AuditReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数错误")
		return
	}
	if err := service.AuditUser(middleware.CurrentUser(c), id, req.Pass, req.Remark); err != nil {
		fail(c, err)
		return
	}
	response.OK(c, nil)
}

// AdminUserList 用户列表
// GET /api/v1/admin/user/list?keyword=&status=-1
func AdminUserList(c *gin.Context) {
	var page dto.PageQuery
	_ = c.ShouldBindQuery(&page)

	status := -1 // -1 表示不限
	if s := c.Query("status"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			status = v
		}
	}

	list, total, err := service.AdminUserList(c.Query("keyword"), status, page)
	if err != nil {
		fail(c, err)
		return
	}
	page.Normalize()
	response.Page(c, list, total, page.Page, page.PageSize)
}

// AdminUserDetail 用户详情
// GET /api/v1/admin/user/:id
func AdminUserDetail(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	detail, err := service.AdminUserDetail(id)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, detail)
}

// AdminBan 封禁用户
// POST /api/v1/admin/user/:id/ban
func AdminBan(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var req dto.BanReq
	_ = c.ShouldBindJSON(&req)

	if err := service.BanUser(middleware.CurrentUser(c), id, req.Reason); err != nil {
		fail(c, err)
		return
	}
	response.OK(c, nil)
}

// AdminUnban 解封用户
// POST /api/v1/admin/user/:id/unban
func AdminUnban(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	if err := service.UnbanUser(middleware.CurrentUser(c), id); err != nil {
		fail(c, err)
		return
	}
	response.OK(c, nil)
}

// AdminSetRole 任命/撤销管理角色
// POST /api/v1/admin/user/:id/role   （仅超级管理员）
func AdminSetRole(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var req dto.SetRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数错误")
		return
	}
	if err := service.SetRole(middleware.CurrentUser(c), id, req.Role); err != nil {
		fail(c, err)
		return
	}
	response.OK(c, nil)
}

// AdminReportList 举报列表
// GET /api/v1/admin/report/list?status=0
func AdminReportList(c *gin.Context) {	var page dto.PageQuery
	_ = c.ShouldBindQuery(&page)

	status := -1
	if s := c.Query("status"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			status = v
		}
	}

	list, total, err := service.AdminReportList(status, page)
	if err != nil {
		fail(c, err)
		return
	}
	page.Normalize()
	response.Page(c, list, total, page.Page, page.PageSize)
}

// AdminHandleReport 处理举报
// POST /api/v1/admin/report/:id/handle
func AdminHandleReport(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var req dto.HandleReportReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数错误")
		return
	}
	if err := service.HandleReport(middleware.CurrentUser(c), id, req.Action, req.Remark); err != nil {
		fail(c, err)
		return
	}
	response.OK(c, nil)
}
