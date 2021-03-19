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
	ProtoHandler protocol.Proto
}

// WriteMsg sends byte array message
func (conn *QuicConnection) WriteMsg(data []byte) error {
	data, err := conn.ProtoHandler.Pack(data)
	if err != nil {
		return err
	}
	conn.Mutex.Lock()
	defer conn.Mutex.Unlock()
	_, err = conn.Stream.Write(data)
	return err
}

// ReadMsg reads byte array message
func (conn *QuicConnection) ReadMsg(data []byte) (int, error) {
	return conn.Stream.Read(data)
}

func (conn *QuicConnection) Addr() string {
	return conn.Stream.StreamID().InitiatedBy().String()
}

func (conn *QuicConnection) Close() error {
	return conn.Stream.Close()
}
