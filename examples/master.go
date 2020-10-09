package main

import (
	"time"
	websocket_cluster "websocket-cluster"

	"github.com/gorilla/websocket"
	"github.com/weblazy/core/logx"
)

func main() {
	// go websocket_cluster.StartMaster(websocket_cluster.NewMasterConf())
	// node()
	client()
}

func client() {
	conn, _, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:9528/p1/client", nil)
	if err != nil {
		logx.Info("dial:", err)
	}

	auth := &websocket_cluster.Auth{
		Password:     "password",
		TransAddress: "ws://localhost:9528/client",
	}
	err = conn.WriteJSON(map[string]interface{}{
		"type": "auth",
		"data": auth,
	})
	if err != nil {
		logx.Info(err)
	}
	time.Sleep(10 * time.Minute)
}
