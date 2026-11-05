package model

type UserMoment struct {
	GormModel
	UserID           int    `gorm:"column:user_id;index;comment:用户ID;default:0;not null" json:"userID"`
	LikeCount        int    `gorm:"column:like_count;comment:点赞数量;default:0;not null" json:"likeCount"`
	SelfLikeCount    int    `gorm:"column:self_like_count;comment:自己点赞数量;default:0;not null" json:"selfLikeCount"`
	CommentCount     int    `gorm:"column:comment_count;comment:评论数量;default:0;not null" json:"commentCount"`
	SelfCommentCount int    `gorm:"column:self_comment_count;comment:自己评论数量;default:0;not null" json:"selfCommentCount"`
	Content          string `gorm:"column:content;size:1024;comment:内容;default:'';not null" json:"content"`
	Picture          string `gorm:"column:picture;size:1024;comment:图片;default:'';not null" json:"picture"`
	IsDelete         int    `gorm:"column:is_delete;type:tinyint;comment:是否删除 0:否1:是;default:0;not null" json:"isDelete"`
	Status           int    `gorm:"column:status;type:tinyint;comment:状态 0:未展示1:展示;default:0;not null" json:"status"`
	ReviewStatus     int    `gorm:"column:review_status;type:tinyint;comment:状态 0:未审核1:通过2:拒绝;default:0;not null" json:"reviewStatus"`
	ReviewReason     string `gorm:"column:review_reason;size:1024;comment:审核原因;default:'';not null" json:"reviewReason"`
}

func (*UserMoment) TableName() string {
	return "user_moment"
}

type MomentLike struct {
	GormModel
	MomentID int `gorm:"column:moment_id;index;comment:动态ID;default:0;not null" json:"momentID"`
	UserID   int `gorm:"column:user_id;index;comment:用户ID;default:0;not null" json:"userID"`
	IsDelete int `gorm:"column:is_delete;type:tinyint;comment:是否删除 0:否1:是;default:0;not null" json:"isDelete"`
}

func (*MomentLike) TableName() string {
	return "moment_like"
}

type MomentComment struct {
	GormModel
	MomentID  int    `gorm:"column:moment_id;index;comment:动态ID;default:0;not null" json:"momentID"`
	UserID    int    `gorm:"column:user_id;index;comment:用户ID;default:0;not null" json:"userID"`
	ReplierID int    `gorm:"column:replier_id;index;comment:回复人ID;default:0;not null" json:"replierID"`
	Content   string `gorm:"column:content;size:1024;comment:内容;default:'';not null" json:"content"`
	IsDelete  int    `gorm:"column:is_delete;type:tinyint;comment:是否删除 0:否1:是;default:0;not null" json:"isDelete"`
}

func (*MomentComment) TableName() string {
	return "moment_comment"
}
