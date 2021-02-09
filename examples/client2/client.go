package main

import (
	"encoding/json"
	"time"

	socket_cluster "github.com/weblazy/socket-cluster"

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
		"msg_type": "login",
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
		"msg_type": "chat_msg_list",
		"data": map[string]interface{}{
			"list":        list,
			"receive_uid": "1",
		},
	})
	if err != nil {
		logx.Info(err)
	}
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logx.Info(err)
			}
			break
		}
		logx.Info(string(msg))
		OnClientMsg(&socket_cluster.Connection{Conn: conn}, msg)
	}
}

func OnClientMsg(conn *socket_cluster.Connection, msg []byte) {
	msgMap := make(map[string]interface{})
	err := json.Unmarshal(msg, &msgMap)
	if err != nil {
		logx.Info(err)
	}
	v1, ok := msgMap["msg_type"]
	if !ok {
		logx.Info("msg_type is nil")
	}
	switch v1 {
	case "auth":

	case "chat_msg_list":
		data := msgMap["data"].(map[string]interface{})
		list := data["list"].([]interface{})
		msgIdList := make([]int64, 0)
		for k1 := range list {
			msgIdList = append(msgIdList, list[k1].(map[string]interface{})["id"].(int64))
		}
		err = conn.WriteJSON(map[string]interface{}{
			"msg_type": "ack_receive",
			"data": map[string]interface{}{
				"msg_id_list": msgIdList,
			},
		})
		if err != nil {
			logx.Info(err)
		}
	}

}
