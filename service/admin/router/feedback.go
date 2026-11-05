package router

import (
	"github.com/gin-gonic/gin"

	"tongcheng/service/admin/controller/feedback"
)

func registerFeedbackRouter(router *gin.RouterGroup) {
	feedbackRouter := router.Group("/feedback")
	feedbackRouter.GET("/list", feedback.GetList)
	feedbackRouter.POST("/reply", feedback.Reply)
}
