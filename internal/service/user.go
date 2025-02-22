package service

import (
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"QA-System/internal/dao"
	"QA-System/internal/model"
	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nfnt/resize"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

// GetSurveyByID 根据ID获取问卷
func GetSurveyByID(id int) (*model.Survey, error) {
	survey, err := d.GetSurveyByID(ctx, id)
	return survey, err
}

// GetQuestionsBySurveyID 根据问卷ID获取问题
func GetQuestionsBySurveyID(sid int) ([]model.Question, error) {
	var questions []model.Question
	questions, err := d.GetQuestionsBySurveyID(ctx, sid)
	return questions, err
}

// GetOptionsByQuestionID 根据问题ID获取选项
func GetOptionsByQuestionID(questionId int) ([]model.Option, error) {
	var options []model.Option
	options, err := d.GetOptionsByQuestionID(ctx, questionId)
	return options, err
}

// GetQuestionByID 根据问卷ID获取问题
func GetQuestionByID(id int) (*model.Question, error) {
	var question *model.Question
	question, err := d.GetQuestionByID(ctx, id)
	return question, err
}

// SubmitSurvey 提交问卷
func SubmitSurvey(sid int, data []dao.QuestionsList, t string) error {
	var answerSheet dao.AnswerSheet
	answerSheet.SurveyID = sid
	answerSheet.Time = t
	answerSheet.Unique = true
	answerSheet.AnswerID = primitive.NewObjectID()
	qids := make([]int, 0)
	for _, q := range data {
		var answer dao.Answer
		question, err := d.GetQuestionByID(ctx, q.QuestionID)
		if err != nil {
			return err
		}
		if question.QuestionType == 3 && question.Unique {
			qids = append(qids, q.QuestionID)
		}
		answer.QuestionID = q.QuestionID
		answer.Content = q.Answer
		answerSheet.Answers = append(answerSheet.Answers, answer)
	}
	err := d.SaveAnswerSheet(ctx, answerSheet, qids)
	if err != nil {
		return err
	}
	err = d.IncreaseSurveyNum(ctx, sid)
	return err
}

// HandleImgUpload 处理图片上传
func HandleImgUpload(c *gin.Context) (string, error) {
	// 保存图片文件
	file, err := c.FormFile("img")
	if err != nil {
		return "", errors.New("获取文件失败")
	}
	// 检查文件类型是否为图像
	if !isImageFile(file) {
		return "", errors.New("文件类型不是图片")
	}
	// 检查文件大小是否超出限制
	if file.Size > 10<<20 {
		return "", errors.New("文件大小超出限制")
	}
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "tempdir")
	if err != nil {
		return "", errors.New("创建临时目录失败: " + err.Error())
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			zap.L().Error("删除临时目录失败", zap.Error(err))
		}
	}() // 在处理完之后删除临时目录及其中的文件
	// 在临时目录中创建临时文件
	tempFile := filepath.Join(tempDir, file.Filename)
	f, err := safeCreateFile(tempFile)
	if err != nil {
		return "", err
	}
	// 将上传的文件保存到临时文件中
	src, err := file.Open()
	if err != nil {
		return "", errors.New("打开文件失败: " + err.Error())
	}
	defer func() {
		if err := src.Close(); err != nil {
			zap.L().Error("关闭文件失败", zap.Error(err))
		}
	}()

	_, err = io.Copy(f, src)
	if err != nil {
		return "", errors.New("保存文件失败: " + err.Error())
	}
	// 判断文件的MIME类型是否为图片
	mime, err := mimetype.DetectFile(tempFile)
	if err != nil || !strings.HasPrefix(mime.String(), "image/") {
		return "", errors.New("文件类型不是图片")
	}
	// 保存原始图片
	filename := uuid.New().String() + ".jpg"
	dst := "./public/static/" + filename
	err = c.SaveUploadedFile(file, dst)
	if err != nil {
		return "", errors.New("保存文件失败: " + err.Error())
	}

	// 转换图像为JPG格式并压缩
	jpgFile := filepath.Join(tempDir, "compressed.jpg")
	err = convertAndCompressImage(dst, jpgFile)
	if err != nil {
		return "", errors.New("转换和压缩图像失败: " + err.Error())
	}

	// 替换原始文件为压缩后的JPG文件
	err = os.Rename(jpgFile, dst)
	if err != nil {
		err = copyFile(jpgFile, dst)
		if err != nil {
			return "", errors.New("替换文件失败: " + err.Error())
		}
		err = os.Remove(jpgFile)
		if err != nil {
			return "", errors.New("删除临时文件失败: " + err.Error())
		}
	}

	urlHost := GetConfigUrl()
	url := urlHost + "/public/static/" + filename

	return url, nil
}

