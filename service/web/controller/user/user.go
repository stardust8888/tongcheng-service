package user

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"tongcheng/db/mysql"
	"tongcheng/library/ginx"
	"tongcheng/logs"
	"tongcheng/model"
	"tongcheng/repo"
	"tongcheng/utils/common"
	"tongcheng/utils/constant"
)

type GetProfileResp struct {
	User  *model.User      `json:"user"`
	Child *model.UserChild `json:"child"`
}

func GetProfile(c *gin.Context) {
	ctx := c.Request.Context()
	userID := ginx.GetUserID(c)
	user := model.User{}
	err := mysql.FindByID(ctx, mysql.GetUserDB(), &user, userID)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	user.Wx = repo.DecryptUserInfo(ctx, userID, user.Wx)

	userChild := model.UserChild{}
	_, err = mysql.Exist(ctx, mysql.GetUserDB(), &userChild, map[string]any{"user_id": userID})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	resp := GetProfileResp{
		User:  &user,
		Child: &userChild,
	}
	ginx.ResponseSuccess(c, resp)
}

type GetHomeReq struct {
	UserID int `form:"user_id"`
}

type GetHomeResp struct {
	User    *model.User        `json:"user"`
	Child   any                `json:"child"`
	Watch   int                `json:"watch"`
	Moments []*repo.UserMoment `json:"moments"`
}

func GetHome(c *gin.Context) {
	ctx := c.Request.Context()
	req := GetHomeReq{}
	userID := ginx.GetUserID(c)
	err := c.ShouldBindQuery(&req)
	if err != nil {
		ginx.ResponseParamError(c)
		return
	}

	resp := GetHomeResp{}
	targetID := req.UserID

	if userID != targetID {
		blacked, err := repo.IsBlacked(ctx, targetID, userID)
		if err != nil {
			ginx.ResponseServerError(c, err)
			return
		}
		if blacked {
			ginx.ResponseNotify(c, "无法访问对方主页")
			return
		}
	}

	user := model.User{}
	err = mysql.FindByID(ctx, mysql.GetUserDB(), &user, targetID)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	resp.User = &user

	//访问记录
	if userID != targetID {
		_ = mysql.Create(ctx, mysql.GetUserDB(), &model.UserVisitRecord{
			UserID:    targetID,
			VisiterID: userID,
			Page:      1,
		})
	}

	if userID == targetID {
		userChild := &model.UserChild{}
		_, err = mysql.Exist(ctx, mysql.GetUserDB(), userChild, map[string]any{"user_id": targetID})
		if err != nil {
			ginx.ResponseServerError(c, err)
			return
		}
		resp.Child = userChild
	} else {
		userChild := &model.UserChildDisplay{}
		_, err = mysql.Exist(ctx, mysql.GetUserDB(), userChild, map[string]any{"user_id": targetID})
		if err != nil {
			ginx.ResponseServerError(c, err)
			return
		}
		resp.Child = userChild
	}

	if userID != targetID {
		exist, err := mysql.Exist(ctx, mysql.GetUserDB(), &model.UserWatch{}, map[string]any{
			"user_id":  userID,
			"watch_id": targetID,
		})
		if err != nil {
			ginx.ResponseServerError(c, err)
			return
		}
		if exist {
			resp.Watch = 1
		}
	}

	userMoments := make([]*model.UserMoment, 0)
	where := map[string]any{
		"user_id":       targetID,
		"review_status": 1,
		"status":        1,
		"is_delete":     0,
		mysql.Order:     "id desc",
	}
	mysql.SetPage(where, 1, 10)
	_, err = mysql.PageQuery(ctx, mysql.GetUserDB(), &userMoments, where)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	repoMoments, err := repo.GetUserMoments(ctx, userMoments, userID)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	err = repo.GetMomentComments(ctx, repoMoments, userID)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	resp.Moments = repoMoments
	ginx.ResponseSuccess(c, resp)
}

