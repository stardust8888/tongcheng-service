package pay

import (
	"github.com/gin-gonic/gin"

	"tongcheng/db/redis"
	"tongcheng/library/ginx"
	"tongcheng/library/httpclient/wxapi"
	"tongcheng/utils/constant"
	"tongcheng/utils/x/timex"
)

type WxPrePayRequest struct {
	Money   float64 `json:"money"`
	Product string  `json:"product"`
}

func WxPrepay(c *gin.Context) {
	ctx := c.Request.Context()
	req := WxPrePayRequest{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		ginx.ResponseParamError(c)
		return
	}

	userID := ginx.GetUserID(c)
	lockKey := constant.RedisKey(constant.RedisWxPrepayLock, userID)
	err = redis.SetNX(ctx, lockKey, req.Money, 10*timex.Second)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	defer func() {
		_ = redis.Del(ctx, lockKey)
	}()

	wxOpenID := ginx.GetWxOpenID(c)
	if wxOpenID == "" {
		ginx.ResponseParamError(c)
		return
	}

	prepayReq := wxapi.PrepayRequest{}
	prepayReq.Payer.Openid = wxOpenID
	prepayReq.Amount.Total = req.Money
	prepayReq.Description = "购买" + req.Product
	wxResp, err := wxapi.GetClient().PrePay(ctx, &prepayReq)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	respMap := map[string]any{
		"prepayID": wxResp.PrepayID,
	}
	ginx.ResponseSuccess(c, respMap)
}

func WxPayCallBack(c *gin.Context) {

}
