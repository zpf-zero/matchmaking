package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"hunjie/internal/dto"
	"hunjie/internal/middleware"
	"hunjie/internal/model"
	"hunjie/internal/pkg/response"
	"hunjie/internal/service"
)

// Recommend 推荐流
// GET /api/v1/discover/recommend
func Recommend(c *gin.Context) {
	var page dto.PageQuery
	_ = c.ShouldBindQuery(&page)

	list, total, err := service.Recommend(middleware.CurrentUser(c), page)
	if err != nil {
		fail(c, err)
		return
	}
	page.Normalize()
	response.Page(c, list, total, page.Page, page.PageSize)
}

// Search 条件筛选
// GET /api/v1/discover/search
func Search(c *gin.Context) {
	var q dto.SearchQuery
	_ = c.ShouldBindQuery(&q)

	list, total, err := service.Search(middleware.CurrentUser(c), q)
	if err != nil {
		fail(c, err)
		return
	}
	q.Normalize()
	response.Page(c, list, total, q.Page, q.PageSize)
}

// UserDetail 他人资料详情
// GET /api/v1/user/detail/:id
func UserDetail(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	resp, err := service.GetUserDetail(middleware.CurrentUser(c), id)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, resp)
}

// Like 喜欢
// POST /api/v1/interaction/like
func Like(c *gin.Context) {
	var req dto.TargetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数错误")
		return
	}
	res, err := service.Like(middleware.CurrentUser(c), req.ToUserID)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, res)
}

// Pass 不喜欢
// POST /api/v1/interaction/pass
func Pass(c *gin.Context) {
	var req dto.TargetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数错误")
		return
	}
	if err := service.Pass(middleware.CurrentUser(c), req.ToUserID); err != nil {
		fail(c, err)
		return
	}
	response.OK(c, nil)
}

// Unlike 取消喜欢
// DELETE /api/v1/interaction/like/:id
func Unlike(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	if err := service.Unlike(middleware.CurrentUser(c), id); err != nil {
		fail(c, err)
		return
	}
	response.OK(c, nil)
}

// LikedMe 谁喜欢我
// GET /api/v1/interaction/liked-me
func LikedMe(c *gin.Context) {
	listInteraction(c, service.ListLikedMe)
}

// MyLikes 我喜欢谁
// GET /api/v1/interaction/my-likes
func MyLikes(c *gin.Context) {
	listInteraction(c, service.ListMyLikes)
}

// ViewedMe 谁看过我
// GET /api/v1/interaction/viewed-me
func ViewedMe(c *gin.Context) {
	listInteraction(c, service.ListViewedMe)
}

func listInteraction(c *gin.Context,
	fn func(*model.User, dto.PageQuery) ([]dto.UserCard, int64, error)) {

	var page dto.PageQuery
	_ = c.ShouldBindQuery(&page)

	list, total, err := fn(middleware.CurrentUser(c), page)
	if err != nil {
		fail(c, err)
		return
	}
	page.Normalize()
	response.Page(c, list, total, page.Page, page.PageSize)
}

// UnreadCount 各类未读数（红点）
// GET /api/v1/interaction/unread-count
func UnreadCount(c *gin.Context) {
	response.OK(c, service.UnreadCount(middleware.CurrentUser(c)))
}

// MarkRead 标记已读
// PUT /api/v1/interaction/read?type=2
func MarkRead(c *gin.Context) {
	t, err := strconv.Atoi(c.Query("type"))
	if err != nil || (t != model.InteractionLike && t != model.InteractionView) {
		response.ParamError(c, "type 只能是 1(浏览) 或 2(喜欢)")
		return
	}
	if err := service.MarkRead(middleware.CurrentUser(c), t); err != nil {
		fail(c, err)
		return
	}
	response.OK(c, nil)
}
