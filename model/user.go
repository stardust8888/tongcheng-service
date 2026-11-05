package model

import "tongcheng/utils/x/timex"

type User struct {
	GormModel
	Username    string `gorm:"column:username;size:64;comment:用户名;default:'';not null" json:"username"`
	Phone       string `gorm:"column:phone;size:64;comment:手机号;default:'';not null" json:"-"`
	PhoneSuffix string `gorm:"column:phone_suffix;size:32;comment:手机号;default:'';not null" json:"-"`
	Wx          string `gorm:"column:wx;size:64;comment:微信;default:'';not null" json:"wx"`
	WxOpenID    string `gorm:"column:wx_open_id;size:64;uniqueIndex;comment:微信openid;default:'';not null" json:"wxOpenID"`
	WxUnionID   string `gorm:"column:wx_union_id;size:64;comment:微信unionid;default:'';not null" json:"-"`
	SessionKey  string `gorm:"column:session_key;size:64;comment:微信sessionkey;default:'';not null" json:"-"`
	Avatar      string `gorm:"column:avatar;size:256;comment:头像;default:'';not null" json:"avatar"`
	Role        int    `gorm:"column:role;type:tinyint;comment:角色 0:未记录1:父亲2:母亲;default:0;not null" json:"role"`
	IsVip       int    `gorm:"column:is_vip;type:tinyint;comment:是否会员 0:否1:是;default:0;not null" json:"isVip"`
	Status      int    `gorm:"column:status;type:tinyint;comment:状态 0:正常1:禁用;default:0;not null" json:"status"`
	RegisterIP  string `gorm:"column:register_ip;size:64;comment:注册ip;default:'';not null" json:"registerIP"`
	Promoter    string `gorm:"column:promoter;size:64;comment:推广人;default:'';not null" json:"promoter"`
}

func (*User) TableName() string {
	return "user"
}

type UserChild struct {
	GormModel
	UserID        int    `gorm:"column:user_id;index;comment:关系用户ID;default:0;not null" json:"userID"`
	Relation      int    `gorm:"column:relation;type:tinyint;comment:关系 0:未记录1:儿子2:女儿;default:0;not null" json:"relation"`
	Height        int    `gorm:"column:height;type:int;comment:身高(cm);default:0;not null" json:"height"`
	Weight        int    `gorm:"column:weight;type:int;comment:体重(kg);default:0;not null" json:"weight"`
	Birth         string `gorm:"column:birth;size:32;comment:生日;default:'';:not null" json:"birth"`
	Education     int    `gorm:"column:education;type:tinyint;comment:学历 0:未记录1:高中2:大专3:本科4:硕士5:博士;default:0;not null" json:"education"`
	MaritalStatus int    `gorm:"column:marital_status;type:tinyint;comment:婚姻状况 0:未记录1:未婚2:离异3:丧偶;default:0;not null" json:"maritalStatus"`
	MaritalPlan   int    `gorm:"column:marital_plan;type:tinyint;comment:结婚时间 0:未记录1:尽快2:一年内3:两年内4:时机成熟;default:0;not null" json:"maritalPlan"`
	IsOnlyChild   int    `gorm:"column:is_only_child;type:tinyint;comment:独生子女 0:未记录1:不是2:是;default:0;not null" json:"isOnlyChild"`
	House         int    `gorm:"column:house;type:tinyint;comment:房产 0:未记录1:暂未购房2:一套3:两套及以上;default:0;not null" json:"house"`
	Car           int    `gorm:"column:car;type:tinyint;comment:车辆 0:未记录1:暂未购车2:一辆3:两辆及以上;default:0;not null" json:"car"`
	Company       int    `gorm:"column:company;type:tinyint;comment:公司;default:0;not null" json:"company"`
	Work          int    `gorm:"column:work;type:tinyint;comment:工作;default:0;not null" json:"work"`
	YearIncome    int    `gorm:"column:year_income;type:tinyint;comment:年收入 0:未记录1:10w以内2:10-20w3:20-30w4:30-50w5:50w以上;default:0;not null" json:"yearIncome"`
	Province      string `gorm:"column:province;size:32;comment:居住省份;default:'';not null" json:"province"`
	City          string `gorm:"column:city;size:32;comment:居住城市;default:'';not null" json:"city"`
	District      string `gorm:"column:district;size:32;comment:居住区域;default:'';not null" json:"district"`
	Photo         string `gorm:"column:photo;size:1024;comment:照片;default:'';not null" json:"photo"`
	Interest      string `gorm:"column:interest;size:256;comment:兴趣爱好;default:'';not null" json:"interest"`
	Description   string `gorm:"column:description;size:1024;comment:描述;default:'';not null" json:"description"`
	Status        int    `gorm:"column:status;type:tinyint;comment:状态 0:未展示1:展示;default:0;not null" json:"status"`
	ReviewStatus  int    `gorm:"column:review_status;type:tinyint;comment:状态 0:未审核1:通过2:拒绝;default:0;not null" json:"reviewStatus"`
	ReviewReason  string `gorm:"column:review_reason;size:1024;comment:审核原因;default:'';not null" json:"reviewReason"`
}

func (*UserChild) TableName() string {
	return "user_child"
}

