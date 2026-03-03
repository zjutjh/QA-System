package extension

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
)

// 默认单例实例插件管理器实例
var (
	defaultManager     *PluginManager
	defaultManagerOnce sync.Once
)

// extension包自己的init函数，初始化默认插件管理器实例
func init() {
	defaultManager = NewPluginManager(slog.Default())
}

// PluginManager 插件管理器类，实现了PluginManagerInterface接口
type PluginManager struct {
	plugins map[string]Plugin
	mu      sync.Mutex
	logger  *slog.Logger
}

// PluginManagerInterface 插件管理器接口
type PluginManagerInterface interface {
	RegisterPlugin(p Plugin)
	GetPlugin(name string) (Plugin, bool)
	LoadPlugins(pluginNames []string) ([]Plugin, error)
	ExecutePlugin(name string, params map[string]any) error
	ExecutePluginList(pluginNames []string) error
}

// NewPluginManager 创建一个新的插件管理器实例
func NewPluginManager(logger *slog.Logger) *PluginManager {
	return &PluginManager{
		plugins: make(map[string]Plugin),
		logger:  logger,
	}
}

// GetDefaultManager 获取默认的插件管理器单例实例，确保线程安全的初始化
func GetDefaultManager() *PluginManager {
	defaultManagerOnce.Do(func() {
		if defaultManager == nil {
			defaultManager = NewPluginManager(GetPluginLogger())
		}
	})
	return defaultManager
}

// RegisterPlugin 向插件管理器注册插件
func (pm *PluginManager) RegisterPlugin(p Plugin) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	metadata := p.GetMetadata()

	// 插件元数据有效性检查
	if metadata.Name == "" || metadata.Version == "" {
		pm.logger.Error("invalid plugin metadata: name or version is empty")
		return errors.New("invalid plugin metadata: name or version is empty")
	}

	// 插件名称唯一性检查
	if _, exists := pm.plugins[metadata.Name]; exists {
		pm.logger.Error("Duplicated Plugin! The former won't be effective",
			"name", metadata.Name)
		return errors.New("plugin duplicated, please check the config file")
	}

	pm.plugins[metadata.Name] = p

	pm.logger.Info("Plugin registered successfully",
		"name", metadata.Name,
		"version", metadata.Version)
	return nil
}

// ExecutePlugin 执行特定插件
func (pm *PluginManager) ExecutePlugin(name string, params map[string]any) error {
	plugin, ok := pm.GetPlugin(name)
	if !ok {
		pm.logger.Warn("Plugin not found, skipping execution", "name", name)
		return fmt.Errorf("plugin %s not found", name)
	}

	metadata := plugin.GetMetadata()

	// 如果插件实现了健康检查接口，先检查健康状态
	if healthChecker, ok := plugin.(PluginHealthChecker); ok {
		if !healthChecker.IsHealthy() {
			pm.logger.Warn("Plugin is unhealthy, skipping execution",
				"name", metadata.Name,
				"status", healthChecker.GetStatus())
			return fmt.Errorf("plugin %s is unhealthy: %s", name, healthChecker.GetStatus())
		}
	}

	pm.logger.Info("Executing specific plugin",
		"name", metadata.Name,
		"version", metadata.Version)

	if err := plugin.Execute(params); err != nil {
		pm.logger.Error("Plugin execution failed",
			"name", metadata.Name,
			"error", err)
		return err
	}

	return nil
}

// ExecutePluginSafely 安全执行插件，不会因为插件失败而中断程序
func (pm *PluginManager) ExecutePluginSafely(name string, params map[string]any) {
	if err := pm.ExecutePlugin(name, params); err != nil {
		pm.logger.Warn("Plugin execution failed, but continuing",
			"name", name,
			"error", err)
	}
}

// ExecutePluginList 链式执行插件列表（这个功能还在调）
func (pm *PluginManager) ExecutePluginList(pluginNames []string) error {
	pluginList, err := pm.LoadPlugins(pluginNames)
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
				"name", metadata.Name,
				"version", metadata.Version)

			// 每个插件使用空参数执行
			if err := plugin.Execute(nil); err != nil {
				pm.logger.Error("Plugin service failed",
					"name", metadata.Name,
					"error", err)
			}
		}(p)
	}
	wg.Wait()

	return nil
}

// GetPlugin 从已经注册到插件管理器中的插件集合里获取特定的插件实例
func (pm *PluginManager) GetPlugin(name string) (Plugin, bool) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	p, ok := pm.plugins[name]
	return p, ok
}

// LoadPlugins 按照给定的插件名称顺序加载已注册的插件并返回实例列表。
// pluginNames 为空时返回空列表。
func (pm *PluginManager) LoadPlugins(pluginNames []string) ([]Plugin, error) {
	pm.logger.Info("Detecting plugins",
		"plugin_names", pluginNames)
	pluginList := make([]Plugin, 0, len(pluginNames))

	for _, name := range pluginNames {
		p, ok := pm.GetPlugin(name)
		if !ok {
			return nil, fmt.Errorf("plugin %s not found", name)
		}
		pluginList = append(pluginList, p)
	}

	return pluginList, nil
}

// 包级别便捷函数

// RegisterPlugin （包级）向默认插件管理器注册插件
func RegisterPlugin(p Plugin) error {
	return GetDefaultManager().RegisterPlugin(p)
}

// ExecutePluginSafely （包级）安全执行特定插件
func ExecutePluginSafely(name string, params map[string]any) {
	GetDefaultManager().ExecutePluginSafely(name, params)
}

// GetPluginStatus （包级）获取插件状态信息
func GetPluginStatus(name string) (string, bool) {
	plugin, ok := GetDefaultManager().GetPlugin(name)
	if !ok {
		return "not found", false
	}

	if healthChecker, ok := plugin.(PluginHealthChecker); ok {
		return healthChecker.GetStatus(), healthChecker.IsHealthy()
	}

	return "no health check available", true
}
