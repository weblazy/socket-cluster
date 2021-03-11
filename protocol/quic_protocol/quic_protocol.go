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

	"github.com/lucas-clemente/quic-go"
	"github.com/weblazy/socket-cluster/protocol"
)

type QuicProtocol struct {
	nodeHandler protocol.Node
}

func (this *QuicProtocol) SetNodeHandler(nodeHandler protocol.Node) {
	this.nodeHandler = nodeHandler
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
}

func (this *QuicProtocol) handleClient(sess quic.Session) {
	// 缓存区设置最大为4G字节， 如果单个消息大于这个值就不能接受了
	stream, err := sess.AcceptStream(context.Background())
	if err != nil {
		panic(err)
	}
	conn := &QuicConnection{Stream: stream}
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

func (this *QuicProtocol) ServeConn(conn protocol.Connection, f func(conn protocol.Connection, p []byte)) error {
	// this.timer.SetTimer(connect.RemoteAddr().String(), conn, authTime)
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()
	return protocol.DefaultFlowProto.Read(conn.(*QuicConnection), this.nodeHandler.OnTransMsg)
}
