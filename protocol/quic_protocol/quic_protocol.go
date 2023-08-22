package quic_protocol

// import (
// 	"fmt"

// 	"github.com/quic-go/quic-go"
// 	"github.com/weblazy/socket-cluster/logx"
// 	"github.com/weblazy/socket-cluster/protocol"
// )

// type QuicProtocol struct {
// }

// func (this *QuicProtocol) ListenAndServe(port int64, onConnect func(conn protocol.Connection)) error {
// 	// Setup a bare-bones TLS config for the server
// 	tlsConf := protocol.GenerateTLSConfigForServer()
// 	listener, err := quic.ListenAddr(fmt.Sprintf(":%d", port), tlsConf, nil)
// 	if err != nil {
// 		logx.LogHandler.Error(err)
// 	}
// 	for {
// 		session, err := listener.Accept()
// 		if err != nil {
// 			logx.LogHandler.Error(err)
// 		} else {
// 			go func(session quic.Session) {
// 				// Use only the first stream
// 				stream, err := session.AcceptStream()
// 				if err != nil {
// 					panic(err)
// 				}
// 				conn := NewQuicConnection(stream, session)
// 				onConnect(conn)
// 			}(session)
// 		}
// 	}
// }

// func (this *QuicProtocol) Dial(addr string) (protocol.Connection, error) {
// 	// Setup a bare-bones TLS config for the client
// 	tlsConf := protocol.GenerateTLSConfigForClient()
// 	session, err := quic.DialAddr(addr, tlsConf, nil)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// Use only the first stream
// 	stream, err := session.OpenStreamSync()
// 	if err != nil {
// 		return nil, err
// 	}
// 	return NewQuicConnection(stream, session), err
// }
