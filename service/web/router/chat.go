package router

import (
	"github.com/gin-gonic/gin"

	"tongcheng/service/web/controller/chat"
)

func registerChatRouter(router *gin.RouterGroup) {
	chatRouter := router.Group("/chat")
	chatRouter.GET("/list", chat.GetList)
	chatRouter.GET("/fresh", chat.GetUnreadAndLatestMessage)
	chatRouter.POST("/create", chat.Create)
	chatRouter.POST("/delete", chat.Delete)
}
