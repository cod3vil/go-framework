// 框架入口：基于 cobra 的 CLI。
//
//	server  启动 HTTP 服务（默认命令）
//	version 打印版本号
//
// P2 将新增 migrate（建表与种子数据）、create-admin（创建管理员）等子命令。
package main

import (
	"fmt"
	"os"

	"github.com/cod3vil/go-framework/internal/app"
	"github.com/spf13/cobra"
)

var configPath string

func main() {
	root := &cobra.Command{
		Use:   "server",
		Short: "go-framework 企业级应用框架",
		// 不带子命令时默认启动服务。
		RunE: func(cmd *cobra.Command, args []string) error { return runServer() },
	}
	root.PersistentFlags().StringVarP(&configPath, "config", "c", "configs/config.yaml", "配置文件路径")

	root.AddCommand(
		&cobra.Command{
			Use:   "serve",
			Short: "启动 HTTP 服务",
			RunE:  func(cmd *cobra.Command, args []string) error { return runServer() },
		},
		&cobra.Command{
			Use:   "version",
			Short: "打印版本号",
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Println(app.Version)
			},
		},
	)

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runServer() error {
	a, err := app.New(configPath)
	if err != nil {
		return err
	}
	return a.Run()
}
