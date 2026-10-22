package handler

import (
	"github.com/gin-gonic/gin"

	"hunjie/internal/dto"
	"hunjie/internal/middleware"
	"hunjie/internal/pkg/response"
	"hunjie/internal/service"
)

// ThreeViewQuestions 三观题库。登录即可取，答题免费。
// GET /api/v1/common/three-view/questions
func ThreeViewQuestions(c *gin.Context) {
	response.OK(c, service.MyThreeViewQuestions())
}

// MyThreeView 我的答题状态与答案
// GET /api/v1/user/three-view
func MyThreeView(c *gin.Context) {
	resp, err := service.MyThreeView(middleware.CurrentUser(c))
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, resp)
}

// SaveThreeView 提交我的三观答案（全量覆盖）。答题免费。
// PUT /api/v1/user/three-view
func SaveThreeView(c *gin.Context) {
	var req dto.SaveThreeViewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "请提交完整的答案")
		return
	}
	if err := service.SaveThreeView(middleware.CurrentUser(c), req); err != nil {
		fail(c, err)
		return
	}
	response.OK(c, nil)
}
