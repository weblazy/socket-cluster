package server

import (
	"context"

	"github.com/weblazy/socket-cluster/grpcs/socket_cluster_gateway/handler"
	"github.com/weblazy/socket-cluster/node"
	"go.uber.org/zap"

	"github.com/weblazy/easy/elog"
	"github.com/weblazy/easy/grpc/grpc_server"
	"github.com/weblazy/easy/grpc/grpc_server/grpc_server_config"
	"github.com/weblazy/socket-cluster/grpcs/socket_cluster_gateway/proto/gateway"
)

func Run(ctx context.Context, nodeConf *node.NodeConf) {
	serverNode, err := node.NewNode(nodeConf)
	if err != nil {
		elog.ErrorCtx(ctx, "msg", zap.Error(err))
	}
	cfg := grpc_server_config.DefaultConfig()
	server := grpc_server.NewGrpcServer(cfg)
	gateway.RegisterGatewayServiceServer(server.Server, handler.NewGatewayService(serverNode))
	err = server.Init()
	if err != nil {
		elog.ErrorCtx(ctx, "server.Init", zap.Error(err))
	}
	err = server.Start()
	if err != nil {
		elog.ErrorCtx(ctx, "server.Start", zap.Error(err))
	}
}