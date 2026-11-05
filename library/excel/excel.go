package excel

import (
	"fmt"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"

	"tongcheng/utils/x/timex"
)

type ExportReport[T any] struct {
	Export    bool
	FileName  string
	SheetName string
	Headers   []string
	Fields    []string
	Data      []T
	FdMap     map[string]map[any]any
}

func (report *ExportReport[T]) ExportReportToExcel(c *gin.Context) {
	if len(report.Headers) == 0 {
		return
	}

	f := excelize.NewFile()

	if report.SheetName == "" {
		report.SheetName = "Sheet1"
	}
	_ = f.SetSheetName("Sheet1", report.SheetName)

	if len(report.Headers) > 0 {
		for j, header := range report.Headers {
			cell := getCellName(j, 0)
			_ = f.SetCellValue(report.SheetName, cell, header)
		}

		// 设置表头样式
		lastCol := getCellName(len(report.Headers)-1, 0)
		headerStyle, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{
				Bold: true,
			},
			Fill: excelize.Fill{
				Type:    "pattern",
				Color:   []string{"#F2F2F2"},
				Pattern: 1,
			},
			Alignment: &excelize.Alignment{
				Horizontal: "left",
				Vertical:   "center",
			},
		})
		_ = f.SetCellStyle(report.SheetName, "A1", lastCol, headerStyle)
		_ = f.SetColWidth(report.SheetName, "A1", lastCol, 30)

	}

	if len(report.Data) > 0 {
		rows := handleData(report.Data, report.Fields)
		for i, row := range rows {
			for j, cellValue := range row {
				cell := getCellName(j, i+1)
				_ = f.SetCellValue(report.SheetName, cell, cellValue)
			}
		}
		dataStyle, _ := f.NewStyle(&excelize.Style{
			Alignment: &excelize.Alignment{
				Horizontal: "left",
				Vertical:   "center",
			},
		})
		_ = f.SetCellStyle(report.SheetName, "A2", getCellName(len(rows[0])-1, len(report.Data)), dataStyle)
	}

	if report.Export {
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Header("Content-Disposition", fmt.Sprintf("attachment;filename=%s_", report.FileName)+time.Now().Format(timex.DateFormatCompact)+".xlsx")
		_ = f.Write(c.Writer)
	} else {
		_ = f.SaveAs(report.FileName)
	}
}

func handleData[T any](data []T, fields []string) [][]any {
	result := make([][]any, len(data))

	for i, item := range data {
		row := make([]any, len(fields))
		itemValue := reflect.ValueOf(item)

		if itemValue.Kind() == reflect.Ptr {
			itemValue = itemValue.Elem()
		}

		for j, field := range fields {
			fieldValue := itemValue.FieldByName(field)
			if fieldValue.IsValid() {
				row[j] = fieldValue.Interface()
			} else {
				row[j] = nil
			}
		}

		result[i] = row
	}

	return result
}

func getCellName(col, row int) string {
	colName := ""
	for col >= 0 {
		colName = string(rune('A'+col%26)) + colName
		col = col/26 - 1
		if col < 0 {
			break
		}
	}

	return fmt.Sprintf("%s%d", colName, row+1)
}
