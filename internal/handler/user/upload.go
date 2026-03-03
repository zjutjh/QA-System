package user

import (
	"errors"
	"mime/multipart"
	"net/http"

	"QA-System/internal/pkg/code"
	"QA-System/internal/pkg/log"
	"QA-System/internal/pkg/oss"
	"QA-System/internal/pkg/utils"

	"github.com/dustin/go-humanize"
	"github.com/gin-gonic/gin"
	"github.com/zjutjh/WeJH-SDK/cube"
	"go.uber.org/zap"
)

type uploadImgData struct {
	Img *multipart.FileHeader `form:"img" binding:"required"`
}

// UploadImg 上传图片
func UploadImg(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 10*humanize.MiByte)

	var data uploadImgData
	if err := c.ShouldBind(&data); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			code.AbortWithException(c, code.FileSizeError, err)
			return
		}
		code.AbortWithException(c, code.ParamError, err)
		return
	}

	fileHeader := data.Img
	file, err := fileHeader.Open()
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			zap.L().Error("Failed to close file", zap.Error(err))
		}
	}(file)

	resp, err := oss.Client.UploadFile(fileHeader.Filename, file, "img", true, true)
	if errors.Is(err, cube.ErrRequestBizCodeNotOK) {
		// 直接使用 Cube 的错误信息
		code.AbortWithException(c, code.NewError(resp.Code, log.LevelError, resp.Msg), err)
		return
	}
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}

	url := oss.Client.GetFileURL(resp.Data.ObjectKey, false)
	utils.JsonSuccessResponse(c, url)
}

type uploadFileData struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}

// UploadFile 上传文件
func UploadFile(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 50*humanize.MiByte)

	var data uploadFileData
	if err := c.ShouldBind(&data); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			code.AbortWithException(c, code.FileSizeError, err)
			return
		}
		code.AbortWithException(c, code.ParamError, err)
		return
	}

	fileHeader := data.File
	file, err := fileHeader.Open()
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			zap.L().Error("Failed to close file", zap.Error(err))
		}
	}(file)

	resp, err := oss.Client.UploadFile(fileHeader.Filename, file, "file", false, true)
	if errors.Is(err, cube.ErrRequestBizCodeNotOK) {
		// 直接使用 Cube 的错误信息
		code.AbortWithException(c, code.NewError(resp.Code, log.LevelError, resp.Msg), err)
		return
	}
	if err != nil {
		code.AbortWithException(c, code.ServerError, err)
		return
	}

	url := oss.Client.GetFileURL(resp.Data.ObjectKey, false)
	utils.JsonSuccessResponse(c, url)
}
