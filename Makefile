# 自定义模块的执行命令
.PHONY: all default help build build-windows build-macos clean fmt

all: help
default: help

##@ Help
help: ## 显示帮助信息，列出所有可用的目标命令。
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\n"} \
		/^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } \
		/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) }' $(MAKEFILE_LIST)

##@ Build
build: ## 直接编译项目
	go build -o QA main.go

build-windows: ## 在 Windows 上编译成 Linux 二进制文件
	SET CGO_ENABLE=0
	SET GOOS=linux
	SET GOARCH=amd64
	@echo "CGO_ENABLE=$(CGO_ENABLE) GOOS=$(GOOS) GOARCH=$(GOARCH)"
	go build -o QA main.go

build-macos: ## 在 macOS 上编译成 Linux 二进制文件
	CGO_ENABLE=0 GOOS=linux GOARCH=amd64 go build -o main main.go

clean: ## 清理编译生成的文件
	rm -f qa main.exe main

##@ Lint
fmt: ## 格式化代码并静态检查
	@echo "Formatting Go files..."
	gofmt -w .
	gci write . -s standard -s default
	@echo "Running Lints..."
	golangci-lint run