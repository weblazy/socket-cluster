package websocket_cluster

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Connection struct {
	Conn  *websocket.Conn
	Mutex sync.Mutex
}

// WriteJSON send json message
func (conn *Connection) WriteJSON(data interface{}) error {
	conn.Mutex.Lock()
	defer conn.Mutex.Unlock()
	return conn.Conn.WriteJSON(data)
}

// WriteMsg send byte array message
func (conn *Connection) WriteMsg(msgType int, data []byte) error {
	conn.Mutex.Lock()
	defer conn.Mutex.Unlock()
	return conn.Conn.WriteMessage(msgType, data)
}
