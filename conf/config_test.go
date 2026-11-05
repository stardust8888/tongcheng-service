package conf

import (
	"fmt"
	"testing"

	"tongcheng/utils"
)

func TestInitConfig(t *testing.T) {
	err := InitConfig()
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(utils.ToJson(config))
}
