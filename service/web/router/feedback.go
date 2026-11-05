package router

import (
	"github.com/gin-gonic/gin"

	"tongcheng/service/web/controller/feedback"
)

func registerFeedbackRouter(router *gin.RouterGroup) {
	feedbackRouter := router.Group("/feedback")
	feedbackRouter.GET("/list", feedback.GetList)
	feedbackRouter.POST("/create", feedback.Create)
}
