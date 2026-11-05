package cron

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	"tongcheng/conf"
	"tongcheng/db/redis"
	"tongcheng/library/ginx"
	"tongcheng/library/redislock"
	"tongcheng/logs"
	"tongcheng/utils/x/timex"
)

type Cron struct {
	cron   *cron.Cron
	locker *redislock.Client
}

func (cc *Cron) AddTask(spec string, task func()) {
	_, _ = cc.cron.AddFunc(spec, task)
}

func (cc *Cron) AddTaskWithLocker(taskName, spec string, task func()) {
	fn := func() {
		locker, err := cc.locker.Obtain(ginx.Context, taskName, 60*time.Second)
		if err != nil {
			logs.Error(ginx.Context, "获取锁失败", zap.Error(err))
			return
		}

		killWatchDog := make(chan struct{}, 1)
		defer func() {
			_ = locker.Release(ginx.Context)
			killWatchDog <- struct{}{}
		}()

		go func() {
			ticker := time.NewTicker(20 * time.Second)
			defer func() {
				ticker.Stop()
			}()
			for {
				select {
				case <-ticker.C:
					err = locker.Refresh(ginx.Context, 60)
					if err != nil {
						logs.Error(ginx.Context, "刷新锁失败", zap.Error(err))
					}
				case <-killWatchDog:
					logs.Info(ginx.Context, "任务结束...")
					return
				}
			}
		}()

		task()
	}

	_, _ = cc.cron.AddFunc(spec, fn)
}

func (cc *Cron) Start() {
	cc.cron.Start()
}

func (cc *Cron) Stop() {
	cc.cron.Stop()
}

func InitCron() error {
	location, err := timex.GetBeijingLocation()
	if err != nil {
		logs.Error(ginx.Context, "获取北京时间时区失败", zap.Error(err))
		return err
	}

	cc := Cron{
		cron:   cron.New(cron.WithLocation(location)),
		locker: redislock.New(redis.GetRedisClient()),
	}

	cronConfig := conf.GetConfiguration().Cron
	cc.AddTask(cronConfig.SaveMessageCron, SaveMessage)
	cc.AddTaskWithLocker("SaveMessage", cronConfig.SaveMessageCron, SaveMessage)
	// 注销申请到期处理：每小时扫描一次，满15天的申请正式删除账号及资料
	cc.AddTaskWithLocker("DeleteUser", cronConfig.DeleteUserCron, DeleteUser)

	cc.Start()
	return nil
}

func HandleCronTask(c *gin.Context) {
	taskName := c.Query("taskName")
	if taskName == "" {
		ginx.ResponseParamError(c)
		return
	}

	switch taskName {
	case "SaveMessage":
		SaveMessage()
	case "DeleteUser":
		DeleteUser()
	default:
		ginx.ResponseParamError(c)
		return
	}

	ginx.ResponseSuccess(c, nil)
}
