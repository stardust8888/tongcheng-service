package child

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"tongcheng/db/mysql"
	"tongcheng/library/ginx"
	"tongcheng/model"
	"tongcheng/repo"
	"tongcheng/utils/common"
	"tongcheng/utils/constant"
)

// GetListReq 子女列表请求，支持多条件筛选（均为可选，不传则不过滤）
type GetListReq struct {
	common.PageReq
	Relation      *int   `form:"relation"`      // 性别（relation字段 1:儿子=男 2:女儿=女）
	MinAge        *int   `form:"minAge"`        // 年龄下限（按年粗算，与birth换算）
	MaxAge        *int   `form:"maxAge"`        // 年龄上限
	MinHeight     *int   `form:"minHeight"`     // 身高下限cm
	MaxHeight     *int   `form:"maxHeight"`     // 身高上限cm
	MinWeight     *int   `form:"minWeight"`     // 体重下限kg
	MaxWeight     *int   `form:"maxWeight"`     // 体重上限kg
	Education     *int   `form:"education"`     // 学历
	MaritalStatus *int   `form:"maritalStatus"` // 婚姻状况
	MaritalPlan   *int   `form:"maritalPlan"`   // 结婚计划
	IsOnlyChild   *int   `form:"isOnlyChild"`   // 独生子女
	House         *int   `form:"house"`         // 房产
	Car           *int   `form:"car"`           // 车辆
	YearIncome    *int   `form:"yearIncome"`    // 年收入
	Residence     string `form:"residence"`     // 居住地关键词（模糊匹配省/市/区，可选）
	Company       *int   `form:"company"`       // 单位性质（公司枚举，精确匹配，可选）
	Work          *int   `form:"work"`          // 工作（职业枚举，精确匹配，可选）
	ReviewStatus  *int   `form:"reviewStatus"`  // 审核状态 0:未审核1:通过2:不通过
	Status        *int   `form:"status"`        // 展示状态 0:未展示1:展示中
}

type GetListResp struct {
	UserChild        *model.UserChild        `json:"userChild"`
	UserChildDisplay *model.UserChildDisplay `json:"userChildDisplay"`
	User             *model.User             `json:"user"`
}

