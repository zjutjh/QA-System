package bootstrap

import (
	global "QA-System/internal/global/config"
	"QA-System/pkg/extension"

	"go.uber.org/zap"
)

// InitPlugins 初始化插件管理器，加载并执行配置文件中声明的插件。
// 插件加载或执行失败仅记录日志，不中断主流程。
func InitPlugins() {
	pm := extension.GetDefaultManager()
	pluginNames := global.Config.GetStringSlice("plugins.order")

	plugins, err := pm.LoadPlugins(pluginNames)
	if err != nil {
		zap.L().Error("Error loading plugins", zap.Error(err))
	}

	// 打印插件状态信息
	for _, plugin := range plugins {
		metadata := plugin.GetMetadata()
		status, healthy := extension.GetPluginStatus(metadata.Name)
		if healthy {
			zap.L().Info("Plugin loaded successfully",
				zap.String("name", metadata.Name),
				zap.String("version", metadata.Version),
				zap.String("status", status))
		} else {
			zap.L().Warn("Plugin loaded but unhealthy",
				zap.String("name", metadata.Name),
				zap.String("version", metadata.Version),
				zap.String("status", status))
		}
	}

	if err := pm.ExecutePluginList(pluginNames); err != nil {
		zap.L().Error("Error executing plugins", zap.Error(err))
	}
}
