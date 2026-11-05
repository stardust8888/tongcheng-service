package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"

	"tongcheng/conf"
	"tongcheng/db/mysql"
	"tongcheng/db/redis"
	"tongcheng/library/snowflake"
	"tongcheng/logs"
	"tongcheng/service/cron/controller/cron"
	"tongcheng/service/cron/router"
	"tongcheng/utils"
)

func main() {
	var (
		ctx = context.WithValue(context.Background(), logs.RequestIDKey, "start...")
		err error
	)

	if err = conf.InitConfig(); err != nil {
		panic("初始化配置文件失败:" + err.Error())
	}

	logs.InitLogger("cron.log")
	defer logs.Sync()
	logs.Info(ctx, "logs init", zap.String("config", utils.ToJson(conf.GetConfiguration())))

	if err = mysql.Init(conf.GetConfiguration().Mysql); err != nil {
		logs.Panic(ctx, "Mysql初始化失败", zap.Error(err))
		panic("Mysql初始化失败:" + err.Error())
	}

	if err = redis.Init(); err != nil {
		logs.Panic(ctx, "Redis初始化失败", zap.Error(err))
		panic("Redis初始化失败:" + err.Error())
	}

	if err = cron.InitCron(); err != nil {
		logs.Panic(ctx, "Cron初始化失败", zap.Error(err))
		panic("Cron初始化失败:" + err.Error())
	}

	snowflake.InitSnowflake(5)

	engine := gin.New()
	router.RegisterRoutes(engine)

	server := &http.Server{
		Addr:    ":12005",
		Handler: engine,
	}

	go func() {
		if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logs.Fatal(ctx, "HTTP服务器启动失败", zap.Error(err))
		}
	}()
	logs.Info(ctx, "服务器已启动...")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logs.Info(ctx, "正在优雅关闭服务器...")

	time.Sleep(3 * time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = server.Shutdown(ctx); err != nil {
		logs.Error(ctx, "服务器优雅关闭失败", zap.Error(err))
	} else {
		logs.Info(ctx, "服务器已安全关闭")
	}

	logs.Info(ctx, "服务已退出")
}
