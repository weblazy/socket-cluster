package main

import (
	websocket_cluster "websocket-cluster"
)

func main() {
	go websocket_cluster.StartMaster(
		websocket_cluster.MasterConf{
			SocketConf: &websocket_cluster.SocketConfig{
				Port: 9090,
			}})
	node()
}
