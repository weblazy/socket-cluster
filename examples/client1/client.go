package main

import (
	"encoding/json"

	socket_cluster "socket-cluster"

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
		"msg_type": "login",
		"data": map[string]interface{}{
			"password": "password1",
			"username": "admin1",
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
	data := make(map[string]interface{})
	err := json.Unmarshal(msg, &data)
	if err != nil {
		logx.Info(err)
	}
	v1, ok := data["msg_type"]
	if !ok {
		logx.Info("msg_type is nil")
	}
	switch v1 {
	case "auth":

	case "chat_msg_list":
		data := data["data"].(map[string]interface{})
		list := data["list"].([]interface{})
		msgIdList := make([]int64, 0)
		for k1 := range list {
			msgIdList = append(msgIdList, int64(list[k1].(map[string]interface{})["id"].(float64)))
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
