package quic_protocol

import (
	"sync"

	"github.com/lucas-clemente/quic-go"
	"github.com/weblazy/socket-cluster/protocol"
)

type QuicConnection struct {
	Stream quic.Stream
	Mutex  sync.Mutex
	protocol.FlowConnection
}

// WriteMsg send byte array message
func (conn *QuicConnection) WriteMsg(data []byte) error {
	conn.Mutex.Lock()
	defer conn.Mutex.Unlock()
	_, err := conn.Stream.Write(data)
	return err
}

func (conn *QuicConnection) ReadMsg(data []byte) (int, error) {
	return conn.Stream.Read(data)
}
