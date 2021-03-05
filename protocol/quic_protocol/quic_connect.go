package quic_protocol

import (
	"encoding/json"
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
func (conn *QuicConnection) WriteMsg(data []byte) error {
	conn.Mutex.Lock()
	defer conn.Mutex.Unlock()
	_, err := conn.Stream.Write(data)
	return err
}
