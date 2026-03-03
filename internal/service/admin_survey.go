package service

import (
	"sort"
	"strings"
	"time"

	"QA-System/internal/dao"
	"QA-System/internal/model"
	"QA-System/internal/pkg/oss"

	"github.com/yitter/idgenerator-go/idgen"
	"go.uber.org/zap"
)

// CreateSurvey 创建问卷
func CreateSurvey(id int, question_list []model.QuestionList, status int, surveyType, limit uint,
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

	return dao.RunInTx(d, ctx, func(store dao.Daos) error {
		createdSurvey, err := store.CreateSurvey(ctx, survey)
		if err != nil {
			return err
		}
		_, err = createQuestionsAndOptionsWithStore(store, question_list, createdSurvey.ID)
		return err
	})
}

// UpdateSurveyStatus 更新问卷状态
func UpdateSurveyStatus(id int64, status int) error {
	err := d.UpdateSurveyStatus(ctx, id, status)
	return err
}

// UpdateSurvey 更新问卷
func UpdateSurvey(id int64, question_list []model.QuestionList, surveyType,
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
	// 在事务内删除原有问题/选项并重建，避免中途失败导致 MySQL 数据半更新。
	err = dao.RunInTx(d, ctx, func(store dao.Daos) error {
		for _, oldQuestion := range oldQuestions {
			// DeleteOption 按 question_id 删除，不能传 option.id
			if err := store.DeleteOption(ctx, oldQuestion.ID); err != nil {
				return err
			}
			if err := store.DeleteQuestion(ctx, oldQuestion.ID); err != nil {
				return err
			}
		}

		if err := store.UpdateSurvey(ctx, id, surveyType, limit, sumLimit, verify, undergradOnly, desc, title, ddl,
			startTime, needNotify); err != nil {
			return err
		}

		imgs, err := createQuestionsAndOptionsWithStore(store, question_list, id)
		if err != nil {
			return err
		}
		newImgs = append(newImgs, imgs...)
		return nil
	})
	if err != nil {
		return err
	}
	// 结构更新后统一清缓存，避免循环内重复全量扫描。
	err = dao.DeleteAllQuestionCache(ctx)
	if err != nil {
		return err
	}
	err = dao.DeleteAllOptionCache(ctx)
	if err != nil {
		return err
	}
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

		switch survey.Status {
		case 1:
			status1Surveys = append(status1Surveys, survey)
		case 2:
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

func createQuestionsAndOptions(questionList []model.QuestionList, sid int64) ([]string, error) {
	return createQuestionsAndOptionsWithStore(d, questionList, sid)
}

func createQuestionsAndOptionsWithStore(store dao.Daos, questionList []model.QuestionList, sid int64) ([]string, error) {
	imgs := make([]string, 0)
	for _, questionItem := range questionList {
		var q model.Question
		q.SerialNum = questionItem.SerialNum
		q.SurveyID = sid
		q.Subject = questionItem.Subject
		q.Description = questionItem.Description
		q.Img = questionItem.Img
		q.Required = questionItem.QuestionSetting.Required
		q.Unique = questionItem.QuestionSetting.Unique
		q.OtherOption = questionItem.QuestionSetting.OtherOption
		q.QuestionType = questionItem.QuestionSetting.QuestionType
		q.MaximumOption = questionItem.QuestionSetting.MaximumOption
		q.MinimumOption = questionItem.QuestionSetting.MinimumOption
		q.Reg = questionItem.QuestionSetting.Reg
		imgs = append(imgs, questionItem.Img)
		q, err := store.CreateQuestion(ctx, q)
		if err != nil {
			return nil, err
		}
		for _, option := range questionItem.Options {
			var o model.Option
			o.Content = option.Content
			o.QuestionID = q.ID
			o.SerialNum = option.SerialNum
			o.Img = option.Img
			o.Description = option.Description
			imgs = append(imgs, option.Img)
			err := store.CreateOption(ctx, o)
			if err != nil {
				return nil, err
			}
		}
	}
	return imgs, nil
}
