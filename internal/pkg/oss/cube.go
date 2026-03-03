package oss

import "github.com/zjutjh/WeJH-SDK/cube"

// Client 对象存储客户端
var Client *cube.Client

// Init 显式初始化 OSS 客户端。必须在使用 Client 前调用。
func Init() {
	Client = cube.New(getConfig())
}
