package quic_protocol

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"sync"

	"github.com/lucas-clemente/quic-go"
	"github.com/weblazy/socket-cluster/protocol"
)

const HEAD_SIZE = 4
const HEADER = "BEGIN"

// 每个消息(包括头部)的最大长度， 这里最大可以设置4G
const MAX_LENGTH = 1024 * 70
const CHAN_MSG_COUNT = 2

type QuicProtocol struct {
	nodeHandler protocol.Node
}

func (this *QuicProtocol) Dial(addr string) (*QuicConnection, error) {
	tlsConf := &tls.Config{NextProtos: []string{"quic-echo-example"}, InsecureSkipVerify: true}
	session, err := quic.DialAddr(addr, tlsConf, nil)
	if err != nil {
		fmt.Println("err" + err.Error())
		return nil, err
	}
	stream, err := session.OpenStreamSync(context.Background())
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &QuicConnection{Stream: stream}, err
}

func (this *QuicProtocol) ListenAndServe(port int64) error {
	tlsConf := generateTLSConfig()
	listener, err := quic.ListenAddr(fmt.Sprintf(":%d", port), tlsConf, nil)
	if err != nil {
		fmt.Println(err)
	}
	for {
		sess, err := listener.Accept(context.Background())
		if err != nil {
			fmt.Println(err)
		} else {
			go this.handleClient(sess)
		}
	}
	return nil
}

func (this *QuicProtocol) handleClient(sess quic.Session) {
	msg := make(chan string, 10) // 这里设置消息channel可以容纳10个消息
	// 缓存区设置最大为4G字节， 如果单个消息大于这个值就不能接受了
	stream, err := sess.AcceptStream(context.Background())
	if err != nil {
		panic(err)
	}
	buffer1 := protocol.NewBuffer(HEADER, MAX_LENGTH)

	var wg sync.WaitGroup
	wg.Add(2) // 主的routine将等待两个routine(读消息, 打印消息)的完成
	go func() {
		doMsg(msg)
		defer wg.Add(-1)
	}()
	go func() {
		conn := &QuicConnection{Stream: stream}
		err := buffer1.Read(conn, this.nodeHandler.OnClientMsg)
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

// Setup a bare-bones TLS config for the server
func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)

	if err != nil {
		panic(err)
	}
	return &tls.Config{NextProtos: []string{"quic-echo-example"}, Certificates: []tls.Certificate{tlsCert}}
}

func doMsg(msg chan string) {
	count := 0
	for v := range msg {
		fmt.Println("第", count, "个消息体长:", len(v))
		count++
	}
}
