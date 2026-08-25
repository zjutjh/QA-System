package oss

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// DownloadFile 拉取对象文件内容
func DownloadFile(objectKey string) (io.ReadCloser, error) {
	url := Client.GetFileURL(objectKey, false)
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("下载文件失败: %s", resp.Status)
	}
	return resp.Body, nil
}
