package tool

import (
	"github.com/gin-gonic/gin"

	"tongcheng/library/ginx"
	"tongcheng/repo"
)

// DecryptReq 解密请求
type DecryptReq struct {
	UserID     int    `form:"userID" binding:"required"`
	CipherText string `form:"ciperText" binding:"required"`
}

// DecryptResp 解密结果（明文）
type DecryptResp struct {
	Plaintext string `json:"plaintext"`
}

// Decrypt 解密某用户的手机号和微信，明文返回（列表仍返回密文，按需调用）
func Decrypt(c *gin.Context) {
	ctx := c.Request.Context()
	req := DecryptReq{}
	if err := c.ShouldBindQuery(&req); err != nil {
		ginx.ResponseParamError(c)
		return
	}

	resp := DecryptResp{
		Plaintext: repo.DecryptUserInfo(ctx, req.UserID, req.CipherText),
	}
	ginx.ResponseSuccess(c, resp)
}
