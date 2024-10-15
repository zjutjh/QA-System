package service

import (
	"QA-System/internal/dao"
	"QA-System/internal/models"
	"QA-System/internal/pkg/log"
	"QA-System/internal/pkg/utils"
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

func GetAdminByUsername(username string) (*models.User, error) {
	user, err := d.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if user.Password != "" {
		aesDecryptPassword(user)
	}
	return user, nil
}

func GetAdminByID(id int) (*models.User, error) {
	user, err := d.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user.Password != "" {
		aesDecryptPassword(user)
	}
	return user, nil
}

func IsAdminExist(username string) error {
	_, err := d.GetUserByUsername(ctx, username)
	return err
}

func CreateAdmin(user models.User) error {
	aesEncryptPassword(&user)
	err := d.CreateUser(ctx, &user)
	return err
}

func GetUserByName(username string) (*models.User, error) {
	user, err := d.GetUserByUsername(ctx, username)
	return user, err
}

func CreatePermission(id int, surveyID int) error {
	err := d.CreateManage(ctx, id, surveyID)
	return err
}

func DeletePermission(id int, surveyID int) error {
	err := d.DeleteManage(ctx, id, surveyID)
	return err
}

func CheckPermission(id int, surveyID int) error {
	err := d.CheckManage(ctx, id, surveyID)
	return err
}

func CreateSurvey(id int, title string, desc string, img string, questions []dao.Question, status int, time time.Time) error {
	var survey models.Survey
	survey.UserID = id
	survey.Title = title
	survey.Desc = desc
	survey.Img = img
	survey.Status = status
	survey.Deadline = time
	survey, err := d.CreateSurvey(ctx, survey)
	if err != nil {
		return err
	}
	_, err = createQuestionsAndOptions(questions, survey.ID)
	return err
}

func UpdateSurveyStatus(id int, status int) error {
	err := d.UpdateSurveyStatus(ctx, id, status)
	return err
}

func UpdateSurvey(id int, title string, desc string, img string, questions []dao.Question, time time.Time) error {
	//遍历原有问题，删除对应选项
	var oldQuestions []models.Question
	var old_imgs []string
	new_imgs := make([]string, 0)
	//获取原有图片
	oldQuestions, err := d.GetQuestionsBySurveyID(ctx, id)
	fmt.Println(id)
	if err != nil {
		return err
	}
	old_imgs, err = getOldImgs(id, oldQuestions)
	if err != nil {
		return err
	}
	//删除原有问题和选项
	for _, oldQuestion := range oldQuestions {
		oldOptions, err := d.GetOptionsByQuestionID(ctx, oldQuestion.ID)
		if err != nil {
			return err
		}
		for _, oldOption := range oldOptions {
			err = d.DeleteOption(ctx, oldOption.ID)
			fmt.Println(oldOption.ID)
			if err != nil {
				return err
			}
		}
		err = d.DeleteQuestion(ctx, oldQuestion.ID)
		if err != nil {
			return err
		}
	}
	//修改问卷信息
	err = d.UpdateSurvey(ctx, id, title, desc, img, time)
	if err != nil {
		return err
	}
	new_imgs = append(new_imgs, img)
	//重新添加问题和选项
	imgs, err := createQuestionsAndOptions(questions, id)
	if err != nil {
		return err
	}
	new_imgs = append(new_imgs, imgs...)
	urlHost := GetConfigUrl()
	//删除无用图片
	for _, old_img := range old_imgs {
		if !contains(new_imgs, old_img) {
			_ = os.Remove("./public/static/" + strings.TrimPrefix(old_img, urlHost+"/public/static/"))
		}
	}
	return nil
}

func UpdateSurveyPart(id int, title string, desc string, img string,  time time.Time) error {
	return d.UpdateSurvey(ctx, id, title, desc, img, time)
}

func UserInManage(uid int, sid int) bool {
	_, err := d.GetManageByUIDAndSID(ctx, uid, sid)
	return err == nil
}

func DeleteSurvey(id int) error {
	var questions []models.Question
	questions, err := d.GetQuestionsBySurveyID(ctx, id)
	if err != nil {
		return err
	}
	var answerSheets []dao.AnswerSheet
	answerSheets, _, err = d.GetAnswerSheetBySurveyID(ctx, id, 0, 0, "", false)
	if err != nil {
		return err
	}
	//删除图片
	imgs, err := getDelImgs(id, questions, answerSheets)
	if err != nil {
		return err
	}
	//删除文件
	files, err := getDelFiles(answerSheets)
	if err != nil {
		return err
	}
	urlHost := GetConfigUrl()
	for _, img := range imgs {
		_ = os.Remove("./public/static/" + strings.TrimPrefix(img, urlHost+"/public/static/"))
	}
	for _, file := range files {
		_ = os.Remove("./public/file/" + strings.TrimPrefix(file, urlHost+"/public/file/"))
	}
	//删除答卷
	err = DeleteAnswerSheetBySurveyID(id)
	if err != nil {
		return err
	}
	//删除问题、选项、问卷、管理
	for _, question := range questions {
		err = d.DeleteOption(ctx, question.ID)
		if err != nil {
			return err
		}
	}
	err = d.DeleteQuestionBySurveyID(ctx, id)
	if err != nil {
		return err
	}
	err = d.DeleteSurvey(ctx, id)
	if err != nil {
		return err
	}
	err = d.DeleteManageBySurveyID(ctx, id)
	return err
}

func GetSurveyAnswers(id int, num int, size int, text string, unique bool) (dao.AnswersResonse, *int64, error) {
	var answerSheets []dao.AnswerSheet
	data := make([]dao.QuestionAnswers, 0)
	time := make([]string, 0)
	var total *int64
	//获取问题
	questions, err := d.GetQuestionsBySurveyID(ctx, id)
	if err != nil {
		return dao.AnswersResonse{}, nil, err
	}
	//初始化data
	for _, question := range questions {
		var q dao.QuestionAnswers
		q.Title = question.Subject
		q.QuestionType = question.QuestionType
		q.Answers = make([]string, 0)
		data = append(data, q)
	}
	//获取答卷
	answerSheets, total, err = d.GetAnswerSheetBySurveyID(ctx, id, num, size, text, unique)
	if err != nil {
		return dao.AnswersResonse{}, nil, err
	}
	//填充data
	for _, answerSheet := range answerSheets {
		time = append(time, answerSheet.Time)
		for _, answer := range answerSheet.Answers {
			question, err := d.GetQuestionByID(ctx, answer.QuestionID)
			if err != nil {
				return dao.AnswersResonse{}, nil, err
			}
			for i, q := range data {
				if q.Title == question.Subject {
					data[i].Answers = append(data[i].Answers, answer.Content)
				}
			}
		}
	}
	return dao.AnswersResonse{QuestionAnswers: data, Time: time}, total, nil
}

func GetAllSurveyByUserID(userId int) ([]models.Survey, error) {
	return d.GetAllSurveyByUserID(ctx, userId)
}

func ProcessResponse(response []interface{}, pageNum, pageSize int, title string) ([]interface{}, *int64) {
	if title != "" {
		filteredResponse := make([]interface{}, 0)
		for _, item := range response {
			itemMap := item.(map[string]interface{})
			if strings.Contains(strings.ToLower(itemMap["title"].(string)), strings.ToLower(title)) {
				filteredResponse = append(filteredResponse, item)
			}
		}
		response = filteredResponse
	}

	num := int64(len(response))
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize < 1 {
		pageSize = 10 // 默认的页大小
	}
	startIdx := (pageNum - 1) * pageSize
	endIdx := startIdx + pageSize
	if startIdx > len(response) {
		return []interface{}{}, &num // 如果起始索引超出范围，返回空数据
	}
	if endIdx > len(response) {
		endIdx = len(response)
	}
	pagedResponse := response[startIdx:endIdx]

	return pagedResponse, &num
}

func GetAllSurvey(pageNum, pageSize int, title string) ([]models.Survey, *int64, error) {
	return d.GetSurveyByTitle(ctx, title, pageNum, pageSize)
}

func SortSurvey(originalSurveys []models.Survey) []models.Survey {
	sort.Slice(originalSurveys, func(i, j int) bool {
		return originalSurveys[i].ID > originalSurveys[j].ID
	})

	status1Surveys := make([]models.Survey, 0)
	status2Surveys := make([]models.Survey, 0)
	status3Surveys := make([]models.Survey, 0)
	for _, survey := range originalSurveys {
		if survey.Deadline.Before(time.Now()) {
			survey.Status = 3
			status3Surveys = append(status3Surveys, survey)
		}

		if survey.Status == 1 {
			status1Surveys = append(status1Surveys, survey)
		} else if survey.Status == 2 {
			status2Surveys = append(status2Surveys, survey)
		}
	}

	status2Surveys = append(status2Surveys, status1Surveys...)
	sortedSurveys := append(status2Surveys, status3Surveys...)
	return sortedSurveys
}

func GetSurveyResponse(surveys []models.Survey) []interface{} {
	response := make([]interface{}, 0)
	for _, survey := range surveys {
		surveyResponse := map[string]interface{}{
			"id":     survey.ID,
			"title":  survey.Title,
			"status": survey.Status,
			"num":    survey.Num,
		}
		response = append(response, surveyResponse)
	}
	return response
}

func GetManageredSurveyByUserID(userId int) ([]models.Manage, error) {
	var manages []models.Manage
	manages, err := d.GetManageByUserID(ctx, userId)
	return manages, err
}

func GetAllSurveyAnswers(id int) (dao.AnswersResonse, error) {
	var data []dao.QuestionAnswers
	var answerSheets []dao.AnswerSheet
	var questions []models.Question
	var time []string
	questions, err := d.GetQuestionsBySurveyID(ctx, id)
	if err != nil {
		return dao.AnswersResonse{}, err
	}
	for _, question := range questions {
		var q dao.QuestionAnswers
		q.Title = question.Subject
		q.QuestionType = question.QuestionType
		data = append(data, q)
	}
	answerSheets, _, err = d.GetAnswerSheetBySurveyID(ctx, id, 0, 0, "", true)
	if err != nil {
		return dao.AnswersResonse{}, err
	}
	for _, answerSheet := range answerSheets {
		time = append(time, answerSheet.Time)
		for _, answer := range answerSheet.Answers {
			question, err := d.GetQuestionByID(ctx, answer.QuestionID)
			if err != nil {
				return dao.AnswersResonse{}, err
			}
			for i, q := range data {
				if q.Title == question.Subject {
					data[i].Answers = append(data[i].Answers, answer.Content)
				}
			}
		}
	}
	return dao.AnswersResonse{QuestionAnswers: data, Time: time}, nil
}

func GetSurveyAnswersBySurveyID(sid int) ([]dao.AnswerSheet, error) {
	answerSheets, _, err := d.GetAnswerSheetBySurveyID(ctx, sid, 0, 0, "", true)
	return answerSheets, err
}

func GetOptionByQIDAndAnswer(qid int, answer string) (*models.Option, error) {
	option, err := d.GetOptionByQIDAndAnswer(ctx, qid, answer)
	return option, err
}

func GetOptionByQIDAndSerialNum(qid int, serialNum int) (*models.Option, error) {
	option, err := d.GetOptionByQIDAndSerialNum(ctx, qid, serialNum)
	return option, err
}

func GetQuestionsByIDs(ids []int) ([]models.Question, error) {
	questions, err := d.GetQuestionsByIDs(ctx, ids)
	return questions, err
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func getOldImgs(id int, questions []models.Question) ([]string, error) {
	var imgs []string
	survey, err := d.GetSurveyByID(ctx, id)
	if err != nil {
		return nil, err
	}
	imgs = append(imgs, survey.Img)
	for _, question := range questions {
		imgs = append(imgs, question.Img)
		var options []models.Option
		options, err = d.GetOptionsByQuestionID(ctx, question.ID)
		if err != nil {
			return nil, err
		}
		for _, option := range options {
			imgs = append(imgs, option.Img)
		}
	}
	return imgs, nil
}

func getDelImgs(id int, questions []models.Question, answerSheets []dao.AnswerSheet) ([]string, error) {
	var imgs []string
	survey, err := d.GetSurveyByID(ctx, id)
	if err != nil {
		return nil, err
	}
	imgs = append(imgs, survey.Img)
	for _, question := range questions {
		imgs = append(imgs, question.Img)
		var options []models.Option
		options, err = d.GetOptionsByQuestionID(ctx, question.ID)
		if err != nil {
			return nil, err
		}
		for _, option := range options {
			imgs = append(imgs, option.Img)
		}
	}
	for _, answerSheet := range answerSheets {
		for _, answer := range answerSheet.Answers {
			question, err := d.GetQuestionByID(ctx, answer.QuestionID)
			if err != nil {
				return nil, err
			}
			if question.QuestionType == 5 {
				imgs = append(imgs, answer.Content)
			}
		}
	}
	return imgs, nil
}

func getDelFiles(answerSheets []dao.AnswerSheet) ([]string, error) {
	var files []string
	for _, answerSheet := range answerSheets {
		for _, answer := range answerSheet.Answers {
			question, err := d.GetQuestionByID(ctx, answer.QuestionID)
			if err != nil {
				return nil, err
			}
			if question.QuestionType == 6 {
				files = append(files, answer.Content)
			}
		}
	}
	return files, nil
}

func createQuestionsAndOptions(questions []dao.Question, sid int) ([]string, error) {
	var imgs []string
	for _, question := range questions {
		var q models.Question
		q.SerialNum = question.SerialNum
		q.SurveyID = sid
		q.Subject = question.Subject
		q.Description = question.Description
		q.Img = question.Img
		q.Required = question.Required
		q.Unique = question.Unique
		q.OtherOption = question.OtherOption
		q.QuestionType = question.QuestionType
		q.Reg = question.Reg
		imgs = append(imgs, question.Img)
		q, err := d.CreateQuestion(ctx, q)
		if err != nil {
			return nil, err
		}
		for _, option := range question.Options {
			var o models.Option
			o.Content = option.Content
			o.QuestionID = q.ID
			o.SerialNum = option.SerialNum
			o.Img = option.Img
			imgs = append(imgs, option.Img)
			err := d.CreateOption(ctx, o)
			if err != nil {
				return nil, err
			}
		}
	}
	return imgs, nil
}

func GetLastLinesFromLogFile(numLines int, logType int) ([]map[string]interface{}, error) {
	levelMap := map[int]string{
		0: "",
		1: "ERROR",
		2: "WARN",
		3: "INFO",
		4: "DEBUG",
	}
	level := levelMap[logType]

	var files []*os.File
	var file *os.File
	var err error

	if logType == 0 {
		// 打开所有相关的日志文件
		files, err = openAllLogFiles()
		if err != nil {
			return nil, err
		}
	} else {
		// 根据 logType 打开特定的日志文件
		file, err = openLogFile(logType)
		if err != nil {
			return nil, err
		}
		if file != nil {
			files = append(files, file)
		}
	}
	defer closeFiles(files)

	if len(files) == 0 {
		return nil, nil
	}

	// 用于存储解析后的日志内容
	var logs []map[string]interface{}

	// 从每个文件中读取内容
	for _, file := range files {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			// 解析 JSON 字符串为 map 类型
			var logData map[string]interface{}
			if err := json.Unmarshal(scanner.Bytes(), &logData); err != nil {
				// 如果解析失败，跳过这行日志继续处理下一行
				continue
			}

			// 根据 logType 筛选日志
			if level != "" {
				if logData["L"] == level {
					logs = append(logs, logData)
				}
			} else {
				logs = append(logs, logData)
			}
		}

		// 检查是否发生了读取错误
		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}

	// 如果文件中的行数不足以满足需求，直接返回所有行
	if len(logs) <= numLines {
		return logs, nil
	}

	// 如果文件中的行数超过需求，提取最后几行并返回
	startIndex := len(logs) - numLines
	return logs[startIndex:], nil
}

// 根据 logType 打开单个日志文件
func openLogFile(logType int) (*os.File, error) {
	var filePath string
	switch logType {
	case 1:
		filePath = log.LogDir + "/" + log.LogName + log.ErrorLogSuffix
	case 2:
		filePath = log.LogDir + "/" + log.LogName + log.WarnLogSuffix
	case 3, 4:
		filePath = log.LogDir + "/" + log.LogName + log.LogSuffix
	}
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // 文件不存在，返回 nil
		}
		return nil, err
	}
	return file, nil
}

