package main

import (
	websocket_cluster "websocket-cluster"

	"github.com/weblazy/core/database/redis"
)

func node() {
	websocket_cluster.StartNode(websocket_cluster.NewNodeConf().
		WithMasterAddress("127.0.0.1:9090").
		WithPassword("skdss").
		WithTransConf(websocket_cluster.SocketConfig{Port: 8080}).
		WithClientConf(websocket_cluster.SocketConfig{Port: 8081}).
		WithRedisConf(redis.RedisConf{Host: "127.0.0.1:6379", Type: "node"}))
}
