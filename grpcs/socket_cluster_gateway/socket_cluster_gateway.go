package socket_cluster_gateway

import (
	"context"
	"log"

	"github.com/urfave/cli/v2"

	"github.com/weblazy/socket-cluster/grpcs/socket_cluster_gateway/config"
	"github.com/weblazy/socket-cluster/grpcs/socket_cluster_gateway/handler"
	"github.com/weblazy/socket-cluster/grpcs/socket_cluster_gateway/proto/gateway"
	"github.com/weblazy/socket-cluster/node"

	"github.com/weblazy/easy/closes"
	"github.com/weblazy/easy/econfig"
	"github.com/weblazy/easy/grpc/grpc_server"
)

var Cmd = &cli.Command{
	// Name:    "socket_cluster_gateway",
	// Aliases: []string{},
	// Usage:   "socket_cluster_gateway start",
	// Subcommands: []*cli.Command{
	// 	{
	// 		Name:   "start",
	// 		Usage:  "start service",
	// 		Action: Run,
	// 	},
	// },
}

func Run(c context.Context, n node.Node) error {
	defer closes.Close()
	econfig.InitGlobalViper(&config.Conf, config.LocalConfig)
	s := grpc_server.NewGrpcServer(config.Conf.GrpcServerConfig)
	gatewayService := handler.NewGatewayService(n)

	gateway.RegisterGatewayServiceServer(s, gatewayService)
	err := s.Init()
	if err != nil {
		log.Fatal(err)
	}
	err = s.Start()
	if err != nil {
		log.Fatal(err)
	}

	return nil
}
