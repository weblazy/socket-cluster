package quic_protocol

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"sync"

	"github.com/lucas-clemente/quic-go"
	"github.com/weblazy/socket-cluster/protocol"
)

type QuicConnection struct {
	Stream quic.Stream
	Mutex  sync.Mutex
	protocol.Connection
}

// WriteJSON send json message
func (conn *QuicConnection) WriteJSON(data interface{}) error {
	conn.Mutex.Lock()
	defer conn.Mutex.Unlock()
	msg, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = conn.Stream.Write(msg)
	return err
}

// WriteMsg send byte array message
func (conn *QuicConnection) WriteMsg(msgType int, data []byte) error {
	conn.Mutex.Lock()
	defer conn.Mutex.Unlock()
	_, err := conn.Stream.Write(data)
	return err
}

type QuicProtocol struct {
	ConnectHandler protocol.Node
}

func (this *QuicProtocol) Dail(addr string) (quic.Stream, error) {
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
	return stream, err
}

func (this *QuicProtocol) Listen(addr string) {
	tlsConf := generateTLSConfig()
	listener, err := quic.ListenAddr(addr, tlsConf, nil)
	if err != nil {
		fmt.Println(err)
	}
	for {
		sess, err := listener.Accept(context.Background())
		if err != nil {
			fmt.Println(err)
		} else {
			go handleClient(sess)
		}
	}
}

func handleClient(sess quic.Session) {
	stream, err := sess.AcceptStream(context.Background())
	if err != nil {
		panic(err)
	} else {
		for {
			buf := make([]byte, 1024)
			_, err = io.ReadFull(stream, buf)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Printf("Client: Got '%s'\n", buf)
		}
	}
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
