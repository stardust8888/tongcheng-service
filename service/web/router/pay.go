package router

import (
	"github.com/gin-gonic/gin"

	"tongcheng/service/web/controller/pay"
)

func registerPayRouter(router *gin.RouterGroup) {
	payRouter := router.Group("/pay")
	wxPay := payRouter.Group("/wx")
	wxPay.POST("/prepay", pay.WxPrepay)
	wxPay.POST("/callback", pay.WxPayCallBack)
}
