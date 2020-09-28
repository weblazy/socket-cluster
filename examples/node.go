package main

import (
	websocket_cluster "websocket-cluster"

	"github.com/weblazy/core/database/redis"
)

func node() {
	websocket_cluster.StartNode(websocket_cluster.NodeConf{
		TransConf: websocket_cluster.SocketConfig{
			Ip:   "127.0.0.1",
			Port: 8080,
		},
		ClientConf: websocket_cluster.SocketConfig{
			Ip:   "127.0.0.1",
			Port: 8080,
		},
		RedisConf: redis.RedisConf{
			Host: "127.0.0.1:6379",
			Type: "node",
		},
		MasterAddress: "127.0.0.1:9090",
		Password:      "skdss",
	})
}
