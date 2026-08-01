APP      := server
MODULE   := github.com/cod3vil/go-framework
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS  := -s -w -X $(MODULE)/internal/app.Version=$(VERSION)
GOFLAGS  := CGO_ENABLED=0

.PHONY: dev build run tidy fmt vet lint test clean

## dev: 以开发模式启动服务
dev:
	go run ./cmd/server

## build: 编译二进制到 bin/
build:
	$(GOFLAGS) go build -trimpath -ldflags '$(LDFLAGS)' -o bin/$(APP) ./cmd/server

## run: 编译并运行
run: build
	./bin/$(APP)

tidy:
	go mod tidy

fmt:
	gofmt -w .

vet:
	go vet ./...

test:
	go test ./...

clean:
	rm -rf bin data logs
