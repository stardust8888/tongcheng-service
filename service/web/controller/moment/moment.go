package moment

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"tongcheng/db/mysql"
	"tongcheng/library/ginx"
	"tongcheng/model"
	"tongcheng/repo"
	"tongcheng/utils/common"
	"tongcheng/utils/constant"
	"tongcheng/utils/x/slicex"
)

type GetListWithCommentReq struct {
	common.PageReq
	MyList int `form:"myList"`
}

func GetListWithComment(c *gin.Context) {
	ctx := c.Request.Context()
	userID := ginx.GetUserID(c)
	req := GetListWithCommentReq{}
	err := c.ShouldBindQuery(&req)
	if err != nil {
		ginx.ResponseParamError(c)
		return
	}
	req.BindDefault()

	userMoments := make([]*model.UserMoment, 0)
	where := map[string]any{
		"is_delete": 0,
		mysql.Order: "id desc",
	}
	if req.MyList == 1 {
		where["user_id"] = userID
	} else {
		where["review_status"] = 1
		where["status"] = 1
	}
	mysql.SetPage(where, req.Page, req.PageSize)
	total, err := mysql.PageQuery(ctx, mysql.GetUserDB(), &userMoments, where)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	//动态
	result, err := repo.GetUserMoments(ctx, userMoments, userID)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	err = repo.GetMomentComments(ctx, result, userID)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	ginx.ResponsePageSuccess(c, result, req.Page, req.PageSize, total)
}

type GetDetailWithCommentReq struct {
	MomentID int `form:"moment_id"`
}

type GetDetailWithCommentResp struct {
	UserMoment *repo.UserMoment `json:"userMoment"`
}

func GetDetailWithComment(c *gin.Context) {
	ctx := c.Request.Context()
	userID := ginx.GetUserID(c)
	req := GetDetailWithCommentReq{}
	err := c.ShouldBindQuery(&req)
	if err != nil {
		ginx.ResponseParamError(c)
		return
	}
	if req.MomentID <= 0 {
		ginx.ResponseParamError(c)
		return
	}

	userMoment := model.UserMoment{}
	exist, err := mysql.Exist(ctx, mysql.GetUserDB(), &userMoment, map[string]any{"id": req.MomentID, "is_delete": 0})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	if !exist {
		ginx.ResponseNotify(c, "动态已被删除")
		return
	}

	repoMoments, err := repo.GetUserMoments(ctx, []*model.UserMoment{&userMoment}, userID)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	err = repo.GetMomentComments(ctx, repoMoments, userID)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	if len(repoMoments) == 0 {
		ginx.ResponseNotify(c, "动态已被删除")
		return
	}

	resp := GetDetailWithCommentResp{
		UserMoment: repoMoments[0],
	}
	ginx.ResponseSuccess(c, resp)
}

func Create(c *gin.Context) {
	ctx := c.Request.Context()
	req := model.UserMoment{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		ginx.ResponseParamError(c)
		return
	}
	if len(req.Content) == 0 {
		ginx.ResponseParamError(c)
		return
	}

	// 内容安全：正文命中违规则禁止发布(图片在上传时已 img_sec_check)
	if repo.CheckMomentText(ctx, req.Content) {
		ginx.ResponseNotify(c, "内容含违规信息，请修改后发布")
		return
	}

	req.UserID = ginx.GetUserID(c)
	err = mysql.Create(ctx, mysql.GetUserDB(), &req)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	ginx.ResponseSuccess(c, req)
}

type DeleteReq struct {
	MomentID int `form:"moment_id" json:"moment_id" binding:"required"`
}

func Delete(c *gin.Context) {
	ctx := c.Request.Context()
	userID := ginx.GetUserID(c)
	req := &DeleteReq{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		ginx.ResponseParamError(c)
		return
	}
	if req.MomentID <= 0 {
		ginx.ResponseParamError(c)
		return
	}

	userMoment := model.UserMoment{}
	condition := map[string]any{"id": req.MomentID, "user_id": userID, "is_delete": 0}
	exist, err := mysql.Exist(ctx, mysql.GetUserDB(), &userMoment, condition)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	if !exist || userMoment.IsDelete == 1 {
		ginx.ResponseSuccess(c, nil)
		return
	}
	if userID != userMoment.UserID {
		ginx.ResponseNotify(c, "只能删除自己的动态")
		return
	}

	err = mysql.SoftDelete(ctx, mysql.GetUserDB(), &userMoment, condition)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	ginx.ResponseSuccess(c, nil)
}

