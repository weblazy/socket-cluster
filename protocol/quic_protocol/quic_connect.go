package quic_protocol

import (
	"sync"

	"github.com/lucas-clemente/quic-go"
	"github.com/weblazy/socket-cluster/protocol"
)

type QuicConnection struct {
	stream              quic.Stream
	session             quic.Session
	mutex               sync.Mutex
	flowProtocolHandler protocol.Proto
}

func NewQuicConnection(stream quic.Stream, session quic.Session) *QuicConnection {

	return &QuicConnection{
		session:             session,
		stream:              stream,
		flowProtocolHandler: protocol.NewFlowProtocol(protocol.HEADER, protocol.MAX_LENGTH, stream),
	}
}

// WriteMsg sends byte array message
func (this *QuicConnection) WriteMsg(data []byte) error {
	data, err := this.flowProtocolHandler.Pack(data)
	if err != nil {
		return err
	}
	this.mutex.Lock()
	defer this.mutex.Unlock()
	_, err = this.stream.Write(data)
	return err
}

// ReadMsg reads byte array message
func (this *QuicConnection) ReadMsg() ([]byte, error) {
	msg, err := this.flowProtocolHandler.ReadMsg()
	if err != nil {
		return nil, err
	}
	return msg, err
}

func (this *QuicConnection) Addr() string {
	return this.session.RemoteAddr().String()
}

func (this *QuicConnection) Close() error {
	if this.stream == nil {
		return nil
	}
	return this.stream.Close()
}
