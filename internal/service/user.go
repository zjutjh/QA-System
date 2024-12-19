package service

import (
	"QA-System/internal/dao"
	"QA-System/internal/models"
	"errors"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nfnt/resize"
	"image/jpeg"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func GetSurveyByID(id int) (*models.Survey, error) {
	survey, err := d.GetSurveyByID(ctx, id)
	return survey, err
}

func GetQuestionsBySurveyID(sid int) ([]models.Question, error) {
	var questions []models.Question
	questions, err := d.GetQuestionsBySurveyID(ctx, sid)
	return questions, err
}

func GetOptionsByQuestionID(questionId int) ([]models.Option, error) {
	var options []models.Option
	options, err := d.GetOptionsByQuestionID(ctx, questionId)
	return options, err
}

func GetQuestionByID(id int) (*models.Question, error) {
	var question *models.Question
	question, err := d.GetQuestionByID(ctx, id)
	return question, err
}

func SubmitSurvey(sid int, data []dao.QuestionsList, time string) error {
	var answerSheet dao.AnswerSheet
	answerSheet.SurveyID = sid
	answerSheet.Time = time
	answerSheet.Unique = true
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
		answer.SerialNum = q.SerialNum
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
			fmt.Println("删除临时目录失败: ", err)
		}
	}() // 在处理完之后删除临时目录及其中的文件
	// 在临时目录中创建临时文件
	tempFile := filepath.Join(tempDir, file.Filename)
	f, err := os.Create(tempFile)
	if err != nil {
		return "", errors.New("创建临时文件失败: " + err.Error())
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println("关闭文件失败: ", err)
		}
	}()
	// 将上传的文件保存到临时文件中
	src, err := file.Open()
	if err != nil {
		return "", errors.New("打开文件失败: " + err.Error())
	}
	defer func() {
		if err := src.Close(); err != nil {
			fmt.Println("关闭文件失败: ", err)
		}
	}()

	_, err = io.Copy(f, src)
	if err != nil {
		return "", errors.New("保存文件失败: " + err.Error())
	}
	// 判断文件的MIME类型是否为图片
	mime, err := mimetype.DetectFile(tempFile)
	if err != nil || !strings.HasPrefix(mime.String(), "image/") {
		return "", errors.New("文件类型不是图片: " + err.Error())
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

	//替换原始文件为压缩后的JPG文件
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
	srcImg, err := imaging.Open(srcPath)
	if err != nil {
		return err
	}

	// 调整图像大小（根据需要进行调整）
	resizedImg := resize.Resize(300, 0, srcImg, resize.Lanczos3)

	// 创建新的JPG文件
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// 以JPG格式保存调整大小的图像，并设置压缩质量为90
	err = jpeg.Encode(dstFile, resizedImg, &jpeg.Options{Quality: 90})
	if err != nil {
		return err
	}

	return nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	return nil
}

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

func CreateOauthRecord(stuId string, time time.Time, sid int) error {
	return d.SaveRecordSheet(ctx, dao.RecordSheet{StudentID: stuId, Time: time}, sid)
}
