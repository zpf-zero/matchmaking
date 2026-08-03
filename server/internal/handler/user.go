package handler

import (
	"github.com/gin-gonic/gin"

	"hunjie/internal/dto"
	"hunjie/internal/middleware"
	"hunjie/internal/pkg/response"
	"hunjie/internal/pkg/upload"
	"hunjie/internal/service"
)

// Me 我的完整信息
// GET /api/v1/user/me
func Me(c *gin.Context) {
	resp, err := service.GetMe(middleware.CurrentUser(c))
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, resp)
}

// SaveProfile 保存我的资料（全量覆盖）
// PUT /api/v1/user/profile
func SaveProfile(c *gin.Context) {
	var req dto.SaveProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "请把必填项填写完整")
		return
	}
	if err := service.SaveProfile(middleware.CurrentUser(c), req); err != nil {
		fail(c, err)
		return
	}
	response.OK(c, nil)
}

// SaveRequirement 保存择偶要求
// PUT /api/v1/user/requirement
func SaveRequirement(c *gin.Context) {
	var req dto.SaveRequirementReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数错误")
		return
	}
	if err := service.SaveRequirement(middleware.CurrentUser(c), req); err != nil {
		fail(c, err)
		return
	}
	response.OK(c, nil)
}

// SubmitAudit 提交资料审核
// POST /api/v1/user/submit-audit
func SubmitAudit(c *gin.Context) {
	if err := service.SubmitAudit(middleware.CurrentUser(c)); err != nil {
		fail(c, err)
		return
	}
	response.OK(c, nil)
}

// AddPhoto 新增相册图片
// POST /api/v1/user/photo
func AddPhoto(c *gin.Context) {
	var req dto.AddPhotoReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "缺少图片地址")
		return
	}
	p, err := service.AddPhoto(middleware.CurrentUser(c), req.URL)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, p)
}

// DeletePhoto 删除相册图片
// DELETE /api/v1/user/photo/:id
func DeletePhoto(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	if err := service.DeletePhoto(middleware.CurrentUser(c), id); err != nil {
		fail(c, err)
		return
	}
	response.OK(c, nil)
}

// SortPhotos 调整相册顺序
// PUT /api/v1/user/photo/sort
func SortPhotos(c *gin.Context) {
	var req dto.SortPhotoReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数错误")
		return
	}
	if err := service.SortPhotos(middleware.CurrentUser(c), req.IDs); err != nil {
		fail(c, err)
		return
	}
	response.OK(c, nil)
}

// UploadImage 图片上传
// POST /api/v1/common/upload/image?scene=avatar
func UploadImage(c *gin.Context) {
	scene := c.DefaultQuery("scene", "photo")

	fh, err := c.FormFile("file")
	if err != nil {
		response.ParamError(c, "请选择要上传的图片")
		return
	}

	url, err := upload.SaveImage(fh, scene)
	if err != nil {
		response.ParamError(c, err.Error())
		return
	}
	response.OK(c, gin.H{"url": url})
}
