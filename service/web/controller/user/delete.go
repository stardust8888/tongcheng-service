package user

import (
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"tongcheng/db/mysql"
	"tongcheng/library/ginx"
	"tongcheng/model"
	"tongcheng/repo"
	"tongcheng/utils/x/timex"
)

type ApplyDeleteResp struct {
	ExpectAt string `json:"expectAt"` // 预计正式注销时间
}

// ApplyDelete 用户提交注销申请。仅记录申请，15 天后由 service/cron 定时任务正式处理；
// 同一用户存在待处理申请时拒绝重复提交。
func ApplyDelete(c *gin.Context) {
	ctx := c.Request.Context()
	userID := ginx.GetUserID(c)

	exist, err := mysql.Exist(ctx, mysql.GetUserDB(), &model.UserDeleteApply{}, map[string]any{
		"user_id": userID,
		"status":  0,
	})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	if exist {
		ginx.ResponseNotify(c, "您已提交注销申请，正在处理中")
		return
	}

	now := timex.GetBeijingTime()
	err = mysql.Create(ctx, mysql.GetUserDB(), &model.UserDeleteApply{
		UserID:  userID,
		ApplyAt: now,
		Status:  0,
	})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	ginx.ResponseSuccess(c, ApplyDeleteResp{
		ExpectAt: now.Add(repo.DeleteApplyDelay).Format(),
	})
}

type DeleteApplyStatusResp struct {
	Applied  bool   `json:"applied"`            // 是否存在待处理注销申请
	ApplyAt  string `json:"applyAt,omitempty"`  // 申请时间
	ExpectAt string `json:"expectAt,omitempty"` // 预计正式注销时间
}

// GetDeleteApply 查询当前用户待处理的注销申请状态，供前端展示「已申请·可取消」状态条。
func GetDeleteApply(c *gin.Context) {
	ctx := c.Request.Context()
	userID := ginx.GetUserID(c)

	var apply model.UserDeleteApply
	err := mysql.FindOne(ctx, mysql.GetUserDB(), &apply, map[string]any{
		"user_id": userID,
		"status":  0,
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ginx.ResponseSuccess(c, DeleteApplyStatusResp{Applied: false})
		return
	}
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	ginx.ResponseSuccess(c, DeleteApplyStatusResp{
		Applied:  true,
		ApplyAt:  apply.ApplyAt.Format(),
		ExpectAt: apply.ApplyAt.Add(repo.DeleteApplyDelay).Format(),
	})
}

// CancelDelete 取消当前用户待处理的注销申请（15 天冷静期内可撤销）。
// 不删除申请记录，而是置 status=2（已取消）并记录处理时间，保留全生命周期留痕；
// 已处理（status=1，账号已删除）的申请不可取消。
func CancelDelete(c *gin.Context) {
	ctx := c.Request.Context()
	userID := ginx.GetUserID(c)

	rows, err := mysql.Updates(ctx, mysql.GetUserDB(), &model.UserDeleteApply{},
		map[string]any{"user_id": userID, "status": 0},
		map[string]any{"status": 2, "process_at": timex.GetBeijingTime()},
	)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	if rows == 0 {
		ginx.ResponseNotify(c, "无待处理的注销申请")
		return
	}

	ginx.ResponseSuccess(c, nil)
}
