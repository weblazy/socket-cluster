package config

import (
	"github.com/weblazy/easy/grpc/grpc_server/grpc_server_config"
)

type Config struct {
	BaseConfig       struct{}
	GrpcServerConfig *grpc_server_config.Config
}

var Conf = Config{
	BaseConfig:       struct{}{},
	GrpcServerConfig: grpc_server_config.DefaultConfig(),
}

var LocalConfig = ""
