package router

import (
	"github.com/gin-gonic/gin"

	"tongcheng/service/web/controller/message"
)

func registerMessageRouter(router *gin.RouterGroup) {
	messageRouter := router.Group("/message")
	messageRouter.GET("/list", message.GetList)
	messageRouter.GET("/search", message.Search)
	messageRouter.POST("/delete", message.Delete)
	messageRouter.POST("/retract", message.Retract)
}
