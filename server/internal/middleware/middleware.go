package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"hunjie/internal/db"
	"hunjie/internal/model"
	"hunjie/internal/pkg/jwt"
	"hunjie/internal/pkg/logger"
	"hunjie/internal/pkg/response"
)

const ctxUserKey = "ctx_user"

// CurrentUser 取出当前登录用户。Auth 中间件之后的 handler 里必定非 nil
func CurrentUser(c *gin.Context) *model.User {
	if v, ok := c.Get(ctxUserKey); ok {
		if u, ok := v.(*model.User); ok {
			return u
		}
	}
	return nil
}

// CurrentUserID 取当前用户 id，未登录返回 0
func CurrentUserID(c *gin.Context) uint64 {
	if u := CurrentUser(c); u != nil {
		return u.ID
	}
	return 0
}

// Auth 鉴权中间件。
//
// 这里直接查库拿完整用户记录并放入 context，后续的 RequireRole / RequireActive
// 都基于这份实时数据判定，而不是 token 里的声明。
// token 有效期 7 天，若只信 token 里的 role/status，用户被降级或封禁后会留出
// 最长 7 天的越权窗口。一次主键查询的成本远低于这个风险。
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr == "" {
			response.Unauthorized(c, "请先登录")
			return
		}

		claims, err := jwt.Parse(tokenStr)
		if err != nil {
			response.Unauthorized(c, "登录已过期，请重新登录")
			return
		}

		var u model.User
		if err := db.DB.First(&u, claims.UserID).Error; err != nil {
			response.Unauthorized(c, "账号不存在")
			return
		}

		// 封禁用户在此统一拦截，无论 role 多高、访问什么接口
		if u.Status == model.StatusBanned {
			msg := "账号已被封禁"
			if u.BanReason != "" {
				msg += "：" + u.BanReason
			}
			response.Abort(c, response.CodeBanned, msg)
			return
		}

		c.Set(ctxUserKey, &u)
		c.Next()
	}
}

// RequireBrowse 要求资料已提交（可浏览他人）。
// 资料未完善的用户不能浏览他人 —— 防止只看不填拉低资料池质量
func RequireBrowse() gin.HandlerFunc {
	return func(c *gin.Context) {
		u := CurrentUser(c)
		if u.Status == model.StatusIncomplete {
			response.Abort(c, response.CodeProfileIncomplete, "请先完善个人资料")
			return
		}
		c.Next()
	}
}

// RequireActive 要求资料审核通过（可发起互动、聊天）
func RequireActive() gin.HandlerFunc {
	return func(c *gin.Context) {
		u := CurrentUser(c)
		switch u.Status {
		case model.StatusIncomplete:
			response.Abort(c, response.CodeProfileIncomplete, "请先完善个人资料")
		case model.StatusAuditing:
			response.Abort(c, response.CodeProfileAuditing, "资料审核中，通过后即可使用该功能")
		case model.StatusRejected:
			response.Abort(c, response.CodeProfileRejected, "资料审核未通过，请修改后重新提交")
		default:
			c.Next()
		}
	}
}

// RequireRole 要求最低角色。角色向上包含：role >= min 即通过。
//
// 这是管理功能的唯一授权判定点。前端只在 role>1 时渲染管理入口，
// 那只是体验优化 —— 用户完全可以直接构造请求打管理接口。
func RequireRole(min int8) gin.HandlerFunc {
	return func(c *gin.Context) {
		u := CurrentUser(c)
		if u.Role < min {
			response.Forbidden(c, "没有操作权限")
			return
		}
		c.Next()
	}
}

// Logger 请求日志
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		if raw := c.Request.URL.RawQuery; raw != "" {
			path += "?" + raw
		}

		c.Next()

		logger.Infof("%s %s %d %v", c.Request.Method, path,
			c.Writer.Status(), time.Since(start).Round(time.Millisecond))
	}
}

// CORS 跨域。小程序本身不受同源策略限制，这里主要是为了 H5 调试方便
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if after, ok := strings.CutPrefix(auth, "Bearer "); ok {
		return after
	}
	// WebSocket 握手无法自定义 header，从 query 取
	return c.Query("token")
}