// GetList 获取所有子女列表（user_child 表），多条件筛选 + 分页，附带家长信息
func GetList(c *gin.Context) {
	ctx := c.Request.Context()
	req := GetListReq{}
	if err := c.ShouldBindQuery(&req); err != nil {
		ginx.ResponseParamError(c)
		return
	}
	req.BindDefault()

	children := make([]*model.UserChild, 0)
	where := map[string]any{
		mysql.Order: "updated_at desc",
	}
	if req.Relation != nil {
		where["relation"] = *req.Relation
	}
	if req.MinAge != nil || req.MaxAge != nil {
		year := time.Now().Year()
		if req.MinAge != nil {
			where["birth <= ?"] = fmt.Sprintf("%d-12-31", year-*req.MinAge)
		}
		if req.MaxAge != nil {
			where["birth >= ?"] = fmt.Sprintf("%d-01-01", year-*req.MaxAge)
		}
	}
	if req.MinHeight != nil {
		where["height >= ?"] = *req.MinHeight
	}
	if req.MaxHeight != nil {
		where["height <= ?"] = *req.MaxHeight
	}
	if req.MinWeight != nil {
		where["weight >= ?"] = *req.MinWeight
	}
	if req.MaxWeight != nil {
		where["weight <= ?"] = *req.MaxWeight
	}
	if req.Education != nil {
		where["education"] = *req.Education
	}
	if req.MaritalStatus != nil {
		where["marital_status"] = *req.MaritalStatus
	}
	if req.MaritalPlan != nil {
		where["marital_plan"] = *req.MaritalPlan
	}
	if req.IsOnlyChild != nil {
		where["is_only_child"] = *req.IsOnlyChild
	}
	if req.House != nil {
		where["house"] = *req.House
	}
	if req.Car != nil {
		where["car"] = *req.Car
	}
	if req.YearIncome != nil {
		where["year_income"] = *req.YearIncome
	}
	if req.Residence != "" {
		where["CONCAT(province, city, district) LIKE ?"] = mysql.LikeQuery(mysql.Like, req.Residence)
	}
	if req.Company != nil {
		where["company"] = *req.Company
	}
	if req.Work != nil {
		where["work"] = *req.Work
	}
	if req.ReviewStatus != nil {
		where["review_status"] = *req.ReviewStatus
	}
	if req.Status != nil {
		where["status"] = *req.Status
	}
	mysql.SetPage(where, req.Page, req.PageSize)

	total, err := mysql.PageQuery(ctx, mysql.GetUserDB(), &children, where)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	resp := make([]*GetListResp, 0, len(children))
	if len(children) == 0 {
		ginx.ResponsePageSuccess(c, resp, req.Page, req.PageSize, total)
		return
	}

	// 批量查家长信息，一次查询避免 N+1
	usersID := make([]int, 0, len(children))
	for _, ch := range children {
		usersID = append(usersID, ch.UserID)
	}
	usersMap, err := repo.GetUsersByID(ctx, usersID)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	// 批量查展示快照(user_child_display)：审核通过时落库的"已发布"版本，
	// 后台据此比对哪些字段在审核后又被用户改过，列表上标记"有修改"
	displaysMap, err := repo.GetUserChildDisplayByUserID(ctx, usersID)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	for _, ch := range children {
		item := &GetListResp{UserChild: ch}
		if user, ok := usersMap[ch.UserID]; ok && user != nil {
			item.User = user
		}
		if display, ok := displaysMap[ch.UserID]; ok && display != nil {
			item.UserChildDisplay = display
		}
		resp = append(resp, item)
	}

	ginx.ResponsePageSuccess(c, resp, req.Page, req.PageSize, total)
}

// GetElderReq 查看长辈（注册该子女的用户）资料
type GetElderReq struct {
	UserID int `form:"userID" binding:"required"`
}

// GetElderResp 长辈资料，phone/wx 返回密文（与用户列表一致，前端按需调解密接口）
type GetElderResp struct {
	User        *model.User `json:"user"`
	Phone       string      `json:"phone"`
	PhoneSuffix string      `json:"phone_suffix"`
	Wx          string      `json:"wx"`
}

// GetElder 查看长辈资料，根据 userID 查 user 表，密文返回
func GetElder(c *gin.Context) {
	ctx := c.Request.Context()
	req := GetElderReq{}
	if err := c.ShouldBindQuery(&req); err != nil {
		ginx.ResponseParamError(c)
		return
	}

	user := model.User{}
	if err := mysql.FindByID(ctx, mysql.GetUserDB(), &user, req.UserID); err != nil {
		ginx.ResponseError(c, constant.CodeUserNotExist)
		return
	}

	resp := GetElderResp{
		User:        &user,
		Phone:       user.Phone,
		PhoneSuffix: user.PhoneSuffix,
		Wx:          user.Wx,
	}
	ginx.ResponseSuccess(c, resp)
}

// ApproveReq 子女审核通过
type ApproveReq struct {
	ChildID int `json:"childID" binding:"required"`
}

