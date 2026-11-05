package moment

import (
	"github.com/gin-gonic/gin"

	"tongcheng/db/mysql"
	"tongcheng/library/ginx"
	"tongcheng/model"
	"tongcheng/repo"
	"tongcheng/utils/common"
)

// GetListReq 动态列表请求，支持多条件筛选（均为可选，不传则不过滤）
type GetListReq struct {
	common.PageReq
	UserID       *int `form:"userID"`       // 发布用户ID（可选）
	ReviewStatus *int `form:"reviewStatus"` // 审核状态 0:未审核1:通过2:拒绝（可选）
	Status       *int `form:"status"`       // 展示状态 0:未展示1:展示中（可选）
	IsDelete     *int `form:"isDelete"`     // 是否删除 0:正常1:已删除（可选，不传则不过滤，含已删除）
}

// GetListResp 动态列表项，内联动态字段，并附带发布者(家长)信息
// phone/wx 为密文（与用户列表/长辈资料一致，前端按需调解密接口）
type GetListResp struct {
	*model.UserMoment
	User        *model.User `json:"user"`         // 发布者(家长)
	Phone       string      `json:"phone"`        // 发布者手机号密文
	PhoneSuffix string      `json:"phone_suffix"` // 手机尾号
	Wx          string      `json:"wx"`           // 微信密文
}

// GetList 获取动态列表（user_moment 表），附带发布者(家长)信息，多条件筛选 + 分页
func GetList(c *gin.Context) {
	ctx := c.Request.Context()
	req := GetListReq{}
	if err := c.ShouldBindQuery(&req); err != nil {
		ginx.ResponseParamError(c)
		return
	}
	req.BindDefault()

	userMoments := make([]*model.UserMoment, 0)
	where := map[string]any{
		mysql.Order: "id desc",
	}
	if req.IsDelete != nil {
		where["is_delete"] = *req.IsDelete
	}
	if req.UserID != nil {
		where["user_id"] = *req.UserID
	}
	if req.ReviewStatus != nil {
		where["review_status"] = *req.ReviewStatus
	}
	if req.Status != nil {
		where["status"] = *req.Status
	}
	mysql.SetPage(where, req.Page, req.PageSize)

	total, err := mysql.PageQuery(ctx, mysql.GetUserDB(), &userMoments, where)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	resp := make([]*GetListResp, 0, len(userMoments))
	if len(userMoments) == 0 {
		ginx.ResponsePageSuccess(c, resp, req.Page, req.PageSize, total)
		return
	}

	// 批量查发布者信息(家长)，一次查询避免 N+1
	usersID := make([]int, 0, len(userMoments))
	for _, m := range userMoments {
		usersID = append(usersID, m.UserID)
	}
	usersMap, err := repo.GetUsersByID(ctx, usersID)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	for _, m := range userMoments {
		item := &GetListResp{UserMoment: m}
		if user, ok := usersMap[m.UserID]; ok && user != nil {
			item.User = user
			item.Phone = user.Phone
			item.PhoneSuffix = user.PhoneSuffix
			item.Wx = user.Wx
		}
		resp = append(resp, item)
	}

	ginx.ResponsePageSuccess(c, resp, req.Page, req.PageSize, total)
}

// ApproveReq 动态审核通过
type ApproveReq struct {
	MomentID int `json:"momentID" binding:"required"`
}

// Approve 审核通过，review_status 置 1 并自动展示(status 置 1)，清空审核原因
func Approve(c *gin.Context) {
	ctx := c.Request.Context()
	req := ApproveReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		ginx.ResponseParamError(c)
		return
	}

	_, err := mysql.UpdateByID(ctx, mysql.GetUserDB(), &model.UserMoment{}, req.MomentID, map[string]any{
		"review_status": 1,
		"review_reason": "",
		"status":        1,
	})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	ginx.ResponseSuccess(c, nil)
}

// RejectReq 动态审核拒绝
type RejectReq struct {
	MomentID int    `json:"momentID" binding:"required"`
	Reason   string `json:"reason" binding:"required"`
}

// Reject 审核拒绝，review_status 置 2，写入拒绝原因
func Reject(c *gin.Context) {
	ctx := c.Request.Context()
	req := RejectReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		ginx.ResponseParamError(c)
		return
	}

	_, err := mysql.UpdateByID(ctx, mysql.GetUserDB(), &model.UserMoment{}, req.MomentID, map[string]any{
		"review_status": 2,
		"review_reason": req.Reason,
	})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	ginx.ResponseSuccess(c, nil)
}

// ShowReq 动态展示
type ShowReq struct {
	MomentID int `json:"momentID" binding:"required"`
}

// Show 展示动态，status 置 1（出现在动态流）
func Show(c *gin.Context) {
	ctx := c.Request.Context()
	req := ShowReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		ginx.ResponseParamError(c)
		return
	}

	_, err := mysql.UpdateByID(ctx, mysql.GetUserDB(), &model.UserMoment{}, req.MomentID, map[string]any{
		"status": 1,
	})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	ginx.ResponseSuccess(c, nil)
}

// HiddenReq 动态隐藏
type HiddenReq struct {
	MomentID int `json:"momentID" binding:"required"`
}

// Hidden 隐藏动态，status 置 0（从动态流下线，不影响审核状态）
func Hidden(c *gin.Context) {
	ctx := c.Request.Context()
	req := HiddenReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		ginx.ResponseParamError(c)
		return
	}

	_, err := mysql.UpdateByID(ctx, mysql.GetUserDB(), &model.UserMoment{}, req.MomentID, map[string]any{
		"status": 0,
	})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	ginx.ResponseSuccess(c, nil)
}
