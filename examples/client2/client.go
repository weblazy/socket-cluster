package main

import (
	"github.com/gorilla/websocket"
	"github.com/weblazy/core/logx"
)

func main() {
	client()
}

func client() {
	conn, _, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:9529/p2/client", nil)
	if err != nil {
		logx.Info("dial:", err)
	}

	// auth := &websocket_cluster.Auth{
	// 	Password:     "password",
	// 	TransAddress: "ws://localhost:9528/client",
	// }
	err = conn.WriteJSON(map[string]interface{}{
		"type": "auth",
		"data": map[string]interface{}{
			"password":      "password",
			"trans_address": "ws://localhost:9528/client",
			"uid":           "457",
		},
	})
	if err != nil {
		logx.Info(err)
	}
	err = conn.WriteJSON(map[string]interface{}{
		"type": "send",
		"data": map[string]interface{}{
			"password":      "password",
			"trans_address": "ws://localhost:9528/client",
			"uid":           "456",
		},
	})
	if err != nil {
		logx.Info(err)
	}
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logx.Info(err)
			}
			break
		}
		logx.Info(string(message))
	}
}
