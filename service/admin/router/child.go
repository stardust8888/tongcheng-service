package router

import (
	"github.com/gin-gonic/gin"

	"tongcheng/service/admin/controller/child"
)

func registerChildRouter(router *gin.RouterGroup) {
	childRouter := router.Group("/child")
	childRouter.GET("/list", child.GetList)
	childRouter.GET("/elder", child.GetElder)
	childRouter.POST("/profile/approve", child.Approve)
	childRouter.POST("/profile/reject", child.Reject)
	childRouter.POST("/profile/show", child.Show)
	childRouter.POST("/profile/hide", child.Hidden)
}
