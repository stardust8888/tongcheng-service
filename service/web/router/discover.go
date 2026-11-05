package router

import (
	"github.com/gin-gonic/gin"

	"tongcheng/service/web/controller/discover"
)

func registerDiscoverRouter(router *gin.RouterGroup) {
	discoverRouter := router.Group("/discover")
	discoverRouter.GET("/list", discover.GetList)
}
