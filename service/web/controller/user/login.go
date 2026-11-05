package user

import (
	"github.com/gin-gonic/gin"

	"tongcheng/conf"
	"tongcheng/db/mysql"
	"tongcheng/library/ginx"
	"tongcheng/library/httpclient/wxapi"
	"tongcheng/library/jwt"
	"tongcheng/model"
	"tongcheng/repo"
	"tongcheng/utils"
	"tongcheng/utils/constant"
	"tongcheng/utils/x/timex"
)

type GetUserPhoneReq struct {
	OpenID string `form:"openID" json:"openID" binding:"required"`
	Code   string `form:"code" json:"code" binding:"required"`
	Phone  string `form:"phone" json:"phone"`
}

type GetUserPhoneResp struct {
	OpenID string `json:"openID"`
	Phone  string `json:"phone"`
}

func GetUserPhone(c *gin.Context) {
	ctx := c.Request.Context()
	req := GetUserPhoneReq{}
	err := c.ShouldBindQuery(&req)
	if err != nil {
		ginx.ResponseParamError(c)
		return
	}

	user := model.User{}
	condition := map[string]any{"wx_open_id": req.OpenID}
	exist, err := mysql.Exist(ctx, mysql.GetUserDB(), &user, condition)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	if !exist {
		ginx.ResponseNotify(c, "用户不存在，请退出重试")
		return
	}

	phone := ""
	wxResp, err := wxapi.GetClient().GetUserPhoneNumber(ctx, req.Code)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	phone = wxResp.PhoneInfo.PhoneNumber
	if len(phone) < 4 {
		ginx.ResponseNotify(c, "不支持的手机号")
		return
	}

	_, err = mysql.UpdateByID(ctx, mysql.GetUserDB(), &user, user.ID, map[string]any{
		"phone":        repo.EncryptUserInfo(ctx, user.ID, phone),
		"phone_suffix": phone[len(phone)-4:],
	})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	resp := GetUserPhoneResp{
		OpenID: user.WxOpenID,
		Phone:  phone,
	}
	ginx.ResponseSuccess(c, resp)
}

type FastLoginReq struct {
	Code string `form:"code" json:"code" binding:"required"`
}

type LoginResp struct {
	Token     string           `json:"token"`
	ExpiresAt timex.Time       `json:"expiresAt"`
	User      *model.User      `json:"user"`
	Child     *model.UserChild `json:"child"`
}

func FastLogin(c *gin.Context) {
	ctx := c.Request.Context()
	req := FastLoginReq{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		ginx.ResponseParamError(c)
		return
	}

	wxResp, err := wxapi.GetClient().GetOpenID(ctx, req.Code)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	user := model.User{}
	exist, err := mysql.Exist(ctx, mysql.GetUserDB(), &user, map[string]any{"wx_open_id": wxResp.OpenID})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	resp := &LoginResp{}
	if !exist {
		user = model.User{
			Username:   constant.PrefixUsername + utils.RandNumberString(6),
			Avatar:     "avatar_0" + utils.RandNumberString(1) + ".png",
			WxOpenID:   wxResp.OpenID,
			WxUnionID:  wxResp.UnionID,
			SessionKey: wxResp.SessionKey,
		}
		err = mysql.Create(ctx, mysql.GetUserDB(), &user)
		if err != nil {
			ginx.ResponseServerError(c, err)
			return
		}
		resp.User = &user
		ginx.ResponseData(c, constant.CodeUserNotExistGoRegister, resp)
		return
	}
	if user.Phone == "" {
		resp.User = &user
		ginx.ResponseData(c, constant.CodeUserNotExistGoRegister, resp)
		return
	}

	if user.Status != 0 {
		ginx.ResponseError(c, constant.CodeUserDisabled)
		return
	}

	// after login 登录查出用户所有会话id，把该用户写入缓存
	_ = repo.AfterRegister(ctx, user.ID)

	secretConfig := conf.GetConfiguration().Secret
	jt, err := jwt.New(secretConfig.JwtSecret, timex.Month)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	simple := constant.SimpleUser{
		ID:       user.ID,
		WxOpenID: wxResp.OpenID,
	}
	token, err := jt.GenerateToken(simple)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	userChild := model.UserChild{}
	_, err = mysql.Exist(ctx, mysql.GetUserDB(), &userChild, map[string]any{"user_id": user.ID})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	resp = &LoginResp{
		Token:     token,
		ExpiresAt: timex.GetBeijingTime().Add(timex.Month),
		User:      &user,
		Child:     &userChild,
	}

	ginx.ResponseSuccess(c, resp)
}

func LoginOut(c *gin.Context) {
	ginx.ResponseSuccess(c, nil)
}
