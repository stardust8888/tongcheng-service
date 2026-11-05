package router

import (
	"github.com/gin-gonic/gin"

	"tongcheng/service/web/controller/moment"
)

func registerMomentRouter(router *gin.RouterGroup) {
	momentRouter := router.Group("/moment")
	momentRouter.GET("/myList", moment.GetListWithComment)
	momentRouter.GET("/list/withComment", moment.GetListWithComment)
	momentRouter.GET("/detail/withComment", moment.GetDetailWithComment)
	momentRouter.POST("/create", moment.Create)
	momentRouter.POST("/delete", moment.Delete)
	momentRouter.POST("/like", moment.Like)
	momentRouter.POST("/unlike", moment.UnLike)
	momentRouter.GET("/comment/list", moment.GetCommentList)
	momentRouter.POST("/comment/create", moment.CreateComment)
	momentRouter.POST("/comment/delete", moment.DeleteComment)
}
