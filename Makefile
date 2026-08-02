APP      := server
MODULE   := github.com/cod3vil/go-framework
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS  := -s -w -X $(MODULE)/internal/app.Version=$(VERSION)
GOFLAGS  := CGO_ENABLED=0

.PHONY: dev build run web web-dev tidy fmt vet lint test clean

## dev: 以开发模式启动服务
dev:
	go run ./cmd/server

## web: 构建管理后台前端到 web/dist（供 go:embed）
web:
	cd web && pnpm install && pnpm build

## web-dev: 前端开发服务器（代理到本地 Go 服务）
web-dev:
	cd web && pnpm dev

## build: 编译二进制到 bin/（前端已 embed，如需重建前端先执行 make web）
build:
	$(GOFLAGS) go build -trimpath -ldflags '$(LDFLAGS)' -o bin/$(APP) ./cmd/server

## all: 构建前端并编译单二进制
all: web build

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
