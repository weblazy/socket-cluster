package main

import (
	"context"
	"os"

	"github.com/weblazy/socket-cluster/grpcs/socket_cluster_gateway"

	"github.com/urfave/cli/v2"
	"github.com/weblazy/easy/elog"
	"github.com/weblazy/easy/print"
)

const (
	ProjectName    = "socket_cluster"
	ProjectVersion = "v1.0.0"
)

func main() {
	// 打印Banner
	print.PrintBanner(ProjectName)
	// 配置cli参数
	cliApp := cli.NewApp()
	cliApp.Name = ProjectName
	cliApp.Version = ProjectVersion

	// 指定命令运行的函数
	cliApp.Commands = []*cli.Command{
		socket_cluster_gateway.Cmd,
	}

	// 启动cli
	if err := cliApp.Run(os.Args); err != nil {
		elog.ErrorCtx(context.Background(), "Failed to start application", elog.FieldError(err))
	}
}
