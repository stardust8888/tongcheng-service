package user

import (
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"tongcheng/db/mysql"
	"tongcheng/library/ginx"
	"tongcheng/model"
	"tongcheng/repo"
	"tongcheng/utils/common"
)

// GetListReq 用户列表请求
type GetListReq struct {
	common.PageReq
	Username    string `form:"username"` // 用户名模糊搜索（可选）
	PhoneSuffix string `form:"phone_suffix"`
	Status      *int   `form:"status"` // 状态过滤 0:正常 1:禁用（可选，不传则不过滤）
}

type GetListResp struct {
	User        *model.User `json:"user"`
	Phone       string      `json:"phone"`
	PhoneSuffix string      `json:"phone_suffix"`
	Wx          string      `json:"wx"`
}

// GetList 用户列表，直接查 user 表
func GetList(c *gin.Context) {
	ctx := c.Request.Context()
	req := GetListReq{}
	if err := c.ShouldBindQuery(&req); err != nil {
		ginx.ResponseParamError(c)
		return
	}
	req.BindDefault()

	users := make([]*model.User, 0)
	where := map[string]any{
		mysql.Order: "id desc",
	}
	if req.Username != "" {
		where["username LIKE ?"] = mysql.LikeQuery(mysql.Like, req.Username)
	}
	if req.PhoneSuffix != "" {
		where["phone_suffix"] = req.PhoneSuffix
	}
	if req.Status != nil {
		where["status"] = *req.Status
	}
	mysql.SetPage(where, req.Page, req.PageSize)

	total, err := mysql.PageQuery(ctx, mysql.GetUserDB(), &users, where)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	resp := make([]GetListResp, 0, len(users))
	for _, user := range users {
		resp = append(resp, GetListResp{
			User:        user,
			Phone:       user.Phone,
			PhoneSuffix: user.PhoneSuffix,
			Wx:          user.Wx,
		})
	}

	ginx.ResponsePageSuccess(c, resp, req.Page, req.PageSize, total)
}

// ChildDetailReq 查看某用户的子女
type ChildDetailReq struct {
	UserID int `form:"userID" binding:"required"`
}

// ChildDetail 查看子女，一个用户仅对应一条子女信息，按 user_id 查单条
func ChildDetail(c *gin.Context) {
	ctx := c.Request.Context()
	req := ChildDetailReq{}
	if err := c.ShouldBindQuery(&req); err != nil {
		ginx.ResponseParamError(c)
		return
	}

	var child model.UserChild
	err := mysql.FindOne(ctx, mysql.GetUserDB(), &child, map[string]any{"user_id": req.UserID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ginx.ResponseSuccess(c, nil) // 该用户暂无子女信息
			return
		}
		ginx.ResponseServerError(c, err)
		return
	}

	ginx.ResponseSuccess(c, &child)
}

// DisableReq 用户禁用请求
type DisableReq struct {
	UserID int `json:"userID" binding:"required"`
}

// Disable 用户禁用，将 user 表 status 置为 1
func Disable(c *gin.Context) {
	ctx := c.Request.Context()
	req := DisableReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		ginx.ResponseParamError(c)
		return
	}

	_, err := mysql.UpdateByID(ctx, mysql.GetUserDB(), &model.User{}, req.UserID, map[string]any{"status": 1})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	ginx.ResponseSuccess(c, nil)
}

// EnableReq 用户启用请求
type EnableReq struct {
	UserID int `json:"userID" binding:"required"`
}

// Enable 用户启用，将 user 表 status 置为 0
func Enable(c *gin.Context) {
	ctx := c.Request.Context()
	req := EnableReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		ginx.ResponseParamError(c)
		return
	}

	_, err := mysql.UpdateByID(ctx, mysql.GetUserDB(), &model.User{}, req.UserID, map[string]any{"status": 0})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	ginx.ResponseSuccess(c, nil)
}

type GetWatchListReq struct {
	UserID    int `form:"userID"`
	WatcherID int `form:"watcherID"`
}

type GetWatchListResp struct {
	UserWatch    *model.UserWatch `json:"userWatch"`
	User         *model.User      `json:"user"`
	UserChild    *model.UserChild `json:"userChild"`
	Watcher      *model.User      `json:"watcher"`
	WatcherChild *model.UserChild `json:"watcherChild"`
}

func GetWatchList(c *gin.Context) {
	ctx := c.Request.Context()
	req := GetWatchListReq{}
	if err := c.ShouldBindQuery(&req); err != nil {
		ginx.ResponseParamError(c)
		return
	}

	// 关注关系（未删除）
	where := map[string]any{
		"is_delete": 0,
		mysql.Order: "id desc",
	}
	if req.UserID > 0 {
		where["user_id"] = req.UserID
	}
	if req.WatcherID > 0 {
		where["watch_id"] = req.WatcherID
	}
	watches := make([]*model.UserWatch, 0)
	if err := mysql.Find(ctx, mysql.GetUserDB(), &watches, where); err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	resp := make([]GetWatchListResp, 0)
	if len(watches) == 0 {
		ginx.ResponseSuccess(c, resp)
		return
	}

	usersID := make([]int, 0)
	for _, watch := range watches {
		usersID = append(usersID, watch.UserID)
		usersID = append(usersID, watch.WatchID)
	}

	usersMap, err := repo.GetUsersByID(ctx, usersID)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	usersChildMap, err := repo.GetUserChildByID(ctx, usersID)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	for _, watch := range watches {
		resp = append(resp, GetWatchListResp{
			UserWatch:    watch,
			User:         usersMap[watch.UserID],
			UserChild:    usersChildMap[watch.UserID],
			Watcher:      usersMap[watch.WatchID],
			WatcherChild: usersChildMap[watch.WatchID],
		})
	}

	ginx.ResponseSuccess(c, resp)
}
