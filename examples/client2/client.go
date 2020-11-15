package main

import (
	"encoding/json"
	"time"
	websocket_cluster "websocket-cluster"

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

	err = conn.WriteJSON(map[string]interface{}{
		"message_type": "login",
		"data": map[string]interface{}{
			"password": "password2",
			"username": "admin2",
		},
	})
	if err != nil {
		logx.Info(err)
	}
	list := []map[string]interface{}{map[string]interface{}{"content": time.Now().Format("2006-01-02 15:04:05")}}
	err = conn.WriteJSON(map[string]interface{}{
		"message_type": "chat_message_list",
		"data": map[string]interface{}{
			"list":        list,
			"receive_uid": "1",
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
	messageMap := make(map[string]interface{})
	err := json.Unmarshal(message, &messageMap)
	if err != nil {
		logx.Info(err)
	}
	v1, ok := messageMap["message_type"]
	if !ok {
		logx.Info("message_type is nil")
	}
	switch v1 {
	case "auth":

	case "chat_message_list":
		data := messageMap["data"].(map[string]interface{})
		list := data["list"].([]interface{})
		messageIdList := make([]int64, 0)
		for k1 := range list {
			messageIdList = append(messageIdList, list[k1].(map[string]interface{})["id"].(int64))
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
