package feedback

import (
	"time"

	"github.com/gin-gonic/gin"

	"tongcheng/db/mysql"
	"tongcheng/library/ginx"
	"tongcheng/model"
	"tongcheng/repo"
	"tongcheng/utils/common"
)

// GetListReq 反馈列表请求，支持按用户ID、回复状态筛选
type GetListReq struct {
	common.PageReq
	UserID      *int `form:"userID"`
	ReplyStatus *int `form:"replyStatus"`
}

type GetListResp struct {
	*model.UserFeedback
	User *model.User `json:"user"`
}

// GetList 获取所有用户反馈（附带用户信息，多条件筛选 + 分页）
func GetList(c *gin.Context) {
	ctx := c.Request.Context()
	req := GetListReq{}
	if err := c.ShouldBindQuery(&req); err != nil {
		ginx.ResponseParamError(c)
		return
	}
	req.BindDefault()

	feedbacks := make([]*model.UserFeedback, 0)
	where := map[string]any{
		mysql.Order: "id desc",
	}
	if req.UserID != nil {
		where["user_id"] = *req.UserID
	}
	if req.ReplyStatus != nil {
		where["reply_status"] = *req.ReplyStatus
	}
	mysql.SetPage(where, req.Page, req.PageSize)

	total, err := mysql.PageQuery(ctx, mysql.GetUserDB(), &feedbacks, where)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	resp := make([]*GetListResp, 0, len(feedbacks))
	if len(feedbacks) == 0 {
		ginx.ResponsePageSuccess(c, resp, req.Page, req.PageSize, total)
		return
	}

	// 批量查用户信息，一次查询避免 N+1
	usersID := make([]int, 0, len(feedbacks))
	for _, f := range feedbacks {
		usersID = append(usersID, f.UserID)
	}
	usersMap, err := repo.GetUsersByID(ctx, usersID)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	for _, f := range feedbacks {
		item := &GetListResp{UserFeedback: f}
		if user, ok := usersMap[f.UserID]; ok && user != nil {
			item.User = user
		}
		resp = append(resp, item)
	}

	ginx.ResponsePageSuccess(c, resp, req.Page, req.PageSize, total)
}

// ReplyReq 回复反馈请求
type ReplyReq struct {
	FeedbackID int    `json:"feedbackID" binding:"required"`
	Reply      string `json:"reply" binding:"required"`
}

// Reply 回复用户反馈，写入回复内容/时间，置回复状态为已回复
func Reply(c *gin.Context) {
	ctx := c.Request.Context()
	req := ReplyReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		ginx.ResponseParamError(c)
		return
	}

	feedback := model.UserFeedback{}
	exist, err := mysql.Exist(ctx, mysql.GetUserDB(), &feedback, map[string]any{"id": req.FeedbackID})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	if !exist {
		ginx.ResponseNotify(c, "反馈不存在")
		return
	}

	now := time.Now()
	_, err = mysql.UpdateByID(ctx, mysql.GetUserDB(), &model.UserFeedback{}, req.FeedbackID, map[string]any{
		"reply":        req.Reply,
		"reply_at":     now,
		"reply_status": 1,
	})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	ginx.ResponseSuccess(c, nil)
}
