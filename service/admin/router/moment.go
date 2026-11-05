package router

import (
	"github.com/gin-gonic/gin"

	"tongcheng/service/admin/controller/moment"
)

func registerMomentRouter(router *gin.RouterGroup) {
	momentRouter := router.Group("/moment")
	momentRouter.GET("/list", moment.GetList)
	momentRouter.POST("/profile/approve", moment.Approve)
	momentRouter.POST("/profile/reject", moment.Reject)
	momentRouter.POST("/profile/show", moment.Show)
	momentRouter.POST("/profile/hide", moment.Hidden)
}
