package model

// Admin 管理员表
type Admin struct {
	GormModel
	Username   string `gorm:"column:username;size:32;uniqueIndex;comment:用户名;default:'';not null" json:"username"`
	Password   string `gorm:"column:password;size:256;comment:密码;default:'';not null" json:"password"`
	Email      string `gorm:"column:email;size:64;index;comment:邮箱;default:'';not null" json:"email"`
	Phone      string `gorm:"column:phone;size:32;index;comment:手机号;default:'';not null" json:"phone"`
	RealName   string `gorm:"column:real_name;size:16;comment:真实姓名;default:'';not null" json:"name"`
	Avatar     string `gorm:"column:avatar;size:256;comment:头像;default:'';not null" json:"avatar"`
	IsSuper    int    `gorm:"column:is_super;type:tinyint;comment:是否超管 0:否1:是 ;default:0;not null" json:"isSuper"`
	Status     int    `gorm:"column:status;type:tinyint;comment:状态 0:禁用1:正常;default:0;not null" json:"status"`
	GoogleAuth string `gorm:"column:google_auth;size:256;comment:谷歌验证码;default:'';not null" json:"googleAuth"`
}

// TableName 自定义表名
func (*Admin) TableName() string {
	return "admin"
}

// AdminRole 管理员角色关联表
type AdminRole struct {
	GormModel
	AdminID int `gorm:"column:admin_id;index;comment:管理员ID;default:0;not null" json:"adminID"`
	RoleID  int `gorm:"column:role_id;index;comment:角色ID;default:0;not null" json:"roleID"`
}

// TableName 自定义表名
func (*AdminRole) TableName() string {
	return "admin_role"
}

// Role 角色表
type Role struct {
	GormModel
	Name        string `gorm:"column:name;size:32;comment:角色名称;default:'';not null" json:"name"`
	Description string `gorm:"column:description;size:256;comment:描述;default:'';not null" json:"description"`
}

// TableName 自定义表名
func (*Role) TableName() string {
	return "role"
}

// RoleMenu 角色菜单关联表
type RoleMenu struct {
	GormModel
	RoleID int `gorm:"column:role_id;index;comment:角色ID;default:0;not null" json:"roleID"`
	MenuID int `gorm:"column:menu_id;index;comment:菜单ID;default:0;not null" json:"menuID"`
}

// TableName 自定义表名
func (*RoleMenu) TableName() string {
	return "role_menu"
}

// Menu 菜单表
type Menu struct {
	GormModel
	Label     string `gorm:"column:label;size:32;comment:菜单标签;default:'';not null" json:"label"`
	LabelEn   string `gorm:"column:label_en;size:32;comment:菜单标签英文;default:'';not null" json:"labelEn"`
	Icon      string `gorm:"column:icon;size:32;comment:图标;default:'';not null" json:"icon"`
	ParentID  int    `gorm:"column:parent_id;comment:父级ID;default:0;not null" json:"parentID"`
	Router    string `gorm:"column:router;size:256;comment:菜单URL;default:'';not null" json:"router"`
	Key       string `gorm:"column:key;size:256;comment:菜单key;default:'';not null" json:"key"`
	Rule      string `gorm:"column:rule;size:256;comment:菜单权限;default:'';not null" json:"rule"`
	Type      int    `gorm:"column:type;type:tinyint;comment:类型 1:目录 2:菜单 3:按钮;default:1;not null" json:"type"`
	IsVisible int    `gorm:"column:is_visible;type:tinyint;comment:是否可见 0:否1:是;default:0;not null" json:"state"`
	Sort      int    `gorm:"column:sort;comment:排序;default:0;not null" json:"order"`
	Remark    string `gorm:"column:remark;size:256;comment:备注;default:'';not null" json:"remark"`
}

// TableName 自定义表名
func (*Menu) TableName() string {
	return "menu"
}
