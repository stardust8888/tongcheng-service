package router

import (
	"github.com/gin-gonic/gin"

	"tongcheng/service/admin/controller/tool"
)

func registerDecryptRouter(router *gin.RouterGroup) {
	decryptRouter := router.Group("/decrypt")
	decryptRouter.GET("/plaintext", tool.Decrypt)
	fileRouter := router.Group("/file")
	fileRouter.GET("/view", tool.View)
}
