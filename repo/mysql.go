package repo

import (
	"context"

	"tongcheng/db/mysql"
	"tongcheng/model"
	"tongcheng/utils/x/slicex"
)

func GetUsersByID(ctx context.Context, usersID []int) (map[int]*model.User, error) {
	users := make([]*model.User, 0)
	usersID = slicex.RemoveDuplicates(usersID)
	err := mysql.Find(ctx, mysql.GetUserDB(), &users, map[string]any{"id": usersID})
	if err != nil {
		return nil, err
	}
	result := make(map[int]*model.User, len(users))
	for i := range users {
		result[users[i].ID] = users[i]
	}
	return result, nil
}

func GetUserChildByID(ctx context.Context, usersID []int) (map[int]*model.UserChild, error) {
	usersChild := make([]*model.UserChild, 0)
	usersID = slicex.RemoveDuplicates(usersID)
	err := mysql.Find(ctx, mysql.GetUserDB(), &usersChild, map[string]any{"user_id": usersID})
	if err != nil {
		return nil, err
	}
	result := make(map[int]*model.UserChild, len(usersChild))
	for i := range usersChild {
		result[usersChild[i].UserID] = usersChild[i]
	}
	return result, nil
}

// GetUserChildDisplayByUserID 批量按 user_id 查子女展示快照(user_child_display)，
// 返回 userID→快照 映射；快照为审核通过时落库的"已发布"版本，供后台比对当前资料与展示版的差异
func GetUserChildDisplayByUserID(ctx context.Context, usersID []int) (map[int]*model.UserChildDisplay, error) {
	displays := make([]*model.UserChildDisplay, 0)
	usersID = slicex.RemoveDuplicates(usersID)
	err := mysql.Find(ctx, mysql.GetUserDB(), &displays, map[string]any{"user_id": usersID})
	if err != nil {
		return nil, err
	}
	result := make(map[int]*model.UserChildDisplay, len(displays))
	for i := range displays {
		result[displays[i].UserID] = displays[i]
	}
	return result, nil
}

func GetUserWatchID(ctx context.Context, userID int) ([]int, error) {
	userWatch := make([]*model.UserWatch, 0)
	err := mysql.Find(ctx, mysql.GetUserDB(), &userWatch, map[string]any{"user_id": userID, "is_delete": 0})
	if err != nil {
		return nil, err
	}

	watchID := make([]int, 0)
	for _, watch := range userWatch {
		watchID = append(watchID, watch.WatchID)
	}
	return watchID, nil
}

func GetUserLikeMomentID(ctx context.Context, userID int) ([]int, error) {
	momentLike := make([]*model.MomentLike, 0)
	err := mysql.Find(ctx, mysql.GetUserDB(), &momentLike, map[string]any{"user_id": userID, "is_delete": 0})
	if err != nil {
		return nil, err
	}

	momentID := make([]int, 0)
	for _, like := range momentLike {
		momentID = append(momentID, like.MomentID)
	}
	return momentID, nil
}

func GetUserMoments(ctx context.Context, userMoments []*model.UserMoment, userID int) ([]*UserMoment, error) {
	momentsID := make([]int, 0)
	for _, userMoment := range userMoments {
		momentsID = append(momentsID, userMoment.ID)
	}

	//关注
	watcher, err := GetUserWatchID(ctx, userID)
	if err != nil {
		return nil, err
	}
	//用户赞的动态
	likeMomentsID, err := GetUserLikeMomentID(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]*UserMoment, 0)
	for _, userMoment := range userMoments {
		watch, like := 0, 0
		if slicex.ContainInSlice(watcher, userMoment.UserID) {
			watch = 1
		}
		if slicex.ContainInSlice(likeMomentsID, userMoment.ID) {
			like = 1
		}
		result = append(result, &UserMoment{
			UserMoment: userMoment,
			Watch:      watch,
			Like:       like,
		})
	}

	return result, nil
}

func GetMomentComments(ctx context.Context, userMoments []*UserMoment, watcher int) error {
	momentsID := make([]int, 0)
	usersID := make([]int, 0)
	momentUser := make(map[int]int) //记动态属于哪个用户
	for _, userMoment := range userMoments {
		momentsID = append(momentsID, userMoment.UserMoment.ID)
		usersID = append(usersID, userMoment.UserMoment.UserID)
		momentUser[userMoment.UserMoment.ID] = userMoment.UserMoment.UserID
		userMoment.Commenters = make(map[int]*model.User)
	}
	comments := make([]*model.MomentComment, 0)
	err := mysql.Find(ctx, mysql.GetUserDB(), &comments, map[string]any{
		"moment_id": momentsID,
		"is_delete": 0,
		mysql.Order: "id asc",
	})
	if err != nil {
		return err
	}

	viewComments := make([]*model.MomentComment, 0)
	commentersID := make([]int, 0)
	for _, comment := range comments {
		//如果自己看自己的动态，全都能看
		if momentUser[comment.MomentID] == watcher {
			commentersID = append(commentersID, comment.UserID)
			if comment.ReplierID > 0 {
				commentersID = append(commentersID, comment.ReplierID)
			}
			viewComments = append(viewComments, comment)
		} else {
			//如果看别人的动态，不是发动态的用户的评论，发动态用户回复别人的评论不能看
			if momentUser[comment.MomentID] == comment.UserID && (comment.ReplierID == 0 || comment.ReplierID == watcher) { //发动态用户的评论,不包含回复别人的评论
				commentersID = append(commentersID, comment.UserID)
				if comment.ReplierID > 0 {
					commentersID = append(commentersID, comment.ReplierID)
				}
				viewComments = append(viewComments, comment)
			}
			if comment.UserID == watcher { //watcher评论的
				commentersID = append(commentersID, comment.UserID)
				viewComments = append(viewComments, comment)
			}
		}
	}

	momentCommentsMap := make(map[int][]*model.MomentComment)
	for _, comment := range viewComments {
		momentCommentsMap[comment.MomentID] = append(momentCommentsMap[comment.MomentID], comment)
	}

	usersID = append(usersID, commentersID...)
	usersMap, err := GetUsersByID(ctx, usersID)
	if err != nil {
		return err
	}

	for _, userMoment := range userMoments {
		if momentComments, ok := momentCommentsMap[userMoment.UserMoment.ID]; ok {
			userMoment.Comments = momentComments
			for _, comment := range momentComments {
				userMoment.Commenters[comment.UserID] = usersMap[comment.UserID]
				if comment.ReplierID > 0 {
					userMoment.Commenters[comment.ReplierID] = usersMap[comment.ReplierID]
				}
			}
		}
		if user, ok := usersMap[userMoment.UserMoment.UserID]; ok {
			userMoment.User = user
		}
	}

	return nil
}
