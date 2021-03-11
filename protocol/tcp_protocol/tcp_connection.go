package tcp_protocol

import (
	"net"
	"sync"

	"github.com/weblazy/socket-cluster/protocol"
)

type TcpConnection struct {
	Conn  net.Conn
	Mutex sync.Mutex
	protocol.FlowConnection
	ProtoHandler protocol.Proto
}

func NewTcpConnection(conn net.Conn) *TcpConnection {
	return &TcpConnection{
		Conn:         conn,
		ProtoHandler: protocol.DefaultFlowProto,
	}
}

// WriteMsg send byte array message
func (conn *TcpConnection) WriteMsg(data []byte) error {
	data, err := conn.ProtoHandler.Pack(data)
	if err != nil {
		return err
	}
	conn.Mutex.Lock()
	defer conn.Mutex.Unlock()
	_, err = conn.Conn.Write(data)
	return err
}

func (conn *TcpConnection) ReadMsg(data []byte) (int, error) {
	return conn.Conn.Read(data)
}

func (conn *TcpConnection) Addr() string {
	return conn.Conn.RemoteAddr().String()
}

func (conn *TcpConnection) Close() error {
	return conn.Conn.Close()
}
