package cron

import (
	"context"

	"tongcheng/repo"
)

// DeleteUser 注销账号处理：扫描满15天的申请并删除账号及关联资料。
// 由 InitCron 注册为带 redis 锁的定时任务，多实例下仅一个执行。
func DeleteUser() {
	repo.ProcessUserDeleteApplies(context.Background())
}
