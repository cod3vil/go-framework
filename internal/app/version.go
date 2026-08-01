package app

// Version 构建版本号，由 Makefile 通过 ldflags 注入：
//
//	-ldflags "-X github.com/cod3vil/go-framework/internal/app.Version=v0.1.0"
var Version = "dev"
