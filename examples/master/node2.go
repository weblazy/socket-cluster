package main

import (
	"flag"
	websocket_cluster "websocket-cluster"

	"github.com/weblazy/core/database/redis"
)

var (
	port2 = flag.Int64("port2", 9529, "the  port")
	host2 = flag.String("host2", "ws://localhost:9529", "the  host")
	path2 = flag.String("path2", "/p2", "the  path")
)

func node2() {
	flag.Parse()
	websocket_cluster.StartNode(websocket_cluster.NewNodeConf(*host2, *path2, *masterAddress, redis.RedisConf{Host: "127.0.0.1:6379", Type: "node"}, []*websocket_cluster.RedisNode{&websocket_cluster.RedisNode{
		RedisConf: redis.RedisConf{Host: "127.0.0.1:6379", Type: "node"},
		Position:  1,
	}}, onMsg).WithPort(*port2).WithRouter(Router))
}
