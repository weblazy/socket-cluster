package tcp_protocol

import (
	"encoding/json"
	"net"
	"sync"

	"github.com/weblazy/socket-cluster/protocol"
)

type TcpConnection struct {
	Conn  net.Conn
	Mutex sync.Mutex
	protocol.Connection
}

// WriteJSON send json message
func (conn *TcpConnection) WriteJSON(data interface{}) error {
	conn.Mutex.Lock()
	defer conn.Mutex.Unlock()
	msg, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = conn.Conn.Write(msg)
	return err
}

// WriteMsg send byte array message
func (conn *TcpConnection) WriteMsg(msgType int, data []byte) error {
	conn.Mutex.Lock()
	defer conn.Mutex.Unlock()
	_, err := conn.Conn.Write(data)
	return err
}
