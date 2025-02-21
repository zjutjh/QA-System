package extension

// 定义一个插件长啥样

// PluginMetadata 插件元数据
type PluginMetadata struct {
	Name        string // 插件名称
	Version     string // 插件版本
	Author      string // 插件作者
	Description string // 插件描述
}

// Plugin 定义插件接口
type Plugin interface {
	GetMetadata() PluginMetadata // GetMetadata 获取插件元数据
	Execute() error              // Execute 执行插件功能
}
