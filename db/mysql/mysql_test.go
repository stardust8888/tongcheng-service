package mysql

import (
	"context"
	"fmt"
	"testing"

	"tongcheng/conf"
	"tongcheng/logs"
	"tongcheng/model"
	"tongcheng/utils"
)

func TestFind(t *testing.T) {
	err := conf.InitConfig()
	if err != nil {
		t.Error(err)
		return
	}
	logs.InitLogger("test.log")

	err = Init(conf.GetConfiguration().Mysql)
	if err != nil {
		t.Error(err)
		return
	}

	//message := model.Message{}
	list := make([]*model.Message, 0)
	where := map[string]any{}
	where["_order"] = "created_at desc"
	where["msg_id"] = 502294222716534784
	where["msg_id = ?"] = 502294222716534784
	where["content like ?"] = "%高兴%"
	SetPage(where, 1, 3)
	_, err = PageQuery(context.Background(), userDB, &list, where)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(utils.ToJson(list))
}
