package protocol

import (
	"sync/atomic"
	"unsafe"
)

type Connection interface {
	WriteJSON(data interface{}) error
	WriteMsg(data []byte) error
	ReadMessage() (p []byte, err error)
	Close() error
	Addr() string
}

type Node interface {
	OnTransMsg(conn Connection, msg []byte)
	OnClientMsg(conn Connection, msg []byte)
	OnConnect(conn Connection)
	OnClose(conn Connection)
}

type Session struct {
	Conn     Connection
	ClientId string
}

// LoadClientId returns the session uid.
func (s *Session) LoadClientId() string {
	pointer := unsafe.Pointer(&s.ClientId)
	return *(*string)(atomic.LoadPointer(&pointer))
}

// StoreClientId sets the session uid.
func (s *Session) StoreClientId(newClientId string) {
	pointer := unsafe.Pointer(&s.ClientId)
	atomic.StorePointer(&pointer, unsafe.Pointer(&newClientId))
}

// CasClientId sets the session uid and return oldClientId
func (s *Session) CasClientId(newClientId string) string {
	newValue := unsafe.Pointer(&newClientId)
	pointer := unsafe.Pointer(&s.ClientId)
	for {
		oldValue := atomic.LoadPointer(&pointer)
		if atomic.CompareAndSwapPointer(&pointer, oldValue, newValue) {
			return *(*string)(oldValue)
		}
	}
}
