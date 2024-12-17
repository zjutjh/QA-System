build:
	go build -o QA main.go

build-windows:
	SET CGO_ENABLE=0
	SET GOOS=linux
	SET GOARCH=amd64
	@echo "CGO_ENABLE=" $(CGO_ENABLE) "GOOS=" $(GOOS) "GOARCH=" $(GOARCH)
	go build -o QA main.go

build-macos:
	GOOS=0 GOOS=linux GOARCH=amd64 go build -o main  main.go


# 格式化代码并检查风格
fmt:
	@echo "Formatting Go files..."
	gofmt -w .
	gci write . -s standard -s default
	@echo "Running Lints..."
	golangci-lint run