package tcp_protocol

import (
	"encoding/json"
	"fmt"
	"log"
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

type TcpProtocol struct {
	ConnectHandler protocol.Node
}

func (this *TcpProtocol) Dail(addr string) (net.Conn, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", addr)
	if err != nil {
		log.Printf("Resolve tcp addr failed: %v\n", err)
		return nil, err
	}

	// 向服务器拨号
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Printf("Dial to server failed: %v\n", err)
		return nil, err
	}

	return conn, err
}

func (this *TcpProtocol) Listen(port string) {
	fmt.Println("tcp run on localhost:7123")
	listener, err := net.Listen("tcp", ":7123")
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println("break 1")

				fmt.Println(err.Error())
				break
			}
			go handleClient(conn)
		}
	}()
	select {}

}

func handleClient(conn net.Conn) {
}
