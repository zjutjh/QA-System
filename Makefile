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

.PHONY: build build-linux