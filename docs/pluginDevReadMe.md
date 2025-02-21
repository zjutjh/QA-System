# 插件开发相关

- [ ] 很多组建可以使用开源实现,比如TTLCache和gopool,降低项目的复杂程度
- [x] 多用面向对象的方法,例如manager
- [ ] 插件不需要以goroutine的方式一直存在,在异步消费的时候再去执行就行了

## v0.0.1  

基本想法：internal/pkg/extension包负责所有插件的管理、调度  
把所有插件作为 plugins 包的一部分，在main函数启动时隐性导入。通过每个插件自己的init函数，所有的插件将自己注册到manager提供的管理器内部。

manager再根据配置文件里写的顺序依次加载、调用插件

### 基本思想

1. **插件接口**：定义一个标准的插件接口，确保所有插件都遵循相同的规范。
2. **插件注册**：通过 `init` 函数在程序启动时自动注册插件。
3. **配置管理**：使用配置文件决定哪些插件需要加载及其顺序。
4. **插件执行**：在主程序中调用插件管理器来加载和执行插件。

### 步骤概述

1. **定义插件接口**
2. **编写具体插件**
3. **配置文件管理**

---

### 1. 定义插件接口

首先，我们查看定义插件接口 `Plugin` 和元数据结构 `PluginMetadata`。这些定义放在 `pkg/extension/interface.go` 文件中。你可以直接参考实例插件plugin1和plugin2来着。

```go
// pkg/extension/interface.go
package extension

// PluginMetadata 插件元数据
type PluginMetadata struct {
    Name        string // 插件名称
    Version     string // 插件版本
    Author      string // 插件作者
    Description string // 插件描述
}

// Plugin 定义插件接口
type Plugin interface {
    GetMetadata() PluginMetadata                 // GetMetadata 获取插件元数据
    Execute() error // Execute 执行插件具体的功能
}
```

### 2. 编写具体插件

每个插件都需要实现 `extension.Plugin` 接口，并通过 `init` 函数在程序启动时注册自己。以下是两个示例插件 `plugin1.go` 和 `plugin2.go`。

示例插件1: `plugin1.go`

```go
package plugins

import (
    "QA-System/pkg/extension"
    "fmt"
)

const (
    PluginName  = "plugin1"
    Version     = "1.0.0"
    Author      = "Author1"
    Description = "This is plugin 1"
)

type Plugin1 struct{}

func (p *Plugin1) GetMetadata() extension.PluginMetadata {
    return extension.PluginMetadata{
        Name:        PluginName,
        Version:     Version,
        Author:      Author,
        Description: Description,
    }
}

func (p *Plugin1) Execute() error {
    fmt.Println("Plugin1 executing with params:", params)
    params["processed_by"] = "plugin1"
    return nil
}

func init() {
    // 注册插件到默认管理器
    extension.GetDefaultManager().RegisterPlugin(&Plugin1{})
}
```

示例插件2: `plugin2.go`

```go
// plugins/plugin2.go
package plugins

import (
    "QA-System/pkg/extension"
    "fmt"
)

type Plugin2 struct{}

func (p *Plugin2) GetMetadata() extension.PluginMetadata {
    return extension.PluginMetadata{
        Name:        "plugin2",
        Version:     "1.0.0",
        Author:      "Author2",
        Description: "This is plugin 2",
    }
}

func (p *Plugin2) Execute() error {
    fmt.Println("Plugin2 executing with params:", params)
    params["processed_by"] = "plugin2"
    return nil
}

func init() {
    // 注册插件到默认管理器
    extension.GetDefaultManager().RegisterPlugin(&Plugin2{})
}
```

### 3. 配置文件管理

编辑conf目录下的配置文件 `config.yaml` :

```yaml
# conf/config.yaml
plugins:
  order:
    - "plugin1"
    - "plugin2"
```

**一定要加引号啊！不然可能读取不到** 
上述的顺序就是先加载```plugin1```再加载```plugin2```

### 4. 额外的插件所需

如果你的插件需要其他东西：比如从主程序传入的参数等等，请自行修改主程序传入参数；
如果插件需要配置，请在```config.yaml```文件中加上去然后配置读取方法

### To Do Next

- [ ] 真动态加载
- [ ] 二进制文件so和dll的支持
- [ ] 插件生命周期管控
- [ ] 更好的微架构融入
- [ ] 性能优化

### 插件管理器的 OOP优化

关键的有待改进点说明

|改进点|原实现问题|改进方案优势
|----|----|----|
|直接依赖具体日志库|耦合zap实现，难以替换日志系统|通过接口抽象，支持任意日志实现|
|配置硬编码|直接调用全局config.Config|依赖接口，方便测试和扩展|
|大而全的接口|接口方法过多违反ISP原则|拆分细粒度接口，按需组合|
|缺乏生命周期管理|没有明确的初始化/销毁方法|添加Start()/Stop()等生命周期方法|