// plugins/plugin2.go
package plugins

import (
	"fmt"
	"time"

	"QA-System/internal/pkg/extension"
)

type Plugin2 struct{}

func (p *Plugin2) GetMetadata() extension.PluginMetadata {
	return extension.PluginMetadata{
		Name:        "plugin2",
		Version:     "1.0.0",
		Author:      "Author2",
		Description: "This is plugin 2",
	}
} //这是另一种写metaData的方式

func (p *Plugin2) Execute() error {
	// 插件的主要逻辑
	fmt.Println("Plugin1 executing at", time.Now().Format(time.RFC3339))
	// 这里可以添加插件的具体功能代码
	return nil
}

func init() {
	extension.RegisterPlugin(&Plugin2{})
}
