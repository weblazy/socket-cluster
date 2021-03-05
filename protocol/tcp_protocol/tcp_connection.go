package tcp_protocol

import (
	"net"
	"sync"

	"github.com/weblazy/socket-cluster/protocol"
)

type TcpConnection struct {
	Conn  net.Conn
	Mutex sync.Mutex
	protocol.Connection
}

// WriteMsg send byte array message
func (conn *TcpConnection) WriteMsg(data []byte) error {
	conn.Mutex.Lock()
	defer conn.Mutex.Unlock()
	_, err := conn.Conn.Write(data)
	return err
}
