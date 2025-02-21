# QA-System
## 项目说明
一个用 go 写成的问卷后端服务项目

使用技术栈有gin、gorm、session、mongodb、viper、asynq、zap

### 项目目录
```
QA-System/
├── main.go                   # 应用程序的入口点，包含启动服务器的代码
├── conf                      # 存放配置文件，如 YAML、JSON 格式的配置
├── docs                      # 项目文档，可能包括 API 文档、开发者指南等
│   └── README.md             # 项目 README 文档
├── go.mod                    # Go Modules 模块依赖文件
├── go.sum                    # Go Modules 模块依赖的校验和
├── hack                      # 构建脚本、CI 配置和辅助工具
│   └── docker                # Docker 相关配置，如 Dockerfile 和 Docker Compose 文件
├── internal                  # 项目内部包，包含服务器、模型、配置等
│   ├── global                # 全局可用的配置和初始化代码
│   │   └── config            # 配置加载和解析
│   ├── router                # 路由注册，定义应用程序的路由结构
│   ├── middleware            # 中间件逻辑，处理跨请求的任务如日志、认证等
│   ├── handle                # 放置 handler 函数，可能用于处理具体的业务逻辑
│   ├── service               # 业务逻辑服务层，实现应用程序的核心业务逻辑
│   ├── dao                   # 数据访问对象层，与数据库交互，执行增删改查等操作(包含部分测试代码)
│   ├── model                 # 数据模型定义
│   │   ├── user.go           # 用户模型定义
│   │   └── ...
│   └── pkg                  # 内部工具包
│       ├── code             # 错误码定义，用于标准化错误处理
│       ├── utils            # 内部使用的工具函数
│       ├── log              # 日志配置和管理，封装日志记录器的配置和使用
│       ├── database         # 数据库连接和初始化，管理数据库连接池
│       ├── session          # 会话管理，处理用户会话和状态
│       └── redis            # Redis 配置和管理，封装 Redis 缓存操作
├── logs                     # 日志文件输出目录，存放应用程序生成的日志文件
├── LICENSE                   # 项目许可证文件
├── Makefile                  # 根 Makefile 文件，包含构建和编译项目的指令
├── pkg                       # 可被外部引用的全局工具包
│   └── util                  # 通用工具代码
├── README.md                  # 项目 README 文档，通常提供项目概览和快速开始指南
├── public                    # 公共静态资源，如未构建的前端资源或可直接访问的静态文件
└── .gitignore                # Git 忽略文件配置
```

### 如何运行
1. 克隆该项目
```sh
git clone https://github.com/zjutjh/QA-System
```
2. 更改配置文件
```sh
mv conf/config.yaml.example conf/config.yaml
```
3. 由于文件是本地存放,因此要创建文件存放的目录并给予权限
```sh
mkdir public
chmod -R 755 ./public
```
4. 
* 本地运行后端程序
```sh
go run main.go
```
* 打包成可执行文件
```sh
#### Windows(cmd)
SET CGO_ENABLE=0
SET GOOS=linux
SET GOARCH=amd64
make build-linux ### 或go build -o QA main.go

#### linux
make build
```
5. 测试代码
第一次先安装[golangci-lint](https://github.com/golangci/golangci-lint/releases)

由于部分功能要求使用diff，因此windows环境下推荐到git bash下执行。
``` bash
golangci-lint run
```