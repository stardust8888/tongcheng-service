package router

import (
	"github.com/gin-gonic/gin"

	"tongcheng/service/admin/controller/user"
)

func registerUserRouter(router *gin.RouterGroup) {
	userRouter := router.Group("/user")
	userRouter.GET("/list", user.GetList)
	userRouter.GET("/child/detail", user.ChildDetail)
	userRouter.POST("/disable", user.Disable)
	userRouter.POST("/enable", user.Enable)
	userRouter.GET("/watch/list", user.GetWatchList)
}
