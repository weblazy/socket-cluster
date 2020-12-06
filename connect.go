package websocket_cluster

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Connection struct {
	Conn  *websocket.Conn
	Mutex sync.Mutex
}

// @desc 发送消息
// @auth liuguoqiang 2020-09-07
// @param
// @return
func (conn *Connection) WriteJSON(data interface{}) error {
	conn.Mutex.Lock()
	defer conn.Mutex.Unlock()
	return conn.Conn.WriteJSON(data)
}

// @desc 发送消息
// @auth liuguoqiang 2020-09-07
// @param
// @return
func (conn *Connection) WriteMsg(msgType int, data []byte) error {
	conn.Mutex.Lock()
	defer conn.Mutex.Unlock()
	return conn.Conn.WriteMessage(msgType, data)
}