func UpdateProfile(c *gin.Context) {
	ctx := c.Request.Context()
	userID := ginx.GetUserID(c)
	req := model.User{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		ginx.ResponseParamError(c)
		return
	}

	user := model.User{}
	err = mysql.FindByID(ctx, mysql.GetUserDB(), &user, userID)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	upMap := map[string]interface{}{}
	if req.Username != "" {
		upMap["username"] = req.Username
	}
	if req.Avatar != "" {
		upMap["avatar"] = req.Avatar
	}
	if req.Role != 0 {
		upMap["role"] = req.Role
	}
	if req.Wx != "" {
		upMap["wx"] = repo.EncryptUserInfo(ctx, userID, req.Wx)
	}
	if req.Promoter != "" && user.Promoter == "" {
		upMap["promoter"] = req.Promoter
	}

	_, err = mysql.UpdateByID(ctx, mysql.GetUserDB(), &model.User{}, userID, upMap)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	ginx.ResponseSuccess(c, nil)
}

type GetChildReq struct {
	ChildID int `json:"childID"`
}

func GetChild(c *gin.Context) {
	ctx := c.Request.Context()
	req := GetChildReq{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		ginx.ResponseParamError(c)
		return
	}

	child := model.UserChild{}
	exist, err := mysql.Exist(ctx, mysql.GetUserDB(), &child, map[string]any{"id": req.ChildID})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	if !exist {
		ginx.ResponseNotify(c, "您尚未添加子女")
		return
	}

	ginx.ResponseSuccess(c, child)
}

func CreateChild(c *gin.Context) {
	ctx := c.Request.Context()
	userID := ginx.GetUserID(c)
	req := model.UserChild{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		ginx.ResponseParamError(c)
		return
	}

	exist, err := mysql.Exist(ctx, mysql.GetUserDB(), &req, map[string]any{"user_id": userID})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	if exist {
		ginx.ResponseError(c, constant.CodeUserOneChildOnly)
		return
	}

	if repo.CheckChildText(ctx, &req) {
		ginx.ResponseNotify(c, "资料含违规内容，请修改后重试")
		return
	}

	req.UserID = userID
	err = mysql.Create(ctx, mysql.GetUserDB(), &req)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	ginx.ResponseSuccess(c, nil)
}

func UpdateChild(c *gin.Context) {
	ctx := c.Request.Context()
	userID := ginx.GetUserID(c)
	req := model.UserChild{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		logs.Error(c, "UpdateChild Param error", zap.Error(err))
		ginx.ResponseParamError(c)
		return
	}
	if req.ID <= 0 {
		ginx.ResponseParamError(c)
		return
	}

	userChild := model.UserChild{}
	err = mysql.FindByID(ctx, mysql.GetUserDB(), &userChild, req.ID)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	if userID != userChild.UserID {
		ginx.ResponseNotify(c, "没有操作权限")
		return
	}

	if repo.CheckChildText(ctx, &req) {
		ginx.ResponseNotify(c, "资料含违规内容，请修改后重试")
		return
	}

	// 与库内记录比对，仅把有差异的字段加入 map 更新
	upMap := map[string]any{}
	if req.Relation != userChild.Relation {
		upMap["relation"] = req.Relation
	}
	if req.Birth != userChild.Birth {
		upMap["birth"] = req.Birth
	}
	if req.Height != userChild.Height {
		upMap["height"] = req.Height
	}
	if req.Weight != userChild.Weight {
		upMap["weight"] = req.Weight
	}
	if req.Education != userChild.Education {
		upMap["education"] = req.Education
	}
	if req.MaritalStatus != userChild.MaritalStatus {
		upMap["marital_status"] = req.MaritalStatus
	}
	if req.MaritalPlan != userChild.MaritalPlan {
		upMap["marital_plan"] = req.MaritalPlan
	}
	if req.IsOnlyChild != userChild.IsOnlyChild {
		upMap["is_only_child"] = req.IsOnlyChild
	}
	if req.House != userChild.House {
		upMap["house"] = req.House
	}
	if req.Car != userChild.Car {
		upMap["car"] = req.Car
	}
	if req.Company != userChild.Company {
		upMap["company"] = req.Company
	}
	if req.Work != userChild.Work {
		upMap["work"] = req.Work
	}
	if req.YearIncome != userChild.YearIncome {
		upMap["year_income"] = req.YearIncome
	}
	if req.Province != userChild.Province {
		upMap["province"] = req.Province
	}
	if req.City != userChild.City {
		upMap["city"] = req.City
	}
	if req.District != userChild.District {
		upMap["district"] = req.District
	}
	if req.Photo != userChild.Photo {
		upMap["photo"] = req.Photo
	}
	if req.Interest != userChild.Interest {
		upMap["interest"] = req.Interest
	}
	if req.Description != userChild.Description {
		upMap["description"] = req.Description
	}

	if len(upMap) == 0 {
		ginx.ResponseSuccess(c, nil)
		return
	}
	upMap["review_status"] = 0

	_, err = mysql.UpdateByID(ctx, mysql.GetUserDB(), &userChild, req.ID, upMap)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	ginx.ResponseSuccess(c, nil)
}

type GetWatchListResp struct {
	User  *model.User      `json:"user"`
	Child *model.UserChild `json:"child"`
}

func GetWatchList(c *gin.Context) {
	ctx := c.Request.Context()
	userWatch := make([]*model.UserWatch, 0)
	err := mysql.Find(ctx, mysql.GetUserDB(), &userWatch, map[string]any{"user_id": ginx.GetUserID(c), "is_delete": 0})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	watchersID := make([]int, 0, len(userWatch))
	for _, w := range userWatch {
		watchersID = append(watchersID, w.WatchID)
	}

	if len(watchersID) == 0 {
		ginx.ResponseSuccess(c, []*GetWatchListResp{})
		return
	}

	users := make([]*model.User, 0)
	err = mysql.Find(ctx, mysql.GetUserDB(), &users, map[string]any{"id": watchersID})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	children := make([]*model.UserChild, 0)
	err = mysql.Find(ctx, mysql.GetUserDB(), &children, map[string]any{"user_id": watchersID})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	userMap := make(map[int]*model.User, len(users))
	for _, user := range users {
		userMap[user.ID] = user
	}
	childrenMap := make(map[int]*model.UserChild, len(children))
	for _, child := range children {
		childrenMap[child.UserID] = child
	}

	result := make([]*GetWatchListResp, 0, len(watchersID))
	for _, userID := range watchersID {
		user := userMap[userID]
		result = append(result, &GetWatchListResp{
			User:  user,
			Child: childrenMap[userID],
		})
	}

	ginx.ResponseSuccess(c, result)
}

type WatchReq struct {
	UserID int `json:"userID"`
}

func Watch(c *gin.Context) {
	ctx := c.Request.Context()
	req := WatchReq{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		ginx.ResponseParamError(c)
		return
	}
	userID := ginx.GetUserID(c)
	if userID == req.UserID {
		ginx.ResponseNotify(c, "不用点自己的关注")
		return
	}

	blacked, err := repo.IsBlacked(ctx, req.UserID, userID)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	if blacked {
		ginx.ResponseNotify(c, "对方已将你加入黑名单")
		return
	}

	exist, err := mysql.Exist(ctx, mysql.GetUserDB(), &model.UserWatch{}, map[string]any{
		"user_id":  userID,
		"watch_id": req.UserID,
	})
	if err != nil {
		ginx.ResponseParamError(c)
		return
	}
	if exist {
		ginx.ResponseNotify(c, "已关注")
		return
	}

	userWatch := &model.UserWatch{
		UserID:  userID,
		WatchID: req.UserID,
	}
	err = mysql.Create(ctx, mysql.GetUserDB(), userWatch)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	ginx.ResponseSuccess(c, nil)
}

type CancelWatchReq struct {
	UserID int `json:"userID"`
}

func WatchCancel(c *gin.Context) {
	ctx := c.Request.Context()
	req := WatchReq{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		ginx.ResponseParamError(c)
		return
	}

	userWatch := model.UserWatch{}
	exist, err := mysql.Exist(ctx, mysql.GetUserDB(), &userWatch, map[string]any{
		"user_id":   ginx.GetUserID(c),
		"watch_id":  req.UserID,
		"is_delete": 0,
	})
	if err != nil {
		ginx.ResponseParamError(c)
		return
	}
	if !exist {
		ginx.ResponseNotify(c, "已取消关注")
		return
	}

	err = mysql.SoftDelete(ctx, mysql.GetUserDB(), &userWatch, map[string]any{
		"user_id":   ginx.GetUserID(c),
		"watch_id":  req.UserID,
		"is_delete": 0,
	})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	ginx.ResponseSuccess(c, nil)
}

type GetInteractionStatResp struct {
	View    int `json:"view"`
	Watch   int `json:"watch"`
	Like    int `json:"like"`
	Comment int `json:"comment"`
}

func GetInteractionStat(c *gin.Context) {
	ctx := c.Request.Context()
	userID := ginx.GetUserID(c)
	view, err := mysql.Count(ctx, mysql.GetUserDB(), &model.UserVisitRecord{}, map[string]any{"user_id": userID})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	watch, err := mysql.Count(ctx, mysql.GetUserDB(), &model.UserWatch{}, map[string]any{"watch_id": userID, "is_delete": 0})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	userMoment := model.UserMoment{}
	sel := "COALESCE(SUM(like_count),0) as like_count, COALESCE(SUM(comment_count),0) as comment_count, " +
		"COALESCE(SUM(self_like_count),0) as self_like_count, COALESCE(SUM(self_comment_count),0) as self_comment_count"
	err = mysql.Aggregate(ctx, mysql.GetUserDB(), &userMoment, sel, map[string]any{
		"user_id": userID,
	})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	resp := &GetInteractionStatResp{
		View:    view,
		Watch:   watch,
		Like:    userMoment.LikeCount - userMoment.SelfLikeCount,
		Comment: userMoment.CommentCount - userMoment.SelfCommentCount,
	}

	ginx.ResponseSuccess(c, resp)
}

type VisitMeReq struct {
	common.PageReq
}

type VisitMeResp struct {
	Record *model.UserVisitRecord `json:"record"`
	User   *model.User            `json:"user"`
	Child  *model.UserChild       `json:"child"`
}

func ViewMe(c *gin.Context) {
	ctx := c.Request.Context()
	userID := ginx.GetUserID(c)
	req := VisitMeReq{}
	err := c.ShouldBindQuery(&req)
	if err != nil {
		ginx.ResponseParamError(c)
		return
	}

	visitMeRecord := make([]*model.UserVisitRecord, 0)
	where := map[string]any{
		"user_id":   userID,
		mysql.Order: "id desc",
	}
	mysql.SetPage(where, req.Page, req.PageSize)
	total, err := mysql.PageQuery(ctx, mysql.GetUserDB(), &visitMeRecord, where)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	visitorsID := make([]int, 0)
	for _, record := range visitMeRecord {
		visitorsID = append(visitorsID, record.VisiterID)
	}

	usersMap, err := repo.GetUsersByID(ctx, visitorsID)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	usersChildMap, err := repo.GetUserChildByID(ctx, visitorsID)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	resp := make([]*VisitMeResp, 0)
	for _, record := range visitMeRecord {
		visiter := &model.User{}
		if user, ok := usersMap[record.VisiterID]; ok {
			visiter = user
		}
		visiterChild := &model.UserChild{}
		if child, ok := usersChildMap[record.VisiterID]; ok {
			visiterChild = child
		}
		resp = append(resp, &VisitMeResp{
			Record: record,
			User:   visiter,
			Child:  visiterChild,
		})
	}

	ginx.ResponsePageSuccess(c, resp, req.Page, req.PageSize, total)
}

type WatchMeResp struct {
	Record *model.UserWatch `json:"record"`
	User   *model.User      `json:"user"`
	Child  *model.UserChild `json:"child"`
}

func WatchMe(c *gin.Context) {
	ctx := c.Request.Context()
	userID := ginx.GetUserID(c)
	req := common.PageReq{}
	err := c.ShouldBindQuery(&req)
	if err != nil {
		ginx.ResponseParamError(c)
		return
	}

	watchRecords := make([]*model.UserWatch, 0)
	where := map[string]any{
		"watch_id":  userID,
		"is_delete": 0,
		mysql.Order: "id desc",
	}
	mysql.SetPage(where, req.Page, req.PageSize)
	total, err := mysql.PageQuery(ctx, mysql.GetUserDB(), &watchRecords, where)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	watchersID := make([]int, 0)
	for _, record := range watchRecords {
		watchersID = append(watchersID, record.UserID)
	}

	usersMap, err := repo.GetUsersByID(ctx, watchersID)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	usersChildMap, err := repo.GetUserChildByID(ctx, watchersID)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	resp := make([]*WatchMeResp, 0)
	for _, record := range watchRecords {
		watcher := &model.User{}
		if u, ok := usersMap[record.UserID]; ok {
			watcher = u
		}
		watcherChild := &model.UserChild{}
		if child, ok := usersChildMap[record.UserID]; ok {
			watcherChild = child
		}
		resp = append(resp, &WatchMeResp{
			Record: record,
			User:   watcher,
			Child:  watcherChild,
		})
	}

	ginx.ResponsePageSuccess(c, resp, req.Page, req.PageSize, total)
}

type LikeMeResp struct {
	Record *model.MomentLike `json:"record"`
	// Moment *model.UserMoment `json:"userMoment`
	User  *model.User      `json:"user"`
	Child *model.UserChild `json:"child"`
}

func LikeMe(c *gin.Context) {
	ctx := c.Request.Context()
	userID := ginx.GetUserID(c)
	req := common.PageReq{}
	err := c.ShouldBindQuery(&req)
	if err != nil {
		ginx.ResponseParamError(c)
		return
	}

	moments := make([]*model.UserMoment, 0)
	err = mysql.Find(ctx, mysql.GetUserDB(), &moments, map[string]any{"user_id": userID, "is_delete": 0})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	momentsID := make([]int, 0)
	for _, moment := range moments {
		momentsID = append(momentsID, moment.ID)
	}

	if len(momentsID) == 0 {
		ginx.ResponsePageSuccess(c, []*LikeMeResp{}, req.Page, req.PageSize, 0)
		return
	}

	likeRecords := make([]*model.MomentLike, 0)
	where := map[string]any{
		"moment_id":    momentsID,
		"user_id <> ?": userID,
		mysql.Order:    "id desc",
	}
	mysql.SetPage(where, req.Page, req.PageSize)
	total, err := mysql.PageQuery(ctx, mysql.GetUserDB(), &likeRecords, where)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	likesID := make([]int, 0)
	for _, record := range likeRecords {
		likesID = append(likesID, record.UserID)
	}

	usersMap, err := repo.GetUsersByID(ctx, likesID)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	usersChildMap, err := repo.GetUserChildByID(ctx, likesID)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	resp := make([]*LikeMeResp, 0)
	for _, record := range likeRecords {
		liker := &model.User{}
		if u, ok := usersMap[record.UserID]; ok {
			liker = u
		}
		likerChild := &model.UserChild{}
		if child, ok := usersChildMap[record.UserID]; ok {
			likerChild = child
		}
		resp = append(resp, &LikeMeResp{
			Record: record,
			User:   liker,
			Child:  likerChild,
		})
	}

	ginx.ResponsePageSuccess(c, resp, req.Page, req.PageSize, total)
}

type CommentMeResp struct {
	Record *model.MomentComment `json:"record"`
	User   *model.User          `json:"user"`
	Child  *model.UserChild     `json:"child"`
}

func CommentMe(c *gin.Context) {
	ctx := c.Request.Context()
	userID := ginx.GetUserID(c)
	req := common.PageReq{}
	err := c.ShouldBindQuery(&req)
	if err != nil {
		ginx.ResponseParamError(c)
		return
	}

	moments := make([]*model.UserMoment, 0)
	err = mysql.Find(ctx, mysql.GetUserDB(), &moments, map[string]any{"user_id": userID, "is_delete": 0})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	momentsID := make([]int, 0)
	for _, moment := range moments {
		momentsID = append(momentsID, moment.ID)
	}

	if len(momentsID) == 0 {
		ginx.ResponsePageSuccess(c, []*CommentMeResp{}, req.Page, req.PageSize, 0)
		return
	}

	commentRecords := make([]*model.MomentComment, 0)
	where := map[string]any{
		"moment_id":    momentsID,
		"user_id <> ?": userID,
		"is_delete":    0,
		mysql.Order:    "id desc",
	}
	mysql.SetPage(where, req.Page, req.PageSize)
	total, err := mysql.PageQuery(ctx, mysql.GetUserDB(), &commentRecords, where)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	commentersID := make([]int, 0)
	for _, record := range commentRecords {
		commentersID = append(commentersID, record.UserID)
	}

	usersMap, err := repo.GetUsersByID(ctx, commentersID)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	usersChildMap, err := repo.GetUserChildByID(ctx, commentersID)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	resp := make([]*CommentMeResp, 0)
	for _, record := range commentRecords {
		commenter := &model.User{}
		if u, ok := usersMap[record.UserID]; ok {
			commenter = u
		}
		commenterChild := &model.UserChild{}
		if child, ok := usersChildMap[record.UserID]; ok {
			commenterChild = child
		}
		resp = append(resp, &CommentMeResp{
			Record: record,
			User:   commenter,
			Child:  commenterChild,
		})
	}

	ginx.ResponsePageSuccess(c, resp, req.Page, req.PageSize, total)
}

type BlackListResp struct {
	Record *model.UserBlackList `json:"record"`
	User   *model.User          `json:"user"`
}

func GetBlackList(c *gin.Context) {
	ctx := c.Request.Context()
	userID := ginx.GetUserID(c)
	req := common.PageReq{}
	err := c.ShouldBindQuery(&req)
	if err != nil {
		ginx.ResponseParamError(c)
		return
	}
	req.BindDefault()

	records := make([]*model.UserBlackList, 0)
	where := map[string]any{
		"user_id":   userID,
		mysql.Order: "id desc",
	}
	mysql.SetPage(where, req.Page, req.PageSize)
	total, err := mysql.PageQuery(ctx, mysql.GetUserDB(), &records, where)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	blackUsersID := make([]int, 0)
	for _, record := range records {
		blackUsersID = append(blackUsersID, record.BlackUserID)
	}

	usersMap, err := repo.GetUsersByID(ctx, blackUsersID)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	resp := make([]*BlackListResp, 0)
	for _, record := range records {
		u := &model.User{}
		if user, ok := usersMap[record.BlackUserID]; ok {
			u = user
		}
		resp = append(resp, &BlackListResp{
			Record: record,
			User:   u,
		})
	}

	ginx.ResponsePageSuccess(c, resp, req.Page, req.PageSize, total)
}

type AddBlackReq struct {
	BlackUserID int `json:"blackUserID" binding:"required"`
}

func AddBlack(c *gin.Context) {
	ctx := c.Request.Context()
	userID := ginx.GetUserID(c)
	req := AddBlackReq{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		ginx.ResponseParamError(c)
		return
	}

	if userID == req.BlackUserID {
		ginx.ResponseNotify(c, "不能拉黑自己")
		return
	}

	exist, err := mysql.Exist(ctx, mysql.GetUserDB(), &model.UserBlackList{}, map[string]any{
		"user_id":       userID,
		"black_user_id": req.BlackUserID,
	})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	if exist {
		ginx.ResponseSuccess(c, nil)
		return
	}

	err = mysql.Create(ctx, mysql.GetUserDB(), &model.UserBlackList{
		UserID:      userID,
		BlackUserID: req.BlackUserID,
	})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	_ = repo.AddBlackCache(ctx, userID, req.BlackUserID)
	ginx.ResponseSuccess(c, nil)
}

type RemoveBlackReq struct {
	BlackUserID int `json:"blackUserID" binding:"required"`
}

func RemoveBlack(c *gin.Context) {
	ctx := c.Request.Context()
	userID := ginx.GetUserID(c)
	req := RemoveBlackReq{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		ginx.ResponseParamError(c)
		return
	}

	err = mysql.Delete(ctx, mysql.GetUserDB(), &model.UserBlackList{}, map[string]any{
		"user_id":       userID,
		"black_user_id": req.BlackUserID,
	})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	_ = repo.RemoveBlackCache(ctx, userID, req.BlackUserID)
	ginx.ResponseSuccess(c, nil)
}
