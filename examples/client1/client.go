package main

import (
	"encoding/json"

	websocket_cluster "websocket-cluster"

	"github.com/gorilla/websocket"
	"github.com/weblazy/core/logx"
)

func main() {
	client()
}

func client() {
	conn, _, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:9528/p1/client", nil)
	if err != nil {
		logx.Info("dial:", err)
	}

	err = conn.WriteJSON(map[string]interface{}{
		"message_type": "login",
		"data": map[string]interface{}{
			"password": "password1",
			"username": "admin1",
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
		OnClientMessage(&websocket_cluster.Connection{Conn: conn}, message)
	}
}

func OnClientMessage(conn *websocket_cluster.Connection, message []byte) {
	data := make(map[string]interface{})
	err := json.Unmarshal(message, &data)
	if err != nil {
		logx.Info(err)
	}
	v1, ok := data["message_type"]
	if !ok {
		logx.Info("message_type is nil")
	}
	switch v1 {
	case "auth":

	case "chat_message_list":
		data := data["data"].(map[string]interface{})
		list := data["list"].([]interface{})
		messageIdList := make([]int64, 0)
		for k1 := range list {
			messageIdList = append(messageIdList, int64(list[k1].(map[string]interface{})["id"].(float64)))
		}
		err = conn.WriteJSON(map[string]interface{}{
			"message_type": "ackReceive",
			"data": map[string]interface{}{
				"message_id_list": messageIdList,
			},
		})
		if err != nil {
			logx.Info(err)
		}
	}
}
