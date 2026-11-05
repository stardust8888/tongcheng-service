package feedback

import (
	"time"

	"github.com/gin-gonic/gin"

	"tongcheng/db/mysql"
	"tongcheng/db/redis"
	"tongcheng/library/ginx"
	"tongcheng/model"
	"tongcheng/utils/common"
	"tongcheng/utils/constant"
	"tongcheng/utils/x/timex"
)

// 每日反馈次数上限
const maxDailyFeedback = 3

type CreateReq struct {
	Content string `json:"content" binding:"required"`
}

// Create 提交反馈，每用户每天最多3次（redis 计数，key 带日期每日自然重置）
func Create(c *gin.Context) {
	ctx := c.Request.Context()
	userID := ginx.GetUserID(c)
	req := CreateReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		ginx.ResponseParamError(c)
		return
	}
	if len(req.Content) == 0 {
		ginx.ResponseParamError(c)
		return
	}

	// 每日次数限制：redis INCR 原子计数，key 内含日期，次日为新 key 自然重置
	dateStr := time.Now().Format(timex.DateFormatCompact)
	key := constant.RedisKey(constant.RedisFeedbackDaily, userID, dateStr)
	count, err := redis.Incr(ctx, key)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	if count == 1 {
		// 首次计数设置过期作清理；key 内含日期保证次日重置
		_ = redis.Expire(ctx, key, 25*time.Hour)
	}
	if count > maxDailyFeedback {
		ginx.ResponseNotify(c, "今日反馈次数已达上限")
		return
	}

	feedback := model.UserFeedback{
		UserID:  userID,
		Content: req.Content,
	}
	err = mysql.Create(ctx, mysql.GetUserDB(), &feedback)
	if err != nil {
		// 创建失败回滚计数
		_, _ = redis.Decr(ctx, key)
		ginx.ResponseServerError(c, err)
		return
	}
	ginx.ResponseSuccess(c, feedback)
}

type GetListReq struct {
	common.PageReq
}

// GetList 获取当前用户的反馈列表（分页，按id倒序）
func GetList(c *gin.Context) {
	ctx := c.Request.Context()
	userID := ginx.GetUserID(c)
	req := GetListReq{}
	if err := c.ShouldBindQuery(&req); err != nil {
		ginx.ResponseParamError(c)
		return
	}
	req.BindDefault()

	feedbacks := make([]*model.UserFeedback, 0)
	where := map[string]any{
		"user_id":   userID,
		mysql.Order: "id desc",
	}
	mysql.SetPage(where, req.Page, req.PageSize)
	total, err := mysql.PageQuery(ctx, mysql.GetUserDB(), &feedbacks, where)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	ginx.ResponsePageSuccess(c, feedbacks, req.Page, req.PageSize, total)
}
