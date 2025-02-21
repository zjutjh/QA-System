package plugins

import (
	"fmt"
	"time"

	"QA-System/pkg/extension"
)

const (
	pluginName  = "plugin1"          // PluginName插件名
	version     = "0.0.1"            // Version版本号（学习AWS的版本名谢谢）
	author      = "Author1"          // Author作者
	description = "This is plugin 1" // Descrtption 插件描述
)

// Plugin1 示例插件1的结构，要给manager的
type Plugin1 struct{}

// GetMetadata 获取插件的元数据
func (p *Plugin1) GetMetadata() extension.PluginMetadata {
	_ = p
	return extension.PluginMetadata{
		Name:        pluginName,
		Version:     version,
		Author:      author,
		Description: description,
	}
}

// Execute 插件的执行函数
func (p *Plugin1) Execute() error {
	_ = p
	// 插件的主要逻辑
	fmt.Println("Plugin1 executing at", time.Now().Format(time.RFC3339))
	// 这里可以添加插件的具体功能代码
	return nil
}

func init() {
	// 注册插件到默认管理器
	extension.GetDefaultManager().RegisterPlugin(&Plugin1{})
}
