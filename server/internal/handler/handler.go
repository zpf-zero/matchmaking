package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"hunjie/internal/pkg/logger"
	"hunjie/internal/pkg/response"
	"hunjie/internal/service"
)

// fail 统一错误出口。
//
// 业务错误（service.BizError）的文案直接展示给用户；
// 其余一律视为系统错误，记日志后只回一句通用提示，不把数据库/网络细节暴露到前端。
func fail(c *gin.Context, err error) {
	var be *service.BizError
	if errors.As(err, &be) {
		response.Fail(c, response.CodeParamError, be.Msg)
		return
	}
	logger.Errorf("%s %s 内部错误: %v", c.Request.Method, c.Request.URL.Path, err)
	response.ServerError(c, "")
}

// pathID 取路径参数中的 id
func pathID(c *gin.Context, key string) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param(key), 10, 64)
	if err != nil || id == 0 {
		response.ParamError(c, "参数不正确")
		return 0, false
	}
	return id, true
}