// Approve 审核通过，review_status 置 1，清空审核原因
func Approve(c *gin.Context) {
	ctx := c.Request.Context()
	req := ApproveReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		ginx.ResponseParamError(c)
		return
	}

	userChild := model.UserChild{}
	err := mysql.FindByID(ctx, mysql.GetUserDB(), &userChild, req.ChildID)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	userChildDisplay := model.UserChildDisplay{}
	exist, err := mysql.Exist(ctx, mysql.GetUserDB(), &userChildDisplay, map[string]any{"user_id": userChild.UserID})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	fn := func(tx *gorm.DB) error {
		_, err = mysql.UpdateByID(ctx, tx, &userChild, req.ChildID, map[string]any{
			"review_status": 1,
			"review_reason": "",
		})
		if err != nil {
			return err
		}
		display := model.UserChildDisplay{
			UserID:        userChild.UserID,
			Relation:      userChild.Relation,
			Height:        userChild.Height,
			Weight:        userChild.Weight,
			Birth:         userChild.Birth,
			Education:     userChild.Education,
			MaritalStatus: userChild.MaritalStatus,
			MaritalPlan:   userChild.MaritalPlan,
			IsOnlyChild:   userChild.IsOnlyChild,
			House:         userChild.House,
			Car:           userChild.Car,
			Company:       userChild.Company,
			Work:          userChild.Work,
			YearIncome:    userChild.YearIncome,
			Province:      userChild.Province,
			City:          userChild.City,
			District:      userChild.District,
			Photo:         userChild.Photo,
			Interest:      userChild.Interest,
			Description:   userChild.Description,
		}
		if !exist {
			err = mysql.Create(ctx, tx, &display)
		} else {
			display.ID = userChildDisplay.ID
			_, err = mysql.Update(ctx, tx, &display)
		}
		if err != nil {
			return err
		}
		return nil
	}
	err = mysql.Transaction(ctx, mysql.GetUserDB(), fn)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	ginx.ResponseSuccess(c, nil)
}

// RejectReq 子女审核拒绝
type RejectReq struct {
	ChildID int    `json:"childID" binding:"required"`
	Reason  string `json:"reason" binding:"required"`
}

// Reject 审核拒绝，review_status 置 2，写入拒绝原因
func Reject(c *gin.Context) {
	ctx := c.Request.Context()
	req := RejectReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		ginx.ResponseParamError(c)
		return
	}

	_, err := mysql.UpdateByID(ctx, mysql.GetUserDB(), &model.UserChild{}, req.ChildID, map[string]any{
		"review_status": 2,
		"review_reason": req.Reason,
	})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	ginx.ResponseSuccess(c, nil)
}

type ShowReq struct {
	ChildID int `json:"childID" binding:"required"`
}

func Show(c *gin.Context) {
	ctx := c.Request.Context()
	req := ShowReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		ginx.ResponseParamError(c)
		return
	}

	userChild := model.UserChild{}
	err := mysql.FindByID(ctx, mysql.GetUserDB(), &userChild, req.ChildID)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	fn := func(tx *gorm.DB) error {
		_, err = mysql.UpdateByID(ctx, tx, &userChild, req.ChildID, map[string]any{
			"status": 1,
		})
		if err != nil {
			return err
		}
		// user_child_display 主键是自增 ID，与 childID 无关；按 user_id 定位该用户的展示行
		_, err = mysql.Updates(ctx, tx, &model.UserChildDisplay{}, map[string]any{"user_id": userChild.UserID}, map[string]any{
			"status": 1,
		})
		if err != nil {
			return err
		}
		return nil
	}
	err = mysql.Transaction(ctx, mysql.GetUserDB(), fn)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	ginx.ResponseSuccess(c, nil)
}

type HiddenReq struct {
	ChildID int `json:"childID" binding:"required"`
}

func Hidden(c *gin.Context) {
	ctx := c.Request.Context()
	req := HiddenReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		ginx.ResponseParamError(c)
		return
	}

	userChild := model.UserChild{}
	err := mysql.FindByID(ctx, mysql.GetUserDB(), &userChild, req.ChildID)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	fn := func(tx *gorm.DB) error {
		_, err = mysql.UpdateByID(ctx, tx, &userChild, req.ChildID, map[string]any{
			"status": 0,
		})
		if err != nil {
			return err
		}
		// user_child_display 主键是自增 ID，与 childID 无关；按 user_id 定位该用户的展示行
		_, err = mysql.Updates(ctx, tx, &model.UserChildDisplay{}, map[string]any{"user_id": userChild.UserID}, map[string]any{
			"status": 0,
		})
		if err != nil {
			return err
		}
		return nil
	}
	err = mysql.Transaction(ctx, mysql.GetUserDB(), fn)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	ginx.ResponseSuccess(c, nil)
}
