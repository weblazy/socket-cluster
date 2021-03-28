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
}

func (this *QuicProtocol) ListenAndServe(port int64, onConnect func(conn protocol.Connection)) error {
	tlsConf := generateTLSConfig()
	listener, err := quic.ListenAddr(fmt.Sprintf(":%d", port), tlsConf, nil)
	if err != nil {
		fmt.Println(err)
	}
	for {
		session, err := listener.Accept(context.Background())
		if err != nil {
			fmt.Println(err)
		} else {
			go func(session quic.Session) {
				// Use only the first stream
				stream, err := session.AcceptStream(context.Background())
				if err != nil {
					panic(err)
				}
				conn := NewQuicConnection(stream, session)
				onConnect(conn)
			}(session)
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

func (this *QuicProtocol) Dial(addr string) (*QuicConnection, error) {
	tlsConf := &tls.Config{NextProtos: []string{"quic-echo-example"}, InsecureSkipVerify: true}
	session, err := quic.DialAddr(addr, tlsConf, nil)
	if err != nil {
		return nil, err
	}
	// Use only the first stream
	stream, err := session.OpenStreamSync(context.Background())
	if err != nil {
		return nil, err
	}
	return NewQuicConnection(stream, session), err
}
