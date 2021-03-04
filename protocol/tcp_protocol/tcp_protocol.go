package tcp_protocol

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/weblazy/core/logx"
	"github.com/weblazy/socket-cluster/protocol"
)

const HEAD_SIZE = 4
const HEADER = "BEGIN"

// 每个消息(包括头部)的最大长度， 这里最大可以设置4G
const BUFFER_LENGTH = 1024 * 70
const CHAN_MSG_COUNT = 2

type TcpProtocol struct {
	nodeHandler protocol.Node
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
			go handleClient(conn)
		}
	}()
	return nil
}

func handleClient(conn net.Conn) {
	msg := make(chan string, 10) // 这里设置消息channel可以容纳10个消息
	// 缓存区设置最大为4G字节， 如果单个消息大于这个值就不能接受了
	buffer1 := protocol.NewBuffer(conn, HEADER, BUFFER_LENGTH)

	var wg sync.WaitGroup
	wg.Add(2) // 主的routine将等待两个routine(读消息, 打印消息)的完成
	go func() {
		doMsg(msg)
		defer wg.Add(-1)
	}()
	go func() {
		err := buffer1.Read(msg)
		if err != nil {
			if err.Error() == "EOF" {
				close(msg) // 对等方关闭了, 这里关闭chan, 通知接收消息的routine别等了，人家都关了
			} else {
				panic(err)
			}
		}
		defer wg.Add(-1)
	}()
	wg.Wait()
	fmt.Println("一个客户端处理的消息处理完毕")
}

func doMsg(msg chan string) {
	count := 0
	for v := range msg {
		fmt.Println("第", count, "个消息体长:", len(v))
		count++
	}
}
