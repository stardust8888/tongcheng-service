package message

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"tongcheng/db/mysql"
	"tongcheng/db/redis"
	"tongcheng/library/ginx"
	"tongcheng/logs"
	"tongcheng/model"
	"tongcheng/repo"
	"tongcheng/utils/common"
	"tongcheng/utils/constant"
	"tongcheng/utils/tool/jsoniter"
)

type GetListReq struct {
	common.PageReq
	ChatID string `form:"chatID" json:"chatID" binding:"required"`
}

// 不需要使用页码和总数，直接滚动拉数据即可
func GetList(c *gin.Context) {
	ctx := c.Request.Context()
	req := GetListReq{}
	err := c.ShouldBindQuery(&req)
	if err != nil {
		ginx.ResponseParamError(c)
		return
	}
	req.BindDefault()
	if req.ChatID == "" {
		ginx.ResponseParamError(c)
		return
	}

	fromMysql := false
	MessageQueueKey := constant.RedisKey(constant.RedisKey(constant.RedisChatMessageQueue, req.ChatID))
	count, err := redis.ZCard(ctx, MessageQueueKey)
	if err != nil {
		fromMysql = true
	}

	total := int(count)
	if total < req.Page*req.PageSize {
		fromMysql = true
	}

	messages := make([]*model.Message, 0)
	if !fromMysql {
		ss := make([]string, 0)
		ss, err = redis.ZRevRange(ctx, MessageQueueKey, int64((req.Page-1)*req.PageSize), int64(req.Page*req.PageSize))
		if err != nil {
			logs.Error(c, "redis.ZRevRange", zap.Error(err))
			fromMysql = true
		} else {
			for i := range ss {
				message := &model.Message{}
				err = jsoniter.Unmarshal([]byte(ss[i]), message)
				if err != nil {
					logs.Error(c, "jsoniter.Unmarshal", zap.Error(err))
					continue
				}
				messages = append(messages, message)
			}
		}
	}

	if fromMysql {
		where := map[string]any{
			"chat_id":   req.ChatID,
			mysql.Order: "seq_id desc",
		}
		mysql.SetPage(where, req.Page, req.PageSize)
		total, err = mysql.PageQuery(ctx, mysql.GetUserDB(), &messages, where)
		if err != nil {
			ginx.ResponseServerError(c, err)
			return
		}

		go func() {
			ctxGo := context.WithoutCancel(ctx)
			_ = repo.RefreshMessageCache(ctxGo, req.ChatID)
		}()
	}

	ginx.ResponsePageSuccess(c, messages, req.Page, req.PageSize, total)
}

type SearchReq struct {
	common.PageReq
	ChatID  string `form:"chatID" json:"chatID"`
	Keyword string `form:"keyword" json:"keyword" binding:"required"`
}

func Search(c *gin.Context) {
	ctx := c.Request.Context()
	req := SearchReq{}
	err := c.ShouldBindQuery(&req)
	if err != nil {
		ginx.ResponseParamError(c)
		return
	}
	req.BindDefault()
	if req.Keyword != "" {
		ginx.ResponseParamError(c)
		return
	}

	where := map[string]any{}
	where["content like ?"] = mysql.LikeQuery(mysql.Like, req.Keyword)
	if req.ChatID != "" {
		where["chat_id"] = req.ChatID
	}

	messages := make([]*model.Message, 0)
	mysql.SetPage(where, req.Page, req.PageSize)
	total, err := mysql.PageQuery(ctx, mysql.GetUserDB(), &messages, where)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	ginx.ResponsePageSuccess(c, messages, req.Page, req.PageSize, total)
}

type DeleteReq struct {
	ChatID string `json:"chatID" binding:"required"`
	SeqID  []int  `json:"seqID" binding:"required"`
}

func Delete(c *gin.Context) {
	ctx := c.Request.Context()
	req := DeleteReq{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		ginx.ResponseParamError(c)
		return
	}

	where := map[string]any{
		"chat_id": req.ChatID,
		"seq_id":  req.SeqID,
	}
	_, err = mysql.Updates(ctx, mysql.GetUserDB(), &model.Message{}, where, map[string]any{"status": 1})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	ginx.ResponseSuccess(c, nil)
}

type RetractReq struct {
	ChatID string `json:"chatID" binding:"required"`
	SeqID  int    `json:"seqID" binding:"required"`
}

func Retract(c *gin.Context) {
	ctx := c.Request.Context()
	req := RetractReq{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		ginx.ResponseParamError(c)
		return
	}

	where := map[string]any{
		"chat_id": req.ChatID,
		"seq_id":  req.SeqID,
	}
	_, err = mysql.Updates(ctx, mysql.GetUserDB(), &model.Message{}, where, map[string]any{"status": 2})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	ginx.ResponseSuccess(c, nil)
}
