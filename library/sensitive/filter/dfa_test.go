package filter

import (
	"fmt"
	"testing"

	"tongcheng/utils"
)

func TestDFA(t *testing.T) {
	model := NewDfaModel()
	model.AddWords("敏感词1", "敏感词2", "敏感词")
	utils.Spew(model)
	fmt.Println(model.FindAll("敏感词1敏感词2敏感词3"))
	model.DelWord("敏感词1")
	fmt.Println("result:", model.FindAll("敏感词1敏感词2敏感词3"))
	utils.Spew(model)
}
