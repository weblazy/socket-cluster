package tcp_protocol

import (
	"fmt"
	"log"
	"net"

	"github.com/weblazy/core/logx"
	"github.com/weblazy/socket-cluster/protocol"
)

type TcpProtocol struct {
	nodeHandler protocol.Node
}

func (this *TcpProtocol) SetNodeHandler(nodeHandler protocol.Node) {
	this.nodeHandler = nodeHandler
}

func (this *TcpProtocol) Dial(addr string) (protocol.Connection, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", addr)
	if err != nil {
		log.Printf("Resolve tcp addr failed: %v\n", err)
		return nil, err
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Printf("Dial to server failed: %v\n", err)
		return nil, err
	}
	return NewTcpConnection(conn), err
}

func (this *TcpProtocol) ListenAndServe(port int64) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				logx.Info(err)
				break
			}
			go this.handleClient(conn)
		}
	}()
	return nil
}

func (this *TcpProtocol) handleClient(connect net.Conn) {
	// 缓存区设置最大为4G字节， 如果单个消息大于这个值就不能接受了
	conn := NewTcpConnection(connect)
	this.nodeHandler.OnConnect(conn)
	go func() {
		defer func() {
			this.nodeHandler.OnClose(conn)
			if conn != nil {
				conn.Close()
			}
		}()
		err := protocol.DefaultFlowProto.Read(conn, this.nodeHandler.OnClientMsg)
		if err != nil {
			if err.Error() == "EOF" {
				// 对等方关闭了, 这里关闭chan, 通知接收消息的routine别等了，人家都关了
			} else {
				panic(err)
			}
		}
	}()
}

func (this *TcpProtocol) ServeConn(conn protocol.Connection, f func(conn protocol.Connection, p []byte)) error {
	// this.timer.SetTimer(connect.RemoteAddr().String(), conn, authTime)
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()
	return protocol.DefaultFlowProto.Read(conn.(*TcpConnection), this.nodeHandler.OnTransMsg)
}
