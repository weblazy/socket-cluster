package tcp_protocol

import (
	"fmt"
	"net"

	"github.com/weblazy/socket-cluster/logx"
	"github.com/weblazy/socket-cluster/protocol"
)

type TcpProtocol struct {
}

func (this *TcpProtocol) ListenAndServe(port int64, onConnect func(conn protocol.Connection)) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	go func() {
		for {
			connect, err := listener.Accept()
			if err != nil {
				logx.LogHandler.Error(err)
				break
			}
			go func(connect net.Conn) {
				conn := NewTcpConnection(connect)
				onConnect(conn)
			}(connect)
		}
	}()
	return nil
}

func (this *TcpProtocol) Dial(addr string) (protocol.Connection, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}
	return NewTcpConnection(conn), err
}
