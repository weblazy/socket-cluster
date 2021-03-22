package tcp_protocol

import (
	"net"
	"sync"

	"github.com/weblazy/socket-cluster/protocol"
)

type TcpConnection struct {
	Conn                net.Conn
	Mutex               sync.Mutex
	flowProtocolHandler protocol.Proto
}

func NewTcpConnection(conn net.Conn) *TcpConnection {
	return &TcpConnection{
		Conn:                conn,
		flowProtocolHandler: protocol.NewFlowProtocol(protocol.HEADER, protocol.MAX_LENGTH, conn),
	}
}

// WriteMsg send byte array message
func (this *TcpConnection) WriteMsg(data []byte) error {
	data, err := this.flowProtocolHandler.Pack(data)
	if err != nil {
		return err
	}
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	_, err = this.Conn.Write(data)
	return err
}

func (this *TcpConnection) ReadMsg() ([]byte, error) {
	msg, err := this.flowProtocolHandler.ReadMsg()
	if err != nil {
		return nil, err
	}
	return msg, err
}

func (this *TcpConnection) Addr() string {
	return this.Conn.RemoteAddr().String()
}

func (this *TcpConnection) Close() error {
	return this.Conn.Close()
}
