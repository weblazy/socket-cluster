package quic_protocol

import (
	"sync"

	"github.com/lucas-clemente/quic-go"
	"github.com/weblazy/socket-cluster/protocol"
)

type QuicConnection struct {
	Stream              quic.Stream
	Mutex               sync.Mutex
	flowProtocolHandler protocol.Proto
}

func NewQuicConnection(stream quic.Stream) *QuicConnection {
	return &QuicConnection{
		Stream:              stream,
		flowProtocolHandler: protocol.NewFlowProtocol(protocol.HEADER, protocol.MAX_LENGTH),
	}
}

// WriteMsg sends byte array message
func (this *QuicConnection) WriteMsg(data []byte) error {
	data, err := this.flowProtocolHandler.Pack(data)
	if err != nil {
		return err
	}
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	_, err = this.Stream.Write(data)
	return err
}

// ReadMsg reads byte array message
func (this *QuicConnection) ReadMsg() ([]byte, error) {
	msg, err := this.flowProtocolHandler.ReadMsg(this.Stream)
	if err != nil {
		return nil, err
	}
	return msg, err
}

func (this *QuicConnection) Addr() string {
	return this.Stream.StreamID().InitiatedBy().String()
}

func (this *QuicConnection) Close() error {
	if this.Stream == nil {
		return nil
	}
	return this.Stream.Close()
}
