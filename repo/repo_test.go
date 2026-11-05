package repo

import (
	"context"
	"fmt"
	"testing"

	"tongcheng/conf"
	"tongcheng/logs"
)

func TestEncryptUserInfo(t *testing.T) {
	if err := conf.InitConfig(); err != nil {
		panic("初始化配置文件失败:" + err.Error())
	}
	logs.InitLogger("test.log")

	fmt.Println(EncryptUserInfo(context.Background(), 1, "wechat001"))
}

func TestDecryptUserInfo(t *testing.T) {
	if err := conf.InitConfig(); err != nil {
	}
	logs.InitLogger("test.log")
	fmt.Println(DecryptUserInfo(context.Background(), 1, "3879a8e59a0edd56ee6bbeef75549a6d336c5f287fb14d0b3d33267bf65cfa91"))
}