type UserChildDisplay struct {
	GormModel
	UserID        int    `gorm:"column:user_id;index;comment:关系用户ID;default:0;not null" json:"userID"`
	Relation      int    `gorm:"column:relation;type:tinyint;comment:关系 0:未记录1:儿子2:女儿;default:0;not null" json:"relation"`
	Height        int    `gorm:"column:height;type:int;comment:身高(cm);default:0;not null" json:"height"`
	Weight        int    `gorm:"column:weight;type:int;comment:体重(kg);default:0;not null" json:"weight"`
	Birth         string `gorm:"column:birth;size:32;comment:生日;default:'';:not null" json:"birth"`
	Education     int    `gorm:"column:education;type:tinyint;comment:学历 0:未记录1:高中2:大专3:本科4:硕士5:博士;default:0;not null" json:"education"`
	MaritalStatus int    `gorm:"column:marital_status;type:tinyint;comment:婚姻状况 0:未记录1:未婚2:离异3:丧偶;default:0;not null" json:"maritalStatus"`
	MaritalPlan   int    `gorm:"column:marital_plan;type:tinyint;comment:结婚时间 0:未记录1:尽快2:一年内3:两年内4:时机成熟;default:0;not null" json:"maritalPlan"`
	IsOnlyChild   int    `gorm:"column:is_only_child;type:tinyint;comment:独生子女 0:未记录1:不是2:是;default:0;not null" json:"isOnlyChild"`
	House         int    `gorm:"column:house;type:tinyint;comment:房产 0:未记录1:暂未购房2:一套3:两套及以上;default:0;not null" json:"house"`
	Car           int    `gorm:"column:car;type:tinyint;comment:车辆 0:未记录1:暂未购车2:一辆3:两辆及以上;default:0;not null" json:"car"`
	Company       int    `gorm:"column:company;type:tinyint;comment:公司;default:0;not null" json:"company"`
	Work          int    `gorm:"column:work;type:tinyint;comment:工作;default:0;not null" json:"work"`
	YearIncome    int    `gorm:"column:year_income;type:tinyint;comment:年收入 0:未记录1:10w以内2:10-20w3:20-30w4:30-50w5:50w以上;default:0;not null" json:"yearIncome"`
	Province      string `gorm:"column:province;size:32;comment:居住省份;default:'';not null" json:"province"`
	City          string `gorm:"column:city;size:32;comment:居住城市;default:'';not null" json:"city"`
	District      string `gorm:"column:district;size:32;comment:居住区域;default:'';not null" json:"district"`
	Photo         string `gorm:"column:photo;size:1024;comment:照片;default:'';not null" json:"photo"`
	Interest      string `gorm:"column:interest;size:256;comment:兴趣爱好;default:'';not null" json:"interest"`
	Description   string `gorm:"column:description;size:1024;comment:描述;default:'';not null" json:"description"`
	Status        int    `gorm:"column:status;type:tinyint;comment:状态 0:未展示1:展示;default:0;not null" json:"status"`
}

func (*UserChildDisplay) TableName() string {
	return "user_child_display"
}

type UserWatch struct {
	GormModel
	UserID   int `gorm:"column:user_id;index;comment:用户ID;default:0;not null" json:"userID"`
	WatchID  int `gorm:"column:watch_id;index;comment:关注用户ID;default:0;not null" json:"watchID"`
	IsDelete int `gorm:"column:is_delete;type:tinyint;comment:是否删除 0:否1:是;default:0;not null" json:"isDelete"`
}

func (*UserWatch) TableName() string {
	return "user_watch"
}

type UserVisitRecord struct {
	GormModel
	UserID    int `gorm:"column:user_id;index;comment:用户ID;default:0;not null" json:"userID"`
	VisiterID int `gorm:"column:visiter_id;index;comment:访问者ID;default:0;not null" json:"visiterID"`
	Page      int `gorm:"column:page;type:tinyint;comment:访问了哪里 0:未知1:个人主页;default:0;not null" json:"page"`
}

func (*UserVisitRecord) TableName() string {
	return "user_visit_record"
}

type UserPayRecord struct {
	GormModel
	UserID int `gorm:"column:user_id;index;comment:用户ID;default:0;not null" json:"userID"`
	PayID  int `gorm:"column:pay_id;index;comment:支付ID;default:0;not null" json:"payID"`
	Amount int `gorm:"column:amount;comment:金额;default:0;not null" json:"amount"`
}

func (*UserPayRecord) TableName() string {
	return "user_pay_record"
}

type UserChat struct {
	GormModel
	UserID     int    `gorm:"column:user_id;index;comment:用户ID;default:0;not null" json:"userID"`
	ChatID     string `gorm:"column:chat_id;size:64;index;comment:聊天ID;default:'';not null" json:"chatID"`
	TalkerID   int    `gorm:"column:talker_id;index;comment:会话对象ID;default:0;not null" json:"talkerID"`
	IsDelete   int    `gorm:"column:is_delete;type:tinyint;comment:是否删除 0:否1:是;default:0;not null" json:"isDelete"`
	UserDelete int    `gorm:"column:user_delete;comment:删除会话用户id;default:0;not null" json:"userDelete"`
}

func (*UserChat) TableName() string {
	return "user_chat"
}

type UserBlackList struct {
	GormModel
	UserID      int `gorm:"column:user_id;index;comment:用户ID;default:0;not null" json:"userID"`
	BlackUserID int `gorm:"column:black_user_id;index;comment:拉黑对象ID;default:0;not null" json:"blackUserID"`
}

func (*UserBlackList) TableName() string {
	return "user_black_list"
}

type UserDeleteApply struct {
	GormModel
	UserID    int        `gorm:"column:user_id;index;comment:用户ID;default:0;not null" json:"userID"`
	ApplyAt   timex.Time `gorm:"column:apply_at;type:timestamp;comment:申请时间;not null" json:"applyAt"`
	ProcessAt timex.Time `gorm:"column:process_at;type:timestamp;comment:处理时间(到期删除或取消时间)" json:"processAt"`
	Status    int        `gorm:"column:status;type:tinyint;comment:状态 0:待处理 1:已处理 2:已取消;default:0;not null" json:"status"`
}

func (*UserDeleteApply) TableName() string {
	return "user_delete_apply"
}
