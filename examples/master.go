package main

import (
	websocket_cluster "websocket-cluster"
)

func main() {
	go websocket_cluster.StartMaster(
		websocket_cluster.NewMasterConf().WithPassword("admin"))
	node()
}
