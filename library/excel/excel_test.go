package excel

import (
	"fmt"
	"testing"

	"tongcheng/utils"
)

func TestGetCellName(t *testing.T) {
	fmt.Println(getCellName(0, 0))
}

type User struct {
	Name string
	Age  int
}

func TestHandleData(t *testing.T) {
	data := []User{
		{"John", 18},
		{"Mike", 20},
		{"Jane", 16},
	}
	result := handleData(data, []string{"Name", "Age"})
	fmt.Println(utils.ToJson(result))
}

func TestExportReport_ExportReportToExcel(t *testing.T) {
	report := &ExportReport[User]{
		Export:    false,
		FileName:  "test.xlsx",
		SheetName: "Sheet1",
		Headers:   []string{"Name", "Age"},
		Fields:    []string{"Age", "Name"},
		Data: []User{
			{"John", 18},
			{"Mike", 20},
			{"Jane", 16},
		},
	}
	report.ExportReportToExcel(nil)
}
