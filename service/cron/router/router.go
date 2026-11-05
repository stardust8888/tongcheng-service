package router

import (
	"github.com/gin-gonic/gin"

	"tongcheng/library/ginx"
	"tongcheng/service/cron/controller/cron"
	"tongcheng/utils/constant"
)

func RegisterRoutes(engine *gin.Engine) {
	router := engine.Group(constant.BaseCronUrlV1)
	router.GET("/health", ginx.Check)

	cronRouter := router.Group("cron")
	cronRouter.GET("/task", cron.HandleCronTask)
}
