package repo

import "tongcheng/model"

type UserMoment struct {
	User       *model.User            `json:"user"`
	UserMoment *model.UserMoment      `json:"user_moment"`
	Like       int                    `json:"like"`
	Watch      int                    `json:"watch"`
	Comments   []*model.MomentComment `json:"comments"`
	Commenters map[int]*model.User    `json:"commenters"`
}
