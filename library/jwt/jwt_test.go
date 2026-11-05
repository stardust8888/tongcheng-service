package jwt

import (
	"fmt"
	"testing"

	"tongcheng/conf"
	"tongcheng/utils/constant"
	"tongcheng/utils/x/timex"
)

func TestJWT(t *testing.T) {
	err := conf.InitConfig()
	if err != nil {
		fmt.Println(err)
		return
	}

	config := conf.GetConfiguration()
	jt, err := New(config.Secret.JwtSecret, timex.Week)
	if err != nil {
		fmt.Println(err)
		return
	}

	t.Run("print token", func(t *testing.T) {
		simple1 := constant.SimpleUser{ID: 1}
		token1, err := jt.GenerateToken(simple1)
		if err != nil {
			fmt.Println(err)
			return
		}
		simple2 := constant.SimpleUser{ID: 2}
		token2, err := jt.GenerateToken(simple2)
		if err != nil {
			fmt.Println(err)
			return
		}
		simple3 := constant.SimpleUser{ID: 3}
		token3, err := jt.GenerateToken(simple3)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(token1)
		fmt.Println(token2)
		fmt.Println(token3)
	})

}
