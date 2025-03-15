package service

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"QA-System/internal/dao"
	"QA-System/internal/model"
	"QA-System/internal/pkg/utils"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
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
func CreatePermission(id int, surveyID string) error {
	err := d.CreateManage(ctx, id, surveyID)
	return err
}

// DeletePermission 删除权限
func DeletePermission(id int, surveyID string) error {
	err := d.DeleteManage(ctx, id, surveyID)
	return err
}

// CheckPermission 检查权限
func CheckPermission(id int, surveyID string) error {
	err := d.CheckManage(ctx, id, surveyID)
	return err
}

// CreateSurvey 创建问卷
func CreateSurvey(id int, question_list []dao.QuestionList, status int, surveyType, limit uint,
	verify bool, ddl, startTime time.Time, title string, desc string) error {
	var survey model.Survey
	new_uuid := uuid.NewString()
	_, err := d.GetSurveyByUUID(ctx, new_uuid)
	for err == nil {
		new_uuid = uuid.NewString()
		_, err = d.GetSurveyByUUID(ctx, new_uuid)
	}
	survey.UUID = new_uuid
	survey.UserID = id
	survey.Status = status
	survey.Deadline = ddl
	survey.Type = surveyType
	survey.DailyLimit = limit
	survey.Verify = verify
	survey.StartTime = startTime
	survey.Title = title
	survey.Desc = desc
	survey, err = d.CreateSurvey(ctx, survey)
	if err != nil {
		return err
	}
	_, err = createQuestionsAndOptions(question_list, survey.UUID)
	return err
}

// UpdateSurveyStatus 更新问卷状态
func UpdateSurveyStatus(sid string, status int) error {
	err := d.UpdateSurveyStatus(ctx, sid, status)
	return err
}

// UpdateSurvey 更新问卷
func UpdateSurvey(sid string, question_list []dao.QuestionList, surveyType,
	limit uint, verify bool, desc string, title string, ddl, startTime time.Time) error {
	// 遍历原有问题，删除对应选项
	var oldQuestions []model.Question
	var old_imgs []string
	new_imgs := make([]string, 0)
	// 获取原有图片
	oldQuestions, err := d.GetQuestionsBySurveyID(ctx, sid)
	if err != nil {
		return err
	}
	old_imgs, err = getOldImgs(oldQuestions)
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
	err = d.UpdateSurvey(ctx, sid, surveyType, limit, verify, desc, title, ddl, startTime)
	if err != nil {
		return err
	}
	new_imgs = append(new_imgs, question_list[1].Img)
	// 重新添加问题和选项
	imgs, err := createQuestionsAndOptions(question_list, sid)
	if err != nil {
		return err
	}
	new_imgs = append(new_imgs, imgs...)
	urlHost := GetConfigUrl()
	// 删除无用图片
	for _, oldImg := range old_imgs {
		if !contains(new_imgs, oldImg) {
			err = os.Remove("./public/static/" + strings.TrimPrefix(oldImg, urlHost+"/public/static/"))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// UserInManage 用户是否在管理中
func UserInManage(uid int, sid string) bool {
	_, err := d.GetManageByUIDAndSID(ctx, uid, sid)
	return err == nil
}

// DeleteSurvey 删除问卷
func DeleteSurvey(sid string) error {
	var questions []model.Question
	questions, err := d.GetQuestionsBySurveyID(ctx, sid)
	if err != nil {
		return err
	}
	var answerSheets []dao.AnswerSheet
	answerSheets, _, err = d.GetAnswerSheetBySurveyID(ctx, sid, 0, 0, "", false)
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
	urlHost := GetConfigUrl()
	for _, img := range imgs {
		err = os.Remove("./public/static/" + strings.TrimPrefix(img, urlHost+"/public/static/"))
		if err != nil {
			return err
		}
	}
	for _, file := range files {
		err = os.Remove("./public/file/" + strings.TrimPrefix(file, urlHost+"/public/file/"))
		if err != nil {
			return err
		}
	}
	// 删除答卷
	err = DeleteAnswerSheetBySurveyID(sid)
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
	err = d.DeleteQuestionBySurveyID(ctx, sid)
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
	err = d.DeleteSurvey(ctx, sid)
	if err != nil {
		return err
	}
	err = d.DeleteManageBySurveyID(ctx, sid)
	return err
}

// GetSurveyAnswers 获取问卷答案
func GetSurveyAnswers(sid string, num int, size int, text string, unique bool) (dao.AnswersResonse, *int64, error) {
	var answerSheets []dao.AnswerSheet
	data := make([]dao.QuestionAnswers, 0)
	times := make([]string, 0)
	aids := make([]primitive.ObjectID, 0)
	var total *int64
	// 获取问题
	questions, err := d.GetQuestionsBySurveyID(ctx, sid)
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
	answerSheets, total, err = d.GetAnswerSheetBySurveyID(ctx, sid, num, size, text, unique)
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
			UUID:       survey.UUID,
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
func GetAllSurveyAnswers(sid string) (dao.AnswersResonse, error) {
	data := make([]dao.QuestionAnswers, 0)
	answerSheets := make([]dao.AnswerSheet, 0)
	questions := make([]model.Question, 0)
	times := make([]string, 0)
	questions, err := d.GetQuestionsBySurveyID(ctx, sid)
	if err != nil {
		return dao.AnswersResonse{}, err
	}
	for _, question := range questions {
		var q dao.QuestionAnswers
		q.Title = question.Subject
		q.QuestionType = question.QuestionType
		data = append(data, q)
	}
	answerSheets, _, err = d.GetAnswerSheetBySurveyID(ctx, sid, 0, 0, "", true)
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
func GetSurveyAnswersBySurveyID(sid string) ([]dao.AnswerSheet, error) {
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
				imgs = append(imgs, question.Img)
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

func createQuestionsAndOptions(question_list []dao.QuestionList, sid string) ([]string, error) {
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
func DeleteAnswerSheetBySurveyID(surveyID string) error {
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
func DeleteOauthRecord(sid string) error {
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