// 仅支持常见的图像文件类型
func isImageFile(file *multipart.FileHeader) bool {
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
	}
	return allowedTypes[file.Header.Get("Content-Type")]
}

// 用于转换和压缩图像的函数
func convertAndCompressImage(srcPath, dstPath string) error {
	// 打开源图像文件
	srcFile, err := safeOpenFile(srcPath)
	if err != nil {
		return err
	}

	// 解码图像
	srcImg, _, err := image.Decode(srcFile)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// 调整图像大小（根据需要进行调整）
	resizedImg := resize.Resize(300, 0, srcImg, resize.Lanczos3)

	// 创建新的JPG文件
	dstFile, err := safeCreateFile(dstPath)
	if err != nil {
		return err
	}

	// 以JPG格式保存调整大小的图像，并设置压缩质量为90
	err = jpeg.Encode(dstFile, resizedImg, &jpeg.Options{Quality: 90})
	if err != nil {
		return err
	}

	return nil
}

func copyFile(src, dst string) error {
	srcFile, err := safeOpenFile(src)
	if err != nil {
		return err
	}
	defer func(srcFile *os.File) {
		err := srcFile.Close()
		if err != nil {
			zap.L().Error("关闭文件失败", zap.Error(err))
		}
	}(srcFile)

	dstFile, err := safeCreateFile(dst)
	if err != nil {
		return err
	}

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	return nil
}

// HandleFileUpload 处理文件上传
func HandleFileUpload(c *gin.Context) (string, error) {
	// 保存文件
	file, err := c.FormFile("file")
	if err != nil {
		return "", errors.New("获取文件失败")
	}
	// 检查文件大小是否超出限制，限制为50MB
	if file.Size > 50<<20 {
		return "", errors.New("文件大小超出限制")
	}
	// 保存文件
	filename := uuid.New().String() + filepath.Ext(file.Filename)
	dst := "./public/file/" + filename
	err = c.SaveUploadedFile(file, dst)
	if err != nil {
		return "", errors.New("保存文件失败: " + err.Error())
	}
	urlHost := GetConfigUrl()
	url := urlHost + "/public/file/" + filename
	return url, nil
}

// CreateOauthRecord 创建一条统一验证记录
func CreateOauthRecord(stuId string, t time.Time, sid int) error {
	return d.SaveRecordSheet(ctx, dao.RecordSheet{StudentID: stuId, Time: t}, sid)
}

func safeCreateFile(tempFile string) (*os.File, error) {
	// 清理路径中的 ".."
	cleanedPath := filepath.Clean(tempFile)

	// 确保路径没有非法部分
	if strings.Contains(cleanedPath, "..") {
		return nil, fmt.Errorf("invalid file path: %s", cleanedPath)
	}

	f, err := os.Create(cleanedPath)
	if err != nil {
		return nil, err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			zap.L().Error("关闭文件失败", zap.Error(err))
		}
	}(f)

	// 进一步处理文件
	return f, nil
}

func safeOpenFile(src string) (*os.File, error) {
	// 获取文件的绝对路径
	absPath, err := filepath.Abs(src)
	if err != nil {
		return nil, err
	}

	// 清理路径，避免路径穿越
	cleanedPath := filepath.Clean(absPath)

	// 指定安全目录前缀
	safeDir := "/public/static/"

	// 使用 strings.HasPrefix 检查路径是否以安全目录前缀开始
	if !strings.HasPrefix(cleanedPath, safeDir) {
		return nil, fmt.Errorf("unsafe file path: %s", cleanedPath)
	}

	// 安全地打开文件
	srcFile, err := os.Open(cleanedPath)
	if err != nil {
		return nil, err
	}
	defer func(srcFile *os.File) {
		err := srcFile.Close()
		if err != nil {
			zap.L().Error("关闭文件失败", zap.Error(err))
		}
	}(srcFile)

	return srcFile, nil
}
