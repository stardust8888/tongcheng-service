package ginx

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"tongcheng/conf"
	"tongcheng/library/jwt"
	"tongcheng/library/snowflake"
	"tongcheng/logs"
	"tongcheng/utils/constant"
	"tongcheng/utils/x/timex"
)

var (
	Context = &gin.Context{}
)

func GetUserID(c *gin.Context) int {
	return c.GetInt(constant.UserID)
}

func SetUserID(c *gin.Context, userID int) {
	c.Set(constant.UserID, userID)
}

func GetWxOpenID(c *gin.Context) string {
	return c.GetString(constant.WxOpenID)
}

func SetWxOpenID(c *gin.Context, wxOpenID string) {
	c.Set(constant.WxOpenID, wxOpenID)
}

func GetAdminID(c *gin.Context) int {
	return c.GetInt(constant.AdminID)
}

func SetAdminID(c *gin.Context, adminID int) {
	c.Set(constant.AdminID, adminID)
}

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		snowID := snowflake.GenerateString()
		c.Set(logs.RequestIDKey, snowID)
		ctx := context.WithValue(c.Request.Context(), logs.RequestIDKey, snowID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// CORSMiddleware 跨域中间件
func CORSMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	})
}

// LoggerMiddleware 日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 请求处理完成后记录日志
		end := time.Now()
		latency := end.Sub(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		comment := c.Errors.ByType(gin.ErrorTypePrivate).String()

		if raw != "" {
			path = path + "?" + raw
		}

		logs.Info(c, "[GIN]",
			zap.Int("status", statusCode),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("ip", clientIP),
			zap.Duration("latency", latency),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.String("errors", comment),
		)
	}
}

// AuthMiddleware 简单认证中间件示例
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader(constant.HeaderAuthorization)
		if token == "" {
			token = c.Query(constant.HeaderAuthTokenInURL)
		}
		if token == "" {
			ResponseError(c, constant.CodeUnauthorized)
			c.Abort()
			return
		}

		jt, err := jwt.New(conf.GetConfiguration().Secret.JwtSecret, timex.Month)
		if err != nil {
			ResponseError(c, constant.CodeUnauthorized)
			c.Abort()
			return
		}

		claims, err := jt.ParseToken(token)
		if err != nil {
			ResponseError(c, constant.CodeUnauthorized)
			c.Abort()
			return
		}
		if claims.UserID == 0 || claims.IsAdmin > 0 {
			ResponseError(c, constant.CodeUnauthorized)
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		SetUserID(c, claims.UserID)
		SetWxOpenID(c, claims.WxOpenID)

		c.Next()
	}
}

// AdminAuthMiddleware 管理员认证中间件，校验 token 有效且 claims.IsAdmin 必须为 true
// 用于 admin 后台路由，隔离 C 端用户 token（C 端 token 的 IsAdmin 为 false 会被拒绝）
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader(constant.HeaderAuthorization)
		if token == "" {
			token = c.Query(constant.HeaderAuthTokenInURL)
		}
		if token == "" {
			ResponseError(c, constant.CodeUnauthorized)
			c.Abort()
			return
		}

		jt, err := jwt.New(conf.GetConfiguration().Secret.JwtSecret, timex.Month)
		if err != nil {
			ResponseError(c, constant.CodeUnauthorized)
			c.Abort()
			return
		}

		claims, err := jt.ParseToken(token)
		if err != nil {
			ResponseError(c, constant.CodeUnauthorized)
			c.Abort()
			return
		}
		if claims.UserID == 0 || claims.IsAdmin == 0 {
			ResponseError(c, constant.CodePermissionDenied)
			c.Abort()
			return
		}

		SetAdminID(c, claims.UserID)
		SetWxOpenID(c, claims.WxOpenID)

		c.Next()
	}
}

// TimeoutMiddleware 超时中间件
func TimeoutMiddleware(timeout int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c, time.Duration(timeout)*time.Second)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		finished := make(chan struct{})

		go func() {
			c.Next()
			finished <- struct{}{}
		}()

		select {
		case <-finished:
		case <-ctx.Done():
			c.JSON(http.StatusRequestTimeout, gin.H{
				"code": constant.CodeTimeout,
				"msg":  constant.GetCodeText(constant.CodeTimeout),
				"data": nil,
			})
			c.Abort()
		}
	}
}
