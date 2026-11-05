package file

import (
	"context"
	"testing"

	"tongcheng/conf"
	"tongcheng/library/minio"
)

func TestFile(t *testing.T) {
	err := conf.InitConfig()
	if err != nil {
		t.Error(err)
		return
	}

	err = minio.Init()
	if err != nil {
		t.Error(err)
		return
	}

	err = minio.DeleteImage(context.Background(), "/moment/1/568993986325266432.png")
	if err != nil {
		t.Error(err)
		return
	}
}
