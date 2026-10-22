package router

import (
	"os"

	"github.com/gin-gonic/gin"

	"hunjie/internal/config"
	"hunjie/internal/handler"
	"hunjie/internal/middleware"
	"hunjie/internal/model"
)

func Setup() *gin.Engine {
	gin.SetMode(gin.DebugMode)
	if config.C.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery(), middleware.Logger(), middleware.CORS())

	// 上传的图片直接走静态路由
	_ = os.MkdirAll(config.C.Upload.Dir, 0o755)
	r.Static("/uploads", config.C.Upload.Dir)

	r.GET("/health", handler.Health)

	v1 := r.Group("/api/v1")

	// ---------- 无需登录 ----------
	{
		v1.POST("/auth/wx-login", handler.WxLogin)
		v1.GET("/common/dict", handler.Dict)
		v1.GET("/common/config", handler.ClientConfig)
		// 三观题库：登录即可取（答题免费，题目本身无隐私）
		v1.GET("/common/three-view/questions", handler.ThreeViewQuestions)
	}

	// ---------- 登录即可访问：资料填写、上传 ----------
	// 不挂 RequireBrowse/RequireActive，否则新用户没法填资料
	authed := v1.Group("", middleware.Auth())
	{
		authed.POST("/auth/bind-phone", handler.BindPhone)
		authed.POST("/auth/logout", handler.Logout)
		authed.DELETE("/auth/account", handler.DeleteAccount)

		authed.GET("/user/me", handler.Me)
		authed.PUT("/user/profile", handler.SaveProfile)
		authed.PUT("/user/requirement", handler.SaveRequirement)
		authed.POST("/user/submit-audit", handler.SubmitAudit)
		authed.POST("/user/photo", handler.AddPhoto)
		authed.PUT("/user/photo/sort", handler.SortPhotos)
		authed.DELETE("/user/photo/:id", handler.DeletePhoto)

		// 三观问答：答题免费（看答案才需权益），故挂在 authed 而非 browse
		authed.GET("/user/three-view", handler.MyThreeView)
		authed.PUT("/user/three-view", handler.SaveThreeView)

		authed.POST("/common/upload/image", handler.UploadImage)

		// 邀请码：登录即可。绑定发生在资料提交审核之前 —— 权益要等
		// 被邀请人通过人工审核才结算，所以这里不挂 RequireBrowse/Active
		authed.GET("/invite/my-code", handler.MyInvite)
		authed.POST("/invite/bind", handler.BindInvite)
		// 邀请人查看结算结果，清除「我的」页红点
		authed.POST("/invite/seen", handler.MarkInviteSeen)
	}

	// ---------- 资料已提交：可浏览他人 ----------
	browse := v1.Group("", middleware.Auth(), middleware.RequireBrowse())
	{
		browse.GET("/discover/recommend", handler.Recommend)
		browse.GET("/discover/search", handler.Search)
		browse.GET("/user/detail/:id", handler.UserDetail)

		browse.GET("/interaction/liked-me", handler.LikedMe)
		browse.GET("/interaction/my-likes", handler.MyLikes)
		browse.GET("/interaction/viewed-me", handler.ViewedMe)
		browse.GET("/interaction/unread-count", handler.UnreadCount)
		browse.PUT("/interaction/read", handler.MarkRead)
	}

	// ---------- 审核通过：可发起互动 ----------
	active := v1.Group("", middleware.Auth(), middleware.RequireActive())
	{
		active.POST("/interaction/like", handler.Like)
		active.POST("/interaction/pass", handler.Pass)
		active.DELETE("/interaction/like/:id", handler.Unlike)

		active.GET("/conversation/list", handler.ListConversations)
		active.DELETE("/conversation/:id", handler.CloseConversation)
		active.GET("/message/list", handler.ListMessages)
		active.POST("/message/send", handler.SendMessage)
		active.PUT("/message/read", handler.ReadMessages)
		active.PUT("/message/recall", handler.RecallMessage)
	}

	// WebSocket。小程序握手无法带自定义 header，token 从 query 取
	r.GET("/ws", middleware.Auth(), middleware.RequireActive(), handler.WebSocket)

	// ---------- 举报 / 拉黑：登录即可，被拒审的用户也要能举报 ----------
	{
		authed.POST("/common/report", handler.CreateReport)
		authed.POST("/common/blacklist", handler.AddBlacklist)
		authed.DELETE("/common/blacklist/:id", handler.RemoveBlacklist)
		authed.GET("/common/blacklist", handler.ListBlacklist)
	}

	// ---------- 管理端：同一套后端，靠 RequireRole 区分 ----------
	// 没有独立的管理员登录接口，管理员就是 role>1 的普通用户
	auditor := v1.Group("/admin", middleware.Auth(), middleware.RequireRole(model.RoleAuditor))
	{
		auditor.GET("/overview", handler.AdminOverview)
		auditor.GET("/audit/list", handler.AdminAuditList)
		auditor.POST("/audit/:id", handler.AdminAudit)
		auditor.GET("/user/list", handler.AdminUserList)
		auditor.GET("/user/:id", handler.AdminUserDetail)
		auditor.GET("/report/list", handler.AdminReportList)
		auditor.POST("/report/:id/handle", handler.AdminHandleReport)
	}

	admin := v1.Group("/admin", middleware.Auth(), middleware.RequireRole(model.RoleAdmin))
	{
		admin.POST("/user/:id/ban", handler.AdminBan)
		admin.POST("/user/:id/unban", handler.AdminUnban)
	}

	// 角色任命只对超级管理员开放：管理员若能自己提权，权限层级就失效了
	superAdmin := v1.Group("/admin", middleware.Auth(), middleware.RequireRole(model.RoleSuperAdmin))
	{
		superAdmin.POST("/user/:id/role", handler.AdminSetRole)
	}

	return r
}
