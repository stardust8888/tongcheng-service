package repo

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"tongcheng/db/mysql"
	"tongcheng/logs"
	"tongcheng/model"
	"tongcheng/utils/x/timex"
)

// DeleteApplyDelay 注销冷静期：申请提交后 15 天正式处理
const DeleteApplyDelay = 15 * timex.Day

// processUserDeleteOne 处理单条到期的注销申请：事务内硬删除该用户及其关联数据。
func processUserDeleteOne(ctx context.Context, apply *model.UserDeleteApply) error {
	uid := apply.UserID
	return mysql.Transaction(ctx, mysql.GetUserDB(), func(tx *gorm.DB) error {
		// 子女资料（含 PII）
		if err := tx.Where("user_id = ?", uid).Delete(&model.UserChild{}).Error; err != nil {
			return err
		}
		// 动态及互动
		if err := tx.Where("user_id = ?", uid).Delete(&model.UserMoment{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", uid).Delete(&model.MomentLike{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", uid).Delete(&model.MomentComment{}).Error; err != nil {
			return err
		}
		// 关注 / 拉黑 / 访问 / 会话（含作为对方一侧的记录）
		if err := tx.Where("user_id = ? OR watch_id = ?", uid, uid).Delete(&model.UserWatch{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ? OR black_user_id = ?", uid, uid).Delete(&model.UserBlackList{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ? OR visiter_id = ?", uid, uid).Delete(&model.UserVisitRecord{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ? OR talker_id = ?", uid, uid).Delete(&model.UserChat{}).Error; err != nil {
			return err
		}
		// 支付记录
		if err := tx.Where("user_id = ?", uid).Delete(&model.UserPayRecord{}).Error; err != nil {
			return err
		}
		// 用户本体
		if err := tx.Where("id = ?", uid).Delete(&model.User{}).Error; err != nil {
			return err
		}
		// 标记申请已处理
		return tx.Model(&model.UserDeleteApply{}).Where("id = ?", apply.ID).
			Updates(map[string]any{"status": 1, "process_at": timex.GetBeijingTime()}).Error
	})
}

// ProcessUserDeleteApplies 扫描到期（申请满 15 天）的注销申请并处理。幂等，可重复调用。
// 由 service/cron 定时任务调度，并经 redis 锁保证多实例下不重复处理。
func ProcessUserDeleteApplies(ctx context.Context) {
	due := timex.GetBeijingTime().Add(-DeleteApplyDelay)
	applies := make([]*model.UserDeleteApply, 0)
	err := mysql.Find(ctx, mysql.GetUserDB(), &applies, map[string]any{
		"status":        0,
		"apply_at <= ?": due,
	})
	if err != nil {
		logs.Error(ctx, "ProcessUserDeleteApplies find err", zap.Error(err))
		return
	}
	for _, apply := range applies {
		if err := processUserDeleteOne(ctx, apply); err != nil {
			logs.Error(ctx, "ProcessUserDeleteApplies process err", zap.Int("userID", apply.UserID), zap.Error(err))
			continue
		}
		logs.Info(ctx, "ProcessUserDeleteApplies done", zap.Int("userID", apply.UserID))
	}
}
