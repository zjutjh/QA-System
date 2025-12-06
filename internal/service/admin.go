package service

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"QA-System/internal/dao"
	"QA-System/internal/model"
	"QA-System/internal/pkg/oss"
	"QA-System/internal/pkg/utils"

	"github.com/xuri/excelize/v2"
	"github.com/yitter/idgenerator-go/idgen"
	"github.com/zjutjh/WeJH-SDK/excel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetAdminByUsername 根据用户名获取管理员
func GetAdminByUsername(username string) (*model.User, error) {
	user, err := d.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if user.Password != "" {
		aesDecryptPassword(user)
	}
	return user, nil
}

// GetAdminByID 根据ID获取管理员
func GetAdminByID(id int) (*model.User, error) {
	user, err := d.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user.Password != "" {
		aesDecryptPassword(user)
	}
	return user, nil
}

// GetUserEmailByID 根据用户ID获取用户邮箱
func GetUserEmailByID(id int) (string, error) {
	email, err := d.GetUserEmailByID(ctx, id)
	if err != nil {
		return "", err
	}
	return email, nil
}

// IsAdminExist 判断管理员是否存在
func IsAdminExist(username string) error {
	_, err := d.GetUserByUsername(ctx, username)
	return err
}

// CreateAdmin 创建管理员
func CreateAdmin(user model.User) error {
	aesEncryptPassword(&user)
	err := d.CreateUser(ctx, &user)
	return err
}

// GetUserByName 根据用户名获取用户
func GetUserByName(username string) (*model.User, error) {
	user, err := d.GetUserByUsername(ctx, username)
	return user, err
}

// CreatePermission 创建权限
func CreatePermission(id int, surveyID int64) error {
	err := d.CreateManage(ctx, id, surveyID)
	return err
}

// DeletePermission 删除权限
func DeletePermission(id int, surveyID int64) error {
	err := d.DeleteManage(ctx, id, surveyID)
	return err
}

// CheckPermission 检查权限
func CheckPermission(id int, surveyID int64) error {
	err := d.CheckManage(ctx, id, surveyID)
	return err
}

// CreateSurvey 创建问卷
func CreateSurvey(id int, question_list []dao.QuestionList, status int, surveyType, limit uint,
	sumLimit uint, verify, undergradOnly bool, ddl, startTime time.Time, title string, desc string,
	neednot bool) error {
	var survey model.Survey
	survey.ID = idgen.NextId()
	survey.UserID = id
	survey.Status = status
	survey.Deadline = ddl
	survey.Type = surveyType
	survey.DailyLimit = limit
	survey.SumLimit = sumLimit
	survey.Verify = verify
	survey.UndergradOnly = undergradOnly
	survey.StartTime = startTime
	survey.Title = title
	survey.Desc = desc
	survey.NeedNotify = neednot
	survey, err := d.CreateSurvey(ctx, survey)
	if err != nil {
		return err
	}
	_, err = createQuestionsAndOptions(question_list, survey.ID)
	return err
}

// UpdateSurveyStatus 更新问卷状态
func UpdateSurveyStatus(id int64, status int) error {
	err := d.UpdateSurveyStatus(ctx, id, status)
	return err
}

// UpdateSurvey 更新问卷
func UpdateSurvey(id int64, question_list []dao.QuestionList, surveyType,
	limit uint, sumLimit uint, verify, undergradOnly bool, desc string, title string, ddl, startTime time.Time,
	needNotify bool) error {
	// 遍历原有问题，删除对应选项
	var oldQuestions []model.Question
	var oldImgs []string
	newImgs := make([]string, 0)
	// 获取原有图片
	oldQuestions, err := d.GetQuestionsBySurveyID(ctx, id)
	if err != nil {
		return err
	}
	oldImgs, err = getOldImgs(oldQuestions)
	if err != nil {
		return err
	}
	// 删除原有问题和选项
	for _, oldQuestion := range oldQuestions {
		oldOptions, err := d.GetOptionsByQuestionID(ctx, oldQuestion.ID)
		if err != nil {
			return err
		}
		for _, oldOption := range oldOptions {
			err = d.DeleteOption(ctx, oldOption.ID)
			if err != nil {
				return err
			}
		}
		err = d.DeleteQuestion(ctx, oldQuestion.ID)
		if err != nil {
			return err
		}
		err = dao.DeleteAllQuestionCache(ctx)
		if err != nil {
			return err
		}
		err = dao.DeleteAllOptionCache(ctx)
		if err != nil {
			return err
		}
	}
	// 修改问卷信息
	err = d.UpdateSurvey(ctx, id, surveyType, limit, sumLimit, verify, undergradOnly, desc, title, ddl, startTime,
		needNotify)
	if err != nil {
		return err
	}
	// 重新添加问题和选项
	imgs, err := createQuestionsAndOptions(question_list, id)
	if err != nil {
		return err
	}
	newImgs = append(newImgs, imgs...)
	// 删除无用图片
	for _, oldImg := range oldImgs {
		if !contains(newImgs, oldImg) {
			_, err = oss.Client.DeleteFile(oss.Client.GetObjectKeyFromUrl(oldImg))
			if err != nil {
				zap.L().Warn("删除旧图片失败", zap.String("img", oldImg), zap.Error(err))
			}
		}
	}
	return nil
}

// UserInManage 用户是否在管理中
func UserInManage(uid int, sid int64) bool {
	_, err := d.GetManageByUIDAndSID(ctx, uid, sid)
	return err == nil
}

// DeleteSurvey 删除问卷
func DeleteSurvey(id int64) error {
	var questions []model.Question
	questions, err := d.GetQuestionsBySurveyID(ctx, id)
	if err != nil {
		return err
	}
	var answerSheets []dao.AnswerSheet
	answerSheets, _, err = d.GetAnswerSheetBySurveyID(ctx, id, 0, 0, "", false)
	if err != nil {
		return err
	}
	// 删除图片
	imgs, err := getDelImgs(questions, answerSheets)
	if err != nil {
		return err
	}
	// 删除文件
	files, err := getDelFiles(answerSheets)
	if err != nil {
		return err
	}
	for _, img := range imgs {
		_, err = oss.Client.DeleteFile(oss.Client.GetObjectKeyFromUrl(img))
		if err != nil {
			zap.L().Warn("删除旧图片失败", zap.String("img", img), zap.Error(err))
		}
	}

	for _, file := range files {
		_, err = oss.Client.DeleteFile(oss.Client.GetObjectKeyFromUrl(file))
		if err != nil {
			zap.L().Warn("删除旧文件失败", zap.String("file", file), zap.Error(err))
		}
	}
	// 删除答卷
	err = DeleteAnswerSheetBySurveyID(id)
	if err != nil {
		return err
	}
	// 删除问题、选项、问卷、管理
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
	err = dao.DeleteAllQuestionCache(ctx)
	if err != nil {
		return err
	}
	err = dao.DeleteAllOptionCache(ctx)
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

// GetSurveyAnswers 获取问卷答案
func GetSurveyAnswers(id int64, num int, size int, text string, unique bool) (dao.AnswersResonse, *int64, error) {
	var answerSheets []dao.AnswerSheet
	data := make([]dao.QuestionAnswers, 0)
	times := make([]string, 0)
	aids := make([]primitive.ObjectID, 0)
	var total *int64
	// 获取问题
	questions, err := d.GetQuestionsBySurveyID(ctx, id)
	if err != nil {
		return dao.AnswersResonse{}, nil, err
	}
	// 初始化data
	for _, question := range questions {
		var q dao.QuestionAnswers
		q.Title = question.Subject
		q.QuestionType = question.QuestionType
		q.Answers = make([]string, 0)
		data = append(data, q)
	}
	// 获取答卷
	answerSheets, total, err = d.GetAnswerSheetBySurveyID(ctx, id, num, size, text, unique)
	if err != nil {
		return dao.AnswersResonse{}, nil, err
	}
	// 填充data
	for _, answerSheet := range answerSheets {
		times = append(times, answerSheet.Time)
		aids = append(aids, answerSheet.AnswerID)
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
	return dao.AnswersResonse{QuestionAnswers: data, AnswerIDs: aids, Time: times}, total, nil
}

// GetSurveyByUserID 获取用户的所有问卷
func GetSurveyByUserID(userId int) ([]model.Survey, error) {
	return d.GetSurveyByUserID(ctx, userId)
}

// ProcessResponse 处理响应
func ProcessResponse(response []model.SurveyResp, pageNum, pageSize int, title string) ([]model.SurveyResp, int) {
	resp := response
	if title != "" {
		filteredResponse := make([]model.SurveyResp, 0)
		for _, item := range response {
			if strings.Contains(strings.ToLower(item.Title), strings.ToLower(title)) {
				filteredResponse = append(filteredResponse, item)
			}
		}
		resp = filteredResponse
	}

	num := len(resp)
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize < 1 {
		pageSize = 10 // 默认的页大小
	}
	startIdx := (pageNum - 1) * pageSize
	endIdx := startIdx + pageSize
	if startIdx > len(resp) {
		return []model.SurveyResp{}, num // 如果起始索引超出范围，返回空数据
	}
	if endIdx > len(resp) {
		endIdx = len(resp)
	}
	pagedResponse := resp[startIdx:endIdx]

	return pagedResponse, num
}

// GetAllSurvey 获取所有问卷
func GetAllSurvey() ([]model.Survey, error) {
	return d.GetAllSurvey(ctx)
}

// SortSurvey 排序问卷
func SortSurvey(originalSurveys []model.Survey) []model.Survey {
	sort.Slice(originalSurveys, func(i, j int) bool {
		return originalSurveys[i].CreatedAt.After(originalSurveys[j].CreatedAt)
	})

	status1Surveys := make([]model.Survey, 0)
	status2Surveys := make([]model.Survey, 0)
	status3Surveys := make([]model.Survey, 0)
	for _, survey := range originalSurveys {
		if survey.Deadline.Before(time.Now()) {
			survey.Status = 3
			status3Surveys = append(status3Surveys, survey)
			continue
		}

		if survey.Status == 1 {
			status1Surveys = append(status1Surveys, survey)
		} else if survey.Status == 2 {
			status2Surveys = append(status2Surveys, survey)
		}
	}

	sortedSurveys := append(append(status2Surveys, status1Surveys...), status3Surveys...)
	return sortedSurveys
}

// GetSurveyResponse 获取问卷响应
func GetSurveyResponse(surveys []model.Survey) []model.SurveyResp {
	response := make([]model.SurveyResp, 0)
	for _, survey := range surveys {
		surveyResponse := model.SurveyResp{
			ID:         survey.ID,
			Title:      survey.Title,
			Status:     survey.Status,
			SurveyType: survey.Type,
			Num:        survey.Num,
		}
		response = append(response, surveyResponse)
	}
	return response
}

// GetManagedSurveyByUserID 获取用户管理的问卷
func GetManagedSurveyByUserID(userId int) ([]model.Manage, error) {
	var manages []model.Manage
	manages, err := d.GetManageByUserID(ctx, userId)
	return manages, err
}

// GetAllSurveyAnswers 获取所有问卷答案
func GetAllSurveyAnswers(id int64) (dao.AnswersResonse, error) {
	data := make([]dao.QuestionAnswers, 0)
	answerSheets := make([]dao.AnswerSheet, 0)
	questions := make([]model.Question, 0)
	times := make([]string, 0)
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
		times = append(times, answerSheet.Time)
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
	return dao.AnswersResonse{QuestionAnswers: data, Time: times}, nil
}

// GetSurveyAnswersBySurveyID 根据问卷编号获取问卷答案
func GetSurveyAnswersBySurveyID(sid int64) ([]dao.AnswerSheet, error) {
	answerSheets, _, err := d.GetAnswerSheetBySurveyID(ctx, sid, 0, 0, "", true)
	return answerSheets, err
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func getOldImgs(questions []model.Question) ([]string, error) {
	imgs := make([]string, 0)
	for _, question := range questions {
		imgs = append(imgs, question.Img)
		var options []model.Option
		options, err := d.GetOptionsByQuestionID(ctx, question.ID)
		if err != nil {
			return nil, err
		}
		for _, option := range options {
			imgs = append(imgs, option.Img)
		}
	}
	return imgs, nil
}

func getDelImgs(questions []model.Question, answerSheets []dao.AnswerSheet) ([]string, error) {
	imgs := make([]string, 0)
	for _, question := range questions {
		if question.Img != "" {
			imgs = append(imgs, question.Img)
		}
		var options []model.Option
		options, err := d.GetOptionsByQuestionID(ctx, question.ID)
		if err != nil {
			return nil, err
		}
		for _, option := range options {
			if option.Img != "" {
				imgs = append(imgs, option.Img)
			}
		}
	}
	for _, answerSheet := range answerSheets {
		for _, answer := range answerSheet.Answers {
			question, err := d.GetQuestionByID(ctx, answer.QuestionID)
			if err != nil {
				return nil, err
			}
			if question.QuestionType == 5 && answer.Content != "" {
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

func createQuestionsAndOptions(question_list []dao.QuestionList, sid int64) ([]string, error) {
	imgs := make([]string, 0)
	for _, question_list := range question_list {
		var q model.Question
		q.SerialNum = question_list.SerialNum
		q.SurveyID = sid
		q.Subject = question_list.Subject
		q.Description = question_list.Description
		q.Img = question_list.Img
		q.Required = question_list.QuestionSetting.Required
		q.Unique = question_list.QuestionSetting.Unique
		q.OtherOption = question_list.QuestionSetting.OtherOption
		q.QuestionType = question_list.QuestionSetting.QuestionType
		q.MaximumOption = question_list.QuestionSetting.MaximumOption
		q.MinimumOption = question_list.QuestionSetting.MinimumOption
		q.Reg = question_list.QuestionSetting.Reg
		imgs = append(imgs, question_list.Img)
		q, err := d.CreateQuestion(ctx, q)
		if err != nil {
			return nil, err
		}
		for _, option := range question_list.Options {
			var o model.Option
			o.Content = option.Content
			o.QuestionID = q.ID
			o.SerialNum = option.SerialNum
			o.Img = option.Img
			o.Description = option.Description
			imgs = append(imgs, option.Img)
			err := d.CreateOption(ctx, o)
			if err != nil {
				return nil, err
			}
		}
	}
	return imgs, nil
}

// DeleteAnswerSheetBySurveyID 根据问卷编号删除问卷答案
func DeleteAnswerSheetBySurveyID(surveyID int64) error {
	err := d.DeleteAnswerSheetBySurveyID(ctx, surveyID)
	return err
}

func aesDecryptPassword(user *model.User) {
	user.Password = utils.AesDecrypt(user.Password)
}

func aesEncryptPassword(user *model.User) {
	user.Password = utils.AesEncrypt(user.Password)
}

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

// UpdateAdminPassword 更新管理员密码
func UpdateAdminPassword(id int, password string) error {
	encryptedPassword := utils.AesEncrypt(password)
	err := d.UpdateUserPassword(ctx, id, encryptedPassword)
	return err
}

// UpdateAdminEmail 更新管理员邮箱
func UpdateAdminEmail(id int, email string) error {
	err := d.UpdateUserEmail(ctx, id, email)
	return err
}

// CreateQuestionPre 创建问题预先信息
func CreateQuestionPre(name string, value []string) error {
	// 将String[]类型转化为String,以逗号分隔
	pre := strings.Join(value, ",")
	err := d.CreateType(ctx, name, pre)
	return err
}

// GetQuestionPre 获取问题预先信息
func GetQuestionPre(name string) ([]string, error) {
	value, err := d.GetType(ctx, name)
	if err != nil {
		return nil, err
	}

	// 将预先信息转化为String[]类型
	pre := strings.Split(value, ",")
	return pre, nil
}

// DeleteOauthRecord 删除统一记录
func DeleteOauthRecord(sid int64) error {
	return d.DeleteRecordSheets(ctx, sid)
}

// DeleteAnswerSheetByAnswerID 根据问卷ID删除问卷
func DeleteAnswerSheetByAnswerID(answerID primitive.ObjectID) error {
	err := d.DeleteAnswerSheetByAnswerID(ctx, answerID)
	return err
}

// GetAnswerSheetByAnswerID 根据答卷ID删除答卷
func GetAnswerSheetByAnswerID(answerID primitive.ObjectID) error {
	err := d.GetAnswerSheetByAnswerID(ctx, answerID)
	return err
}

// GetOptionCount 选项数据
type GetOptionCount struct {
	SerialNum int    `json:"serial_num"` // 选项序号
	Content   string `json:"content"`    // 选项内容
	Count     int    `json:"count"`      // 选项数量
	Percent   string `json:"percent"`    // 占比百分比，保留两位小数
}

// GetChooseStatisticsResponse 问题模型
type GetChooseStatisticsResponse struct {
	SerialNum    int              `json:"serial_num"`    // 问题序号
	Question     string           `json:"question"`      // 问题内容
	QuestionType int              `json:"question_type"` // 问题类型  1:单选 2:多选
	Options      []GetOptionCount `json:"options"`       // 选项内容
}

// GenerateQuestionStats 生成问卷题目统计结果
func GenerateQuestionStats(questions []model.Question, answerSheets []dao.AnswerSheet) []GetChooseStatisticsResponse {
	questionMap := make(map[int]model.Question)
	optionsMap := make(map[int][]model.Option)
	optionAnswerMap := make(map[int]map[string]model.Option)
	optionSerialNumMap := make(map[int]map[int]model.Option)
	for _, question := range questions {
		questionMap[question.ID] = question
		optionAnswerMap[question.ID] = make(map[string]model.Option)
		optionSerialNumMap[question.ID] = make(map[int]model.Option)
		options, err := GetOptionsByQuestionID(question.ID)
		if err != nil {
			log.Println("Error fetching options for questionID:", question.ID)
			continue
		}
		optionsMap[question.ID] = options
		for _, option := range options {
			optionAnswerMap[question.ID][option.Content] = option
			optionSerialNumMap[question.ID][option.SerialNum] = option
		}
	}

	optionCounts := make(map[int]map[int]int)
	for _, sheet := range answerSheets {
		for _, answer := range sheet.Answers {
			options := optionsMap[answer.QuestionID]
			question := questionMap[answer.QuestionID]

			// 初始化外层 map：如果某题还没记录，先创建一个 map[int]int 作为它的值
			if _, ok := optionCounts[question.ID]; !ok {
				optionCounts[question.ID] = make(map[int]int)

				// 初始化所有选项的计数为 0（避免后续统计遗漏）
				for _, option := range options {
					optionCounts[question.ID][option.SerialNum] = 0
				}
			}

			// 支持单选题、多选题：用 "┋" 分割用户选的答案
			if question.QuestionType == 1 || question.QuestionType == 2 {
				answerOptions := strings.Split(answer.Content, "┋")
				questionOptions := optionAnswerMap[answer.QuestionID]

				for _, answerOption := range answerOptions {
					if questionOptions != nil {
						option, exists := questionOptions[answerOption]
						if exists {
							// 该选项存在，序号对应的次数 +1
							optionCounts[answer.QuestionID][option.SerialNum]++
							continue
						}
					}
					// “其他”选项，统一用 SerialNum = 0 表示
					optionCounts[answer.QuestionID][0]++
				}
			}
		}
	}
	response := make([]GetChooseStatisticsResponse, 0, len(optionCounts))
	for qid, optionCountMap := range optionCounts {
		q := questionMap[qid]
		if q.QuestionType != 1 && q.QuestionType != 2 {
			continue
		}
		total := 0
		for _, count := range optionCountMap {
			total += count
		}

		var qOptions []GetOptionCount
		if q.OtherOption {
			qOptions = append(qOptions, GetOptionCount{
				SerialNum: 0,
				Content:   "其他",
				Count:     optionCountMap[0],
				Percent:   fmt.Sprintf("%.2f%%", float64(optionCountMap[0])*100/float64(total)),
			})
		}

		// 排序
		serialNums := make([]int, 0, len(optionCountMap))
		for serial := range optionCountMap {
			if serial != 0 {
				serialNums = append(serialNums, serial)
			}
		}
		sort.Ints(serialNums)

		for _, serial := range serialNums {
			count := optionCountMap[serial]
			op := optionSerialNumMap[qid][serial]
			percent := "0.00%"
			if total > 0 {
				percent = fmt.Sprintf("%.2f%%", float64(count)*100/float64(total))
			}
			qOptions = append(qOptions, GetOptionCount{
				SerialNum: serial,
				Content:   op.Content,
				Count:     count,
				Percent:   percent,
			})
		}

		response = append(response, GetChooseStatisticsResponse{
			SerialNum:    q.SerialNum,
			Question:     q.Subject,
			QuestionType: q.QuestionType,
			Options:      qOptions,
		})
	}
	// 按序号排序
	sort.Slice(response, func(i, j int) bool {
		return response[i].SerialNum < response[j].SerialNum
	})
	return response
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
