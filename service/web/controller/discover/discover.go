package discover

import (
	"github.com/gin-gonic/gin"

	"tongcheng/db/mysql"
	"tongcheng/library/ginx"
	"tongcheng/model"
	"tongcheng/repo"
	"tongcheng/utils/common"
	"tongcheng/utils/x/slicex"
)

type GetListReq struct {
	common.PageReq
	Relation int `form:"relation"`
}

type GetListResp struct {
	User  *model.User             `json:"user"`
	Child *model.UserChildDisplay `json:"child"`
	Watch int                     `json:"watch"`
}

func GetList(c *gin.Context) {
	ctx := c.Request.Context()
	userID := ginx.GetUserID(c)
	req := GetListReq{}
	err := c.ShouldBindQuery(&req)
	if err != nil {
		ginx.ResponseParamError(c)
		return
	}
	req.BindDefault()

	discover := make([]*model.UserChildDisplay, 0)
	where := map[string]any{
		"user_id <> ?": userID,
		"status":       1, // 仅展示上架资料
	}
	if req.Relation > 0 {
		where["relation"] = req.Relation
	}
	mysql.SetPage(where, req.Page, req.PageSize)
	total, err := mysql.PageQuery(ctx, mysql.GetUserDB(), &discover, where)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	usersID := make([]int, 0)
	for _, user := range discover {
		usersID = append(usersID, user.ID)
	}

	//家长
	usersMap, err := repo.GetUsersByID(ctx, usersID)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	//关注
	watcher, err := repo.GetUserWatchID(ctx, ginx.GetUserID(c))
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	result := make([]*GetListResp, len(discover))
	for i := range discover {
		watch := 0
		if slicex.ContainInSlice(watcher, discover[i].UserID) {
			watch = 1
		}
		user := &model.User{}
		if _, ok := usersMap[discover[i].UserID]; ok {
			user = usersMap[discover[i].UserID]
		}
		result[i] = &GetListResp{
			User:  user,
			Child: discover[i],
			Watch: watch,
		}
	}

	ginx.ResponsePageSuccess(c, result, req.Page, req.PageSize, total)
}
