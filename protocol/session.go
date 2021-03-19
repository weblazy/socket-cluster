package protocol

import (
	"sync/atomic"
	"unsafe"
)

type Connection interface {
	ReadMsg() ([]byte, error)
	WriteMsg(data []byte) error
	Close() error
	Addr() string
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
