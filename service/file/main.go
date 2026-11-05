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
	"tongcheng/library/httpclient/wxapi"
	"tongcheng/library/minio"
	"tongcheng/library/snowflake"
	"tongcheng/logs"
	"tongcheng/service/file/router"
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

	logs.InitLogger("file.log")
	defer logs.Sync()
	logs.Info(ctx, "logs init", zap.String("config", utils.ToJson(conf.GetConfiguration())))

	if err = minio.Init(); err != nil {
		logs.Panic(ctx, "MinIO初始化失败", zap.Error(err))
		panic("MinIO初始化失败:" + err.Error())
	}

	snowflake.InitSnowflake(4)

	// 内容安全检测依赖微信 access_token，需在 file 服务初始化 wxapi 客户端
	wxapi.InitClient()

	engine := gin.New()
	router.RegisterRoutes(engine)

	server := &http.Server{
		Addr:    ":12004",
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

	time.Sleep(10 * time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = server.Shutdown(ctx); err != nil {
		logs.Error(ctx, "服务器优雅关闭失败", zap.Error(err))
	} else {
		logs.Info(ctx, "服务器已安全关闭")
	}

	logs.Info(ctx, "服务已退出")
}
