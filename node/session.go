package node

import (
	"sync/atomic"
	"unsafe"

	"github.com/weblazy/socket-cluster/protocol"
)

type Session struct {
	Conn     *protocol.Connection
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
