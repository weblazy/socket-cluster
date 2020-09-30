package websocket_cluster

import (
	"sync/atomic"
	"unsafe"

	"github.com/gorilla/websocket"
)

type Session struct {
	Conn *websocket.Conn
	Uid  string
}

// LoadUid returns the session uid.
func (s *Session) LoadUid() string {
	pointer := unsafe.Pointer(&s.Uid)
	return *(*string)(atomic.LoadPointer(&pointer))
}

// StoreUid sets the session uid.
func (s *Session) StoreUid(newUid string) {
	pointer := unsafe.Pointer(&s.Uid)
	atomic.StorePointer(&pointer, unsafe.Pointer(&newUid))
}

// CasUid sets the session uid and return oldUid
func (s *Session) CasUid(newUid string) string {
	newValue := unsafe.Pointer(&newUid)
	pointer := unsafe.Pointer(&s.Uid)
	for {
		oldValue := atomic.LoadPointer(&pointer)
		if atomic.CompareAndSwapPointer(&pointer, oldValue, newValue) {
			return *(*string)(oldValue)
		}
	}
}
