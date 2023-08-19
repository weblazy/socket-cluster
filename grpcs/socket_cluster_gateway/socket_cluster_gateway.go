package socket_cluster_gateway

import (
	"log"

	"github.com/urfave/cli/v2"

	"socket_cluster/grpcs/socket_cluster_gateway/config"
	"socket_cluster/grpcs/socket_cluster_gateway/handler"
	"socket_cluster/grpcs/socket_cluster_gateway/proto/gateway"

	"github.com/weblazy/easy/closes"
	"github.com/weblazy/easy/econfig"
	"github.com/weblazy/easy/grpc/grpc_server"
)

var Cmd = &cli.Command{
	Name:    "socket_cluster_gateway",
	Aliases: []string{},
	Usage:   "socket_cluster_gateway start",
	Subcommands: []*cli.Command{
		{
			Name:   "start",
			Usage:  "start service",
			Action: Run,
		},
	},
}

func Run(c *cli.Context) error {
	defer closes.Close()
	econfig.InitGlobalViper(&config.Conf, config.LocalConfig)
	s := grpc_server.NewGrpcServer(config.Conf.GrpcServerConfig)
	gatewayService := handler.NewGatewayService()

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
