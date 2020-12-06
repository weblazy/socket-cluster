package master

import (
	"flag"
	websocket_cluster "websocket-cluster"

	"websocket-cluster/examples/router"
)

var (
	port2 = flag.Int64("port2", 9529, "the  port")
	host2 = flag.String("host2", "ws://localhost:9529", "the  host")
	path2 = flag.String("path2", "/p2", "the  path")
)

func Node2() {
	flag.Parse()
	websocket_cluster.StartNode(websocket_cluster.NewNodeConf(*host2, *path2, websocket_cluster.RedisConf{Addr: "127.0.0.1:6379", DB: 0}, []*websocket_cluster.RedisNode{&websocket_cluster.RedisNode{
		RedisConf: websocket_cluster.RedisConf{Addr: "127.0.0.1:6379", DB: 0},
		Position:  1,
	}}, onMsg).WithPort(*port2).WithRouter(router.Router))
}
