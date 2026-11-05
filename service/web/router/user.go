package router

import (
	"github.com/gin-gonic/gin"

	"tongcheng/service/web/controller/user"
)

func registerUserRouter(router *gin.RouterGroup) {
	userRouter := router.Group("/user")
	userRouter.GET("/profile", user.GetProfile)
	userRouter.GET("/home", user.GetHome)
	userRouter.POST("/profile/update", user.UpdateProfile)
	userRouter.GET("/child", user.GetChild)
	userRouter.POST("/child/create", user.CreateChild)
	userRouter.POST("/child/update", user.UpdateChild)
	userRouter.GET("/watch/List", user.GetWatchList)
	userRouter.POST("/watch", user.Watch)
	userRouter.POST("/watch/cancel", user.WatchCancel)
	userRouter.GET("/interaction/stat", user.GetInteractionStat)
	userRouter.GET("/view/me", user.ViewMe)
	userRouter.GET("/watch/me", user.WatchMe)
	userRouter.GET("/like/me", user.LikeMe)
	userRouter.GET("/comment/me", user.CommentMe)
	userRouter.GET("/black/list", user.GetBlackList)
	userRouter.POST("/black/add", user.AddBlack)
	userRouter.POST("/black/remove", user.RemoveBlack)
	userRouter.GET("/delete/apply", user.GetDeleteApply)
	userRouter.POST("/delete/apply", user.ApplyDelete)
	userRouter.POST("/delete/cancel", user.CancelDelete)
}
