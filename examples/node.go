package main

import (
	"flag"
	websocket_cluster "websocket-cluster"

	"github.com/weblazy/core/database/redis"
)

var (
	port          = flag.Int64("port", 9528, "the  port")
	host          = flag.String("host", "ws://localhost:9528", "the  host")
	path          = flag.String("path", "/p1", "the  path")
	masterAddress = flag.String("masterAddress", "ws://127.0.0.1:9527/master", "the  masterAddress")
)

func node() {
	flag.Parse()
	websocket_cluster.StartNode(websocket_cluster.NewNodeConf(*host, *path, *masterAddress, redis.RedisConf{Host: "127.0.0.1:6379", Type: "node"}).WithPort(*port))
}
