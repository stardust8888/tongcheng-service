package model

import (
	"gorm.io/gorm"

	"tongcheng/conf"
	"tongcheng/utils"
)

func MigrateAdminTables(db *gorm.DB) error {
	err := db.AutoMigrate(
		&Admin{},
		&Role{},
		&AdminRole{},
		&Menu{},
		&RoleMenu{},
		&AdminLoginLog{},
		&AdminOperationLog{},
	)
	if err != nil {
		return err
	}

	//admin
	err = db.FirstOrCreate(
		&Admin{GormModel: GormModel{ID: 1}, Username: "superadmin", Status: 1, IsSuper: 1, RealName: "张三丰",
			Password: utils.MD5("2664173ffc6c00744dfb3cd0e0295c6d" + conf.GetConfiguration().Secret.PwdSecret), //Qwer12356
		}).Error
	if err != nil {
		return err
	}

	return nil
}

func MigrateUserTables(db *gorm.DB) error {
	err := db.AutoMigrate(
		&User{},
		&UserChild{},
		&UserChildDisplay{},
		&UserChat{},
		&Message{},
		&UserMoment{},
		&MomentLike{},
		&MomentComment{},
		&UserWatch{},
		&UserVisitRecord{},
		&UserBlackList{},
		&UserDeleteApply{},
		&UserFeedback{},
	)
	if err != nil {
		return err
	}
	return nil
}
