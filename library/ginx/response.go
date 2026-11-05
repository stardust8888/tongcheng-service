package ginx

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"tongcheng/logs"
	"tongcheng/utils/common"
	"tongcheng/utils/constant"
)

// Response 标准响应结构体
type Response struct {
	RequestID  string `json:"requestID"`
	Code       int    `json:"code"`    // 错误码
	Msg        string `json:"message"` // 错误信息
	Data       any    `json:"data"`    // 返回数据
	ServerTime int64  `json:"serverTime"`
}

func ResponseSuccess(c *gin.Context, data any) {
	if data == nil {
		data = struct{}{}
	}
	c.JSON(http.StatusOK, Response{
		RequestID:  logs.Trace(c),
		Code:       constant.CodeSuccess,
		Msg:        constant.GetCodeText(constant.CodeSuccess),
		Data:       data,
		ServerTime: time.Now().Unix(),
	})
}

func ResponseNotify(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Response{
		RequestID:  logs.Trace(c),
		Code:       constant.CodeNotify,
		Msg:        msg,
		Data:       struct{}{},
		ServerTime: time.Now().Unix(),
	})
}

func ResponsePageSuccess[T any](c *gin.Context, items []T, page, pageSize int, total int) {
	totalPage := (total + pageSize - 1) / pageSize // 向上取整计算总页数
	c.JSON(http.StatusOK, Response{
		RequestID: logs.Trace(c),
		Code:      constant.CodeSuccess,
		Msg:       constant.GetCodeText(constant.CodeSuccess),
		Data: common.PageResult[T]{
			Items:     items,
			Page:      page,
			PageSize:  pageSize,
			Total:     total,
			TotalPage: totalPage,
		},
		ServerTime: time.Now().Unix(),
	})
}

func ResponseError(c *gin.Context, code int) {
	c.JSON(http.StatusOK, Response{
		RequestID:  logs.Trace(c),
		Code:       code,
		Msg:        constant.GetCodeText(code),
		Data:       struct{}{},
		ServerTime: time.Now().Unix(),
	})
}

func ResponseData(c *gin.Context, code int, data any) {
	c.JSON(http.StatusOK, Response{
		RequestID:  logs.Trace(c),
		Code:       code,
		Msg:        constant.GetCodeText(code),
		Data:       data,
		ServerTime: time.Now().Unix(),
	})
}

func ResponseNotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, Response{
		RequestID:  logs.Trace(c),
		Code:       constant.CodeNotFound,
		Msg:        constant.GetCodeText(constant.CodeNotFound),
		Data:       struct{}{},
		ServerTime: time.Now().Unix(),
	})
}

func ResponseParamError(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		RequestID:  logs.Trace(c),
		Code:       constant.CodeParamError,
		Msg:        constant.GetCodeText(constant.CodeParamError),
		Data:       struct{}{},
		ServerTime: time.Now().Unix(),
	})
}

func ResponseServerError(c *gin.Context, err error) {
	logs.Error(c, "服务器内部错误", zap.Error(err))
	c.JSON(http.StatusOK, Response{
		RequestID:  logs.Trace(c),
		Code:       constant.CodeServerInternal,
		Msg:        constant.GetCodeText(constant.CodeServerInternal),
		Data:       struct{}{},
		ServerTime: time.Now().Unix(),
	})
}
