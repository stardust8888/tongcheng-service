package sensitive

import (
	"fmt"
	"testing"
	"time"
)

func TestInit(t *testing.T) {
	begin := time.Now()
	err := Init()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(time.Since(begin).Milliseconds())

	text := "日军军妓彩虹剑东风传奇护花剑冷香谷云雨欲魔欲魔艳史"
	fmt.Println(GetSenFilter().FindAll(text))
	fmt.Println(GetSenFilter().FindOne(text))
	fmt.Println(GetSenFilter().IsSensitive(text))
	fmt.Println(GetSenFilter().Replace(text, '*'))
	fmt.Println(GetSenFilter().Remove(text))
}