// 打开所有相关的日志文件
func openAllLogFiles() ([]*os.File, error) {
	filePaths := []string{
		log.LogDir + "/" + log.LogName + log.LogSuffix,
		log.LogDir + "/" + log.LogName + log.ErrorLogSuffix,
		log.LogDir + "/" + log.LogName + log.WarnLogSuffix,
	}

	var openFiles []*os.File
	for _, filePath := range filePaths {
		f, err := os.Open(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			closeFiles(openFiles)
			return nil, err
		}
		openFiles = append(openFiles, f)
	}
	return openFiles, nil
}

// 关闭所有文件
func closeFiles(files []*os.File) {
	for _, file := range files {
		file.Close()
	}
}

func DeleteAnswerSheetBySurveyID(surveyID int) error {
	err := d.DeleteAnswerSheetBySurveyID(ctx, surveyID)
	return err
}

func aesDecryptPassword(user *models.User) {
	user.Password = utils.AesDecrypt(user.Password)
}

func aesEncryptPassword(user *models.User) {
	user.Password = utils.AesEncrypt(user.Password)
}

func HandleDownloadFile(answers dao.AnswersResonse, survey *models.Survey) (string, error) {
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
	rowData := make([]interface{}, 0)
	rowData = append(rowData, excelize.Cell{Value: "序号", StyleID: styleID}, excelize.Cell{Value: "提交时间", StyleID: styleID})
	for _, qa := range questionAnswers {
		rowData = append(rowData, excelize.Cell{Value: qa.Title, StyleID: styleID})
	}
	if err := streamWriter.SetRow("A1", rowData); err != nil {
		return "", errors.New("写入标题行失败原因: " + err.Error())
	}
	// 写入数据
	for i, time := range times {
		row := []interface{}{i + 1, time}
		for j, qa := range questionAnswers {
			if len(qa.Answers) <= i {
				continue
			}
			answer := qa.Answers[i]
			row = append(row, answer)
			colName, _ := excelize.ColumnNumberToName(j + 3)
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
		err := os.Mkdir("./public/xlsx/", 0755)
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


func UpdateAdminPassword(id int, password string) error {
	password = utils.AesEncrypt(password)
	err := d.UpdateUserPassword(ctx, id, password)
	return err
}