package service

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"QA-System/internal/dao"
	"QA-System/internal/model"

	"github.com/xuri/excelize/v2"
	"github.com/zjutjh/WeJH-SDK/excel"
)

// HandleDownloadFile 处理下载文件
func HandleDownloadFile(answers dao.AnswersResonse, survey *model.Survey) (string, error) {
	questionAnswers := answers.QuestionAnswers
	times := answers.Time
	// 创建一个新的Excel文件
	f := excelize.NewFile()
	streamWriter, err := f.NewStreamWriter("Sheet1")
	if err != nil {
		return "", errors.New("创建Excel文件失败原因: " + err.Error())
	}
	// 设置字体样式
	styleID, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})
	if err != nil {
		return "", errors.New("设置字体样式失败原因: " + err.Error())
	}
	// 计算每列的最大宽度
	maxWidths := make(map[int]int)
	maxWidths[0] = 7
	maxWidths[1] = 20
	for i, qa := range questionAnswers {
		maxWidths[i+2] = len(qa.Title)
		for _, answer := range qa.Answers {
			if len(answer) > maxWidths[i+2] {
				maxWidths[i+2] = len(answer)
			}
		}
	}
	// 设置列宽
	for colIndex, width := range maxWidths {
		if width > 255 {
			width = 255
		}
		if err := streamWriter.SetColWidth(colIndex+1, colIndex+1, float64(width)); err != nil {
			return "", errors.New("设置列宽失败原因: " + err.Error())
		}
	}
	// 写入标题行
	rowData := make([]any, 0)
	rowData = append(rowData, excelize.Cell{Value: "序号", StyleID: styleID},
		excelize.Cell{Value: "提交时间", StyleID: styleID})
	for _, qa := range questionAnswers {
		rowData = append(rowData, excelize.Cell{Value: qa.Title, StyleID: styleID})
	}
	if err := streamWriter.SetRow("A1", rowData); err != nil {
		return "", errors.New("写入标题行失败原因: " + err.Error())
	}
	// 写入数据
	for i, t := range times {
		row := []any{i + 1, t}
		for j, qa := range questionAnswers {
			if len(qa.Answers) <= i {
				continue
			}
			answer := qa.Answers[i]
			row = append(row, answer)
			colName, err := excelize.ColumnNumberToName(j + 3)
			if err != nil {
				return "", errors.New("转换列名失败原因: " + err.Error())
			}
			if err := f.SetCellValue("Sheet1", colName+strconv.Itoa(i+2), answer); err != nil {
				return "", errors.New("写入数据失败原因: " + err.Error())
			}
		}
		if err := streamWriter.SetRow(fmt.Sprintf("A%d", i+2), row); err != nil {
			return "", errors.New("写入数据失败原因: " + err.Error())
		}
	}
	// 关闭
	if err := streamWriter.Flush(); err != nil {
		return "", errors.New("关闭失败原因: " + err.Error())
	}
	// 保存Excel文件
	fileName := survey.Title + ".xlsx"
	filePath := "./public/xlsx/" + fileName
	if _, err := os.Stat("./public/xlsx/"); os.IsNotExist(err) {
		err := os.Mkdir("./public/xlsx/", 0750)
		if err != nil {
			return "", errors.New("创建文件夹失败原因: " + err.Error())
		}
	}
	// 删除旧文件
	if _, err := os.Stat(filePath); err == nil {
		if err := os.Remove(filePath); err != nil {
			return "", errors.New("删除旧文件失败原因: " + err.Error())
		}
	}
	// 保存
	if err := f.SaveAs(filePath); err != nil {
		return "", errors.New("保存文件失败原因: " + err.Error())
	}

	urlHost := GetConfigUrl()
	url := urlHost + "/public/xlsx/" + fileName

	return url, nil
}

// HandleChooseStatistics 导出投票结果
func HandleChooseStatistics(survey *model.Survey, response []GetChooseStatisticsResponse) (string, error) {
	sheets := make([]excel.Sheet, 0, len(response))

	for _, stat := range response {
		sheetName := fmt.Sprintf("第%d题", stat.SerialNum)
		headers := []string{"选项内容", "票数", "百分比"}
		var rows [][]any

		for _, opt := range stat.Options {
			row := []any{opt.Content, opt.Count, opt.Percent}
			rows = append(rows, row)
		}

		sheet := excel.Sheet{
			Name:    sheetName,
			Headers: headers,
			Rows:    rows,
		}
		sheets = append(sheets, sheet)
	}

	fileData := excel.File{Sheets: sheets}
	fileName := survey.Title + ".xlsx"
	filePath := "./public/xlsx/"

	// 创建目录
	if err := os.MkdirAll(filePath, 0750); err != nil {
		return "", fmt.Errorf("创建目录失败: %v", err)
	}

	fullPath := filepath.Join(filePath, fileName)

	// 删除旧文件
	if _, err := os.Stat(fullPath); err == nil {
		if err := os.Remove(fullPath); err != nil {
			return "", fmt.Errorf("删除旧文件失败: %v", err)
		}
	}

	// 创建Excel文件
	if _, err := excel.CreateExcelFile(fileData, fileName, filePath); err != nil {
		return "", fmt.Errorf("创建Excel文件失败: %v", err)
	}

	urlHost := GetConfigUrl()
	url := urlHost + "/public/xlsx/" + fileName

	return url, nil
}
