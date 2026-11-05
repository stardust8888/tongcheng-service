package chat

import (
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"tongcheng/conf"
	"tongcheng/library/ginx"
	"tongcheng/library/jwt"
	"tongcheng/logs"
	"tongcheng/model"
	"tongcheng/utils/constant"
	"tongcheng/utils/x/timex"
)

const (
	HandshakeTimeout = 30
	ReadBufferSize   = 1024
	WriteBufferSize  = 1024
)

var upgrader = websocket.Upgrader{
	HandshakeTimeout: time.Duration(HandshakeTimeout) * time.Second,
	ReadBufferSize:   ReadBufferSize,
	WriteBufferSize:  WriteBufferSize,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func Handshake(c *gin.Context) {
	token := c.Query(constant.HeaderAuthTokenInURL)
	if token == "" {
		ginx.ResponseError(c, constant.CodeUnauthorized)
		return
	}

	jt, err := jwt.New(conf.GetConfiguration().Secret.JwtSecret, timex.Month)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	claims, err := jt.ParseToken(token)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logs.Error(c, "升级为websocket失败", zap.Error(err))
		ginx.ResponseServerError(c, err)
		return
	}

	user := constant.SimpleUser{
		ID: claims.UserID,
	}

	client := &Client{
		conn: conn,
		user: &user,
		send: make(chan []byte, SendChannel),
		done: make(chan struct{}),
	}

	WriteTimeout(mediator.register, client, WriteChannelTime)

	err = conn.WriteMessage(websocket.TextMessage, []byte("connect success"))
	if err != nil {
		logs.Error(c, "发送消息失败", zap.Error(err))
	}

	go func() {
		_ = client.Write()
	}()
	go func() {
		_ = client.Read()
	}()
}

func GetConnectCount(c *gin.Context) {
	val := atomic.LoadInt64(&mediator.connect)
	resp := map[string]any{
		"count": val,
	}
	ginx.ResponseSuccess(c, resp)
}

type Notify struct {
	Kind    int            `json:"kind" binding:"required"`
	Message *model.Message `json:"message" binding:"required"`
}

func MessagePush(c *gin.Context) {
	req := Notify{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	req.Message.SetMsgID()

	WriteTimeout(mediator.broadcast, req.Message.Marshal(), WriteChannelTime)

	ginx.ResponseSuccess(c, nil)
}
