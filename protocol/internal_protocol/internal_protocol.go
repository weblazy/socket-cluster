package internal_protocol

import (
	"fmt"
	"log"
	"net"

	"github.com/weblazy/core/logx"
	"github.com/weblazy/socket-cluster/protocol"
	"github.com/weblazy/socket-cluster/protocol/tcp_protocol"
)

const HEAD_SIZE = 4
const HEADER = "BEGIN"

// 每个消息(包括头部)的最大长度， 这里最大可以设置4G
const MAX_LENGTH = 1024 * 70
const CHAN_MSG_COUNT = 2

type TcpProtocol struct {
	nodeHandler protocol.Node
}

func (this *TcpProtocol) Dail(addr string) (*tcp_protocol.TcpConnection, error) {
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
	return &tcp_protocol.TcpConnection{Conn: conn}, err
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

	// conn := &tcp_protocol.TcpConnection{Conn: connect}
	// this.timer.SetTimer(connect.RemoteAddr().String(), conn, authTime)
	// 缓存区设置最大为4G字节， 如果单个消息大于这个值就不能接受了
	buffer1 := protocol.NewBuffer(connect, HEADER, MAX_LENGTH)
	go func() {
		defer func() {
			if connect != nil {
				connect.Close()
			}
		}()
		err := buffer1.Read(this.doMsg)
		if err != nil {
			if err.Error() == "EOF" {
				// 对等方关闭了, 这里关闭chan, 通知接收消息的routine别等了，人家都关了
			} else {
				panic(err)
			}
		}
	}()
}

func (this *TcpProtocol) doMsg(conn net.Conn, msg []byte) {
	this.nodeHandler.OnTransMsg(&tcp_protocol.TcpConnection{Conn: conn}, msg)
	fmt.Println("个消息体长:", len(msg))
}
