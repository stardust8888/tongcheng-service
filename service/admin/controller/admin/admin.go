package admin

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"tongcheng/conf"
	"tongcheng/db/mysql"
	"tongcheng/library/ginx"
	"tongcheng/library/jwt"
	"tongcheng/logs"
	"tongcheng/model"
	"tongcheng/utils"
	"tongcheng/utils/constant"
	"tongcheng/utils/x/timex"
)

type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResp struct {
	Token       string       `json:"token"`
	ExpiresAt   timex.Time   `json:"expiresAt"`
	Permissions []string     `json:"permissions"`
	User        *model.Admin `json:"user"`
}

// Login 管理员登录
func Login(c *gin.Context) {
	ctx := c.Request.Context()
	req := LoginReq{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		ginx.ResponseParamError(c)
		return
	}

	var admin model.Admin
	err = mysql.FindOne(ctx, mysql.GetAdminDB(), &admin, map[string]any{"username": req.Username})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	if admin.Status != 1 {
		ginx.ResponseError(c, constant.CodeUserDisabled)
		return
	}

	secretConfig := conf.GetConfiguration().Secret

	fingerprint := utils.MD5(req.Password + secretConfig.PwdSecret)
	if fingerprint != admin.Password {
		ginx.ResponseError(c, constant.CodePasswordError)
		return
	}

	jt, err := jwt.New(secretConfig.JwtSecret, timex.Day)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	simple := constant.SimpleUser{
		ID:      admin.ID,
		IsAdmin: admin.IsSuper,
	}
	token, err := jt.GenerateToken(simple)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	err = mysql.Save(ctx, mysql.GetAdminDB(), &admin)
	if err != nil {
		logs.Error(c, "更新管理员登录信息失败", zap.Error(err))
		//ginx.ResponseError(c, constant.CodeServerInternal)
		//return
	}
	admin.Password = constant.MaskedString

	loginLog := model.AdminLoginLog{
		AdminID:   admin.ID,
		LoginIP:   c.ClientIP(),
		LoginTime: timex.GetBeijingTime(),
	}
	err = mysql.Create(ctx, mysql.GetAdminDB(), &loginLog)
	if err != nil {
		logs.Error(c, "更新管理员登录日志失败", zap.Error(err))
	}

	response := LoginResp{
		Token:     token,
		ExpiresAt: timex.GetBeijingTime().Add(timex.Day),
		User:      &admin,
	}
	ginx.ResponseSuccess(c, response)
}
