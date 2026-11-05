package mysql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"tongcheng/conf"
	"tongcheng/logs"
	"tongcheng/model"
	"tongcheng/utils/common"
)

var (
	userDB  *gorm.DB
	adminDB *gorm.DB
)

const (
	Offset = "_offset"
	Limit  = "_limit"
	Order  = "_order"
)

func SetPage(where map[string]any, page, pageSize int) {
	if page <= 0 {
		page = common.DefaultPage
	}
	if pageSize <= 0 {
		pageSize = common.DefaultPageSize
	}
	where[Offset] = (page - 1) * pageSize
	where[Limit] = pageSize
}

func TableAlias(table model.TableModel, alias string) string {
	return table.TableName() + " as " + alias
}

func Init(config conf.MysqlConfig) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Username,
		config.Password,
		config.Address,
		config.DBName)

	zapGormLogger := New(logs.Logger(), logger.Config{
		SlowThreshold: 3 * time.Second,
		LogLevel:      logger.Info,
	})

	open, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: zapGormLogger,
	})
	if err != nil {
		return fmt.Errorf("连接数据库失败: %v", err)
	}

	sqlDB, err := open.DB()
	if err != nil {
		return fmt.Errorf("获取数据库对象失败: %v", err)
	}

	sqlDB.SetMaxOpenConns(30)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	switch config.DBName {
	case conf.GetConfiguration().Mysql.DBName:
		userDB = open
		if config.Migrate != 0 {
			err = model.MigrateUserTables(userDB)
			if err != nil {
				logs.Fatal(context.Background(), "MigrateUserTables error: %v", zap.Error(err))
			}
		}
	case conf.GetConfiguration().MysqlAdmin.DBName:
		adminDB = open
		if config.Migrate != 0 {
			err = model.MigrateAdminTables(adminDB)
			if err != nil {
				logs.Fatal(context.Background(), "MigrateAdminTables error: %v", zap.Error(err))
			}
		}
	}

	return nil
}

func Close() {

}

func GetUserDB() *gorm.DB {
	return userDB
}

func GetAdminDB() *gorm.DB {
	return adminDB
}

// Create 创建记录
func Create(ctx context.Context, db *gorm.DB, model model.TableModel) error {
	result := db.WithContext(ctx).Create(model)
	return result.Error
}

func BatchCreate[T model.TableModel](ctx context.Context, db *gorm.DB, models []T) error {
	result := db.WithContext(ctx).CreateInBatches(models, len(models))
	return result.Error
}

// Save 保存记录（更新所有字段）
func Save(ctx context.Context, db *gorm.DB, model model.TableModel) error {
	result := db.WithContext(ctx).Save(model)
	return result.Error
}

func Update(ctx context.Context, db *gorm.DB, model model.TableModel) (int64, error) {
	result := db.WithContext(ctx).Table(model.TableName()).Updates(model)
	return result.RowsAffected, result.Error
}

// Updates 更新记录（根据条件）
func Updates(ctx context.Context, db *gorm.DB, model model.TableModel, where, update map[string]any) (int64, error) {
	result := db.WithContext(ctx).Table(model.TableName()).Where(where).Updates(update)
	return result.RowsAffected, result.Error
}

// UpdateByID 根据ID更新记录
func UpdateByID(ctx context.Context, db *gorm.DB, model model.TableModel, id int, updates map[string]any) (int64, error) {
	result := db.WithContext(ctx).Table(model.TableName()).Where("id = ?", id).Updates(updates)
	return result.RowsAffected, result.Error
}

// Delete 删除记录
func Delete(ctx context.Context, db *gorm.DB, model model.TableModel, where map[string]any) error {
	result := db.WithContext(ctx).Table(model.TableName()).Where(where).Delete(model)
	return result.Error
}

func SoftDelete(ctx context.Context, db *gorm.DB, model model.TableModel, where map[string]any) error {
	result := db.WithContext(ctx).Table(model.TableName()).Where(where).Update("is_delete", 1)
	return result.Error
}

// DeleteByID 根据ID删除记录
func DeleteByID(ctx context.Context, db *gorm.DB, model model.TableModel, id int) error {
	result := db.WithContext(ctx).Table(model.TableName()).Where("id = ?", id).Delete(model)
	return result.Error
}

