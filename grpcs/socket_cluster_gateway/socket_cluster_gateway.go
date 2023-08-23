package socket_cluster_gateway

import (
	"context"
	"log"

	"github.com/urfave/cli/v2"

	"github.com/weblazy/socket-cluster/grpcs/socket_cluster_gateway/handler"
	"github.com/weblazy/socket-cluster/grpcs/socket_cluster_gateway/proto/gateway"
	"github.com/weblazy/socket-cluster/node"

	"github.com/weblazy/easy/grpc/grpc_server"
	"github.com/weblazy/easy/grpc/grpc_server/grpc_server_config"
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

func Run(c context.Context, n node.Node, grpcConf *grpc_server_config.Config) error {
	// defer closes.Close()
	s := grpc_server.NewGrpcServer(grpcConf)
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
