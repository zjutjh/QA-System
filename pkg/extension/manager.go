// extension/manager.go
package extension

import (
	"fmt"
	"sync"

	"QA-System/internal/global/config"

	"go.uber.org/zap"
)

// extension包自己的init函数，用来看一眼extension是不是被导入了
func init() {
	defaultManager = NewPluginManager(zap.NewNop())
	fmt.Println("插件包加载模块初始化成功 阶梯计划成功")
}

// PluginManager 插件管理器类
type PluginManager struct {
	plugins map[string]Plugin
	mu      sync.Mutex
	logger  *zap.Logger
}

// PluginManagerInterface 插件管理器接口，定义了插件管理器的方法（在这集合一下而已）
type PluginManagerInterface interface {
	RegisterPlugin(p Plugin)
	GetPlugin(name string) (Plugin, bool)
	LoadPlugins() ([]Plugin, error)
	ExecutePlugins() error
}

// NewPluginManager 创建一个新的插件管理器实例，可以用自己的，也可以手动指定（真的需要吗？？？）
func NewPluginManager(logger *zap.Logger) *PluginManager {
	return &PluginManager{
		plugins: make(map[string]Plugin),
		logger:  logger,
	}
}

// var (
// 	plugins = make(map[string]Plugin) // 插件名称 -> 插件实例
// 	mu      sync.Mutex
// )

// RegisterPlugin 向插件管理器注册插件
func (pm *PluginManager) RegisterPlugin(p Plugin) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	metadata := p.GetMetadata()
	pm.plugins[metadata.Name] = p
	zap.L().Info("Plugin registered successfully",
		zap.String("name", metadata.Name),
		zap.String("version", metadata.Version))
}

// GetPlugin 从已经注册到插件管理器中的插件集合里获取特定的插件实例
func (pm *PluginManager) GetPlugin(name string) (Plugin, bool) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	p, ok := pm.plugins[name]
	return p, ok
}

// LoadPlugins 从配置文件中加载插件并返回插件实例列表
func (pm *PluginManager) LoadPlugins() ([]Plugin, error) {
	pluginNames := config.Config.GetStringSlice("plugins.order")
	zap.L().Info("Loading plugins from config",
		zap.Strings("plugin_names", pluginNames))
	var pluginList []Plugin

	for _, name := range pluginNames {
		p, ok := pm.GetPlugin(name)
		if !ok {
			return nil, fmt.Errorf("plugin %s not found", name)
		}
		pluginList = append(pluginList, p)
	}

	return pluginList, nil
}

// ExecutePlugins 依次执行插件链
func (pm *PluginManager) ExecutePlugins() error {
	pluginList, err := pm.LoadPlugins()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for _, p := range pluginList {
		wg.Add(1)
		go func(plugin Plugin) {
			defer wg.Done()
			metadata := plugin.GetMetadata()
			pm.logger.Info("Starting plugin service",
				zap.String("name", metadata.Name),
				zap.String("version", metadata.Version))

			if err := plugin.Execute(); err != nil {
				pm.logger.Error("Plugin service failed",
					zap.String("name", metadata.Name),
					zap.Error(err))
			}
		}(p)
	}
	wg.Wait()

	return nil
}

// 默认插件管理器实例
var defaultManager *PluginManager

// GetDefaultManager 获取默认的插件管理器实例
func GetDefaultManager() *PluginManager {
	return defaultManager
}