func Exist(ctx context.Context, db *gorm.DB, model model.TableModel, where map[string]any) (bool, error) {
	tx := db.WithContext(ctx).Table(model.TableName())
	tx = convertCondition(tx, where)
	err := tx.Take(model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// FindByID 根据ID查询记录
func FindByID(ctx context.Context, db *gorm.DB, model model.TableModel, id int) error {
	if id <= 0 {
		return gorm.ErrRecordNotFound
	}

	err := db.WithContext(ctx).Table(model.TableName()).Where("id = ?", id).Take(model).Error
	if err != nil {
		return err
	}
	return nil
}

// FindOne 查询单条记录
func FindOne(ctx context.Context, db *gorm.DB, model model.TableModel, where map[string]any) error {
	tx := db.WithContext(ctx).Table(model.TableName())
	tx = convertCondition(tx, where)
	err := tx.Take(model).Error
	if err != nil {
		return err
	}

	return nil
}

// Find 查询多条记录
func Find(ctx context.Context, db *gorm.DB, model any, where map[string]any) error {
	tx := db.WithContext(ctx).Model(model)
	tx = convertCondition(tx, where)
	tx = convertOrder(tx, where)
	result := tx.Find(model)
	return result.Error
}

// Count 获取记录总数
func Count(ctx context.Context, db *gorm.DB, model model.TableModel, where map[string]any) (int, error) {
	var (
		count int64
	)

	tx := db.WithContext(ctx).Table(model.TableName())
	tx = convertCondition(tx, where)
	result := tx.Count(&count)
	return int(count), result.Error
}

// PageQuery 分页查询
func PageQuery(ctx context.Context, db *gorm.DB, model any, where map[string]any) (int, error) {
	var (
		total int64
		err   error
	)

	tx := db.WithContext(ctx).Model(model)
	tx = convertCondition(tx, where)
	err = tx.Count(&total).Error
	if err != nil {
		return 0, err
	}

	tx = convertPage(tx, where)
	tx = convertOrder(tx, where)
	err = tx.Find(model).Error
	if err != nil {
		return 0, err
	}
	return int(total), err
}

func JoinQuery(ctx context.Context, db *gorm.DB, table string, list any, sel string, where map[string]any, joins ...string) error {
	var (
		err error
	)

	tx := db.WithContext(ctx).Table(table).Select(sel)
	for i := range joins {
		tx = tx.Joins(joins[i])
	}
	tx = convertCondition(tx, where)
	tx = convertOrder(tx, where)
	err = tx.Find(list).Error
	if err != nil {
		return err
	}
	return err
}

func PageJoinQuery(ctx context.Context, db *gorm.DB, table string, list any, sel string, where map[string]any, joins ...string) (int, error) {
	var (
		total int64
		err   error
	)

	tx := db.WithContext(ctx).Table(table).Select(sel)
	for i := range joins {
		tx = tx.Joins(joins[i])
	}
	tx = convertCondition(tx, where)
	err = tx.Count(&total).Error
	if err != nil {
		return 0, err
	}

	tx = convertPage(tx, where)
	tx = convertOrder(tx, where)
	err = tx.Find(list).Error
	if err != nil {
		return 0, err
	}
	return int(total), err
}

func Aggregate(ctx context.Context, db *gorm.DB, model model.TableModel, sel string, where map[string]any) error {
	tx := db.WithContext(ctx).Table(model.TableName()).Select(sel)
	tx = convertCondition(tx, where)
	return tx.Take(model).Error
}

func convertCondition(tx *gorm.DB, where map[string]any) *gorm.DB {
	for condition := range where {
		if condition == Offset {
			continue
		}
		if condition == Limit {
			continue
		}
		if condition == Order {
			continue
		}
		tx = tx.Where(condition, where[condition])
	}

	return tx
}

func convertPage(tx *gorm.DB, where map[string]any) *gorm.DB {
	for condition := range where {
		if condition == Offset {
			offset, ok := where[Offset].(int)
			if !ok {
				offset = common.DefaultOffset
			}
			tx = tx.Offset(offset)
			continue
		}
		if condition == Limit {
			limit, ok := where[Limit].(int)
			if !ok {
				limit = common.DefaultPageSize
			}
			tx = tx.Limit(limit)
			continue
		}
	}

	return tx
}

func convertOrder(tx *gorm.DB, where map[string]any) *gorm.DB {
	for condition := range where {
		if condition == Order {
			tx = tx.Order(where[condition])
			delete(where, condition)
			continue
		}
	}

	return tx
}

func Transaction(ctx context.Context, db *gorm.DB, fn func(tx *gorm.DB) error) error {
	var (
		tx  = db.WithContext(ctx).Begin()
		err error
	)
	defer func() {
		if r := recover(); r != nil {
			logs.Panic(ctx, "Transaction panic: %v", zap.Error(err))
			tx.Rollback()
		}
		if err != nil {
			logs.Error(ctx, "Transaction error: %v", zap.Error(err))
			tx.Rollback()
		}
	}()

	err = fn(tx)
	if err != nil {
		return err
	}

	err = tx.Commit().Error
	if err != nil {
		return err
	}

	return nil
}

func RawScan(ctx context.Context, db *gorm.DB, model model.TableModel, sql string, args ...any) error {
	return db.WithContext(ctx).Raw(sql, args...).Scan(model).Error
}

func Exec(ctx context.Context, db *gorm.DB, sql string, args ...any) error {
	return db.WithContext(ctx).Exec(sql, args...).Error
}

const (
	Like      = 0
	LeftLike  = 1
	RightLike = 2
)

func LikeQuery(kind int, value any) string {
	switch kind {
	case LeftLike:
		return "%" + fmt.Sprint(value)
	case RightLike:
		return fmt.Sprint(value) + "%"
	default:
		return "%" + fmt.Sprint(value) + "%"
	}
}