type LikeReq struct {
	MomentID int `json:"momentID" binding:"required"`
}

func Like(c *gin.Context) {
	ctx := c.Request.Context()
	userID := ginx.GetUserID(c)
	req := &LikeReq{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		ginx.ResponseParamError(c)
		return
	}
	if req.MomentID <= 0 {
		ginx.ResponseParamError(c)
		return
	}

	moment := model.UserMoment{}
	exist, err := mysql.Exist(ctx, mysql.GetUserDB(), &moment, map[string]any{"id": req.MomentID, "is_delete": 0})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	if !exist {
		ginx.ResponseNotify(c, "动态已被删除")
		return
	}

	blacked, err := repo.IsBlacked(ctx, moment.UserID, userID)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	if blacked {
		ginx.ResponseNotify(c, "对方已将你加入黑名单")
		return
	}

	momentLike := model.MomentLike{}
	exist, err = mysql.Exist(ctx, mysql.GetUserDB(), &momentLike, map[string]any{"user_id": userID, "moment_id": req.MomentID, "is_delete": 0})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	if exist {
		ginx.ResponseNotify(c, "已成功点赞")
		return
	}

	fn := func(tx *gorm.DB) error {
		err = tx.Create(&model.MomentLike{
			UserID:   userID,
			MomentID: req.MomentID,
		}).Error
		if err != nil {
			return err
		}

		update := map[string]any{
			"like_count": gorm.Expr("like_count + ?", 1),
		}
		if moment.UserID == userID {
			update["self_like_count"] = gorm.Expr("self_like_count + ?", 1)
		}
		err = tx.Model(&model.UserMoment{}).Where("id = ?", req.MomentID).Updates(update).Error
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

type UnLikeReq struct {
	MomentID int `json:"momentID" binding:"required"`
}

func UnLike(c *gin.Context) {
	ctx := c.Request.Context()
	userID := ginx.GetUserID(c)
	if userID <= 0 {
		ginx.ResponseParamError(c)
		return
	}
	req := &UnLikeReq{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		ginx.ResponseParamError(c)
		return
	}
	if req.MomentID <= 0 {
		ginx.ResponseParamError(c)
		return
	}

	moment := model.UserMoment{}
	exist, err := mysql.Exist(ctx, mysql.GetUserDB(), &moment, map[string]any{"id": req.MomentID, "is_delete": 0})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	if !exist {
		ginx.ResponseNotify(c, "动态已被删除")
		return
	}

	momentLike := model.MomentLike{}
	exist, err = mysql.Exist(ctx, mysql.GetUserDB(), &momentLike, map[string]any{"user_id": userID, "moment_id": req.MomentID, "is_delete": 0})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	if !exist {
		ginx.ResponseNotify(c, "已取消点赞")
		return
	}

	if momentLike.UserID != userID {
		ginx.ResponseError(c, constant.CodePermissionDenied)
		return
	}

	fn := func(tx *gorm.DB) error {
		err = mysql.SoftDelete(ctx, tx, &momentLike, map[string]any{"user_id": userID, "moment_id": req.MomentID, "is_delete": 0})
		if err != nil {
			return err
		}

		//不符合取消点赞条件的跳过
		update := map[string]any{
			"like_count": gorm.Expr("like_count - ?", 1),
		}
		if moment.UserID == userID {
			update["self_like_count"] = gorm.Expr("self_like_count - ?", 1)
		}
		_, err = mysql.UpdateByID(ctx, tx, &moment, moment.ID, update)
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

type GetCommentListReq struct {
	common.PageReq
	MomentID int `form:"momentID" json:"momentID" binding:"required"`
}

type GetCommentListResp struct {
	UserInfo    *model.User          `json:"userInfo"`
	ReplierInfo *model.User          `json:"replierInfo"`
	Comment     *model.MomentComment `json:"comment"`
}

func GetCommentList(c *gin.Context) {
	ctx := c.Request.Context()
	req := GetCommentListReq{}
	err := c.ShouldBindQuery(&req)
	if err != nil {
		ginx.ResponseParamError(c)
		return
	}
	req.BindDefault()

	comments := make([]*model.MomentComment, 0)
	where := map[string]any{
		"moment_id": req.MomentID,
		"is_delete": 0,
	}
	mysql.SetPage(where, req.Page, req.PageSize)
	total, err := mysql.PageQuery(ctx, mysql.GetUserDB(), &comments, where)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	usersID := make([]int, 0)
	commentersID := make([]int, 0)
	repliersID := make([]int, 0)
	commentList := make([]*GetCommentListResp, len(comments))
	for i := range comments {
		commentList[i].Comment = comments[i]
		commentersID = append(commentersID, comments[i].UserID)
		if comments[i].ReplierID > 0 {
			repliersID = append(repliersID, comments[i].ReplierID)
		}
	}
	usersID = append(usersID, commentersID...)
	usersID = append(usersID, repliersID...)
	usersID = slicex.RemoveDuplicates(usersID)

	//评论的用户信息
	commenters := make([]*model.User, 0)
	err = mysql.Find(ctx, mysql.GetUserDB(), &commenters, map[string]any{
		"id": usersID,
	})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}

	for i := range commentList {
		for j := range commenters {
			if commentList[i].Comment.UserID == commenters[j].ID {
				commentList[i].UserInfo = commenters[j]
			}
			if commentList[i].Comment.ReplierID == commenters[j].ID {
				commentList[i].ReplierInfo = commenters[j]
			}
		}
	}

	ginx.ResponsePageSuccess(c, commentList, req.Page, req.PageSize, total)
}

func CreateComment(c *gin.Context) {
	ctx := c.Request.Context()
	userID := ginx.GetUserID(c)
	if userID <= 0 {
		ginx.ResponseParamError(c)
		return
	}

	req := model.MomentComment{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		ginx.ResponseParamError(c)
		return
	}
	if req.Content == "" {
		ginx.ResponseParamError(c)
		return
	}

	moment := model.UserMoment{}
	exist, err := mysql.Exist(ctx, mysql.GetUserDB(), &moment, map[string]any{"id": req.MomentID})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	if !exist {
		ginx.ResponseNotify(c, "动态已被删除")
		return
	}

	blacked, err := repo.IsBlacked(ctx, moment.UserID, userID)
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	if blacked {
		ginx.ResponseNotify(c, "对方已将你加入黑名单")
		return
	}

	req.UserID = userID

	fn := func(tx *gorm.DB) error {
		err = mysql.Create(ctx, tx, &req)
		if err != nil {
			return err
		}

		update := map[string]any{
			"comment_count": gorm.Expr("comment_count + ?", 1),
		}
		if moment.UserID == userID {
			update["self_comment_count"] = gorm.Expr("self_comment_count + ?", 1)
		}
		_, err = mysql.UpdateByID(ctx, tx, &moment, moment.ID, update)
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

	ginx.ResponseSuccess(c, req)
}

type DeleteMomentCommentReq struct {
	MomentID  int `json:"momentID" binding:"required"`
	CommentID int `json:"commentID" binding:"required"`
}

func DeleteComment(c *gin.Context) {
	ctx := c.Request.Context()
	userID := ginx.GetUserID(c)
	if userID <= 0 {
		ginx.ResponseParamError(c)
		return
	}

	req := DeleteMomentCommentReq{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		ginx.ResponseParamError(c)
		return
	}
	if req.MomentID <= 0 {
		ginx.ResponseParamError(c)
		return
	}
	if req.CommentID <= 0 {
		ginx.ResponseParamError(c)
		return
	}

	moment := model.UserMoment{}
	exist, err := mysql.Exist(ctx, mysql.GetUserDB(), &moment, map[string]any{"id": req.MomentID, "is_delete": 0})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	if !exist {
		ginx.ResponseNotify(c, "动态已被删除")
		return
	}

	comment := model.MomentComment{}
	exist, err = mysql.Exist(ctx, mysql.GetUserDB(), &comment, map[string]any{"moment_id": req.CommentID, "is_delete": 0})
	if err != nil {
		ginx.ResponseServerError(c, err)
		return
	}
	if !exist {
		ginx.ResponseSuccess(c, nil)
		return
	}

	if comment.UserID != userID {
		ginx.ResponseError(c, constant.CodePermissionDenied)
		return
	}

	fn := func(tx *gorm.DB) error {
		err = mysql.SoftDelete(ctx, tx, &comment, map[string]any{"id": req.CommentID, "is_delete": 0})
		if err != nil {
			return err
		}

		update := map[string]any{
			"comment_count": gorm.Expr("comment_count - ?", 1),
		}
		if moment.UserID == userID {
			update["self_comment_count"] = gorm.Expr("self_comment_count - ?", 1)
		}
		_, err = mysql.UpdateByID(ctx, tx, &moment, moment.ID, update)
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
