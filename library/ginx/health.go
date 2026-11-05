package ginx

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"tongcheng/logs"
)

func Check(c *gin.Context) {
	logs.Info(c, "check")
	fmt.Println("gin", logs.Trace(c))
	fmt.Println("ctx", logs.Trace(c.Request.Context()))
	testSecondPosition(c)
	ResponseSuccess(c, nil)
}

func testSecondPosition(c *gin.Context) {
	logs.Info(c, "testSecondPosition")
	testFirstPosition(c)
}

func testFirstPosition(c *gin.Context) {
	logs.Info(c.Request.Context(), "testFirstPosition")
}
