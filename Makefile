# 自定义模块的执行命令
.PHONY: all
all: help

default: help

.PHONY: help
help: ## 显示帮助信息，列出所有可用的目标命令。
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Build
.PHONY: build
build: ## 直接编译项目
	go build -o QA main.go

.PHONY: build-windows
build-windows:	## 在windows上编译成linux二进制文件
	SET CGO_ENABLE=0
	SET GOOS=linux
	SET GOARCH=amd64
	@echo "CGO_ENABLE=" $(CGO_ENABLE) "GOOS=" $(GOOS) "GOARCH=" $(GOARCH)
	go build -o QA main.go

.PHONY: build-macos
build-macos:	## 在macos上编译成linux二进制文件
	GOOS=0 GOOS=linux GOARCH=amd64 go build -o main  main.go

##@ Lint
.PHONY: fmt
fmt: ## 格式化代码并静态检查
	@echo "Formatting Go files..."
	gofmt -w .
	gci write . -s standard -s default
	@echo "Running Lints..."
	golangci-lint run