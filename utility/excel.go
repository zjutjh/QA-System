package utility

import (
	"fmt"
	"github.com/tealeg/xlsx"
)

func DataToExcel(titleList []string, dataList [][]string, fileName string) string {
	// 生成一个新的文件
	file := xlsx.NewFile()
	// 添加sheet页
	sheet, _ := file.AddSheet("Sheet1")
	// 插入表头
	titleRow := sheet.AddRow()
	for _, v := range titleList {
		cell := titleRow.AddCell()
		cell.Value = v
		cell.GetStyle().Font.Color = "00000000"
	}
	// 插入内容
	for _, v := range dataList {
		row := sheet.AddRow()
		row.WriteSlice(&v, -1)
	}
	fileName = fmt.Sprintf("%s.xlsx", fileName)
	_ = file.Save(fileName)
	return fileName
}
