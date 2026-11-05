package redislock

import (
	"context"
	"fmt"
	"testing"
	"time"

	"tongcheng/conf"
	"tongcheng/db/redis"
)

func TestRedisLock(t *testing.T) {
	err := conf.InitConfig()
	if err != nil {
		panic("初始化配置文件失败:" + err.Error())
	}

	err = redis.Init()
	if err != nil {
		panic("初始化Redis失败:" + err.Error())
	}

	redisLock := New(redis.GetRedisClient())
	fmt.Println("开始获取锁...")
	locker, err := redisLock.Obtain(context.Background(), "test", 20*time.Second)
	if err != nil {
		fmt.Println("获取锁失败:" + err.Error())
		return
	}

	time.Sleep(10 * time.Second)

	defer func() {
		err = locker.Release(context.Background())
		if err != nil {
			fmt.Println("释放锁失败:" + err.Error())
		}
	}()
}
