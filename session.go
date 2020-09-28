package websocket_cluster

import (
	"sync/atomic"
	"unsafe"

	"github.com/gorilla/websocket"
)

type session struct {
	conn *websocket.Conn
	uid  string
}

// LoadUid returns the session uid.
func (s *session) LoadUid() string {
	pointer := unsafe.Pointer(&s.uid)
	return *(*string)(atomic.LoadPointer(&pointer))
}

// StoreUid sets the session uid.
func (s *session) StoreUid(newUid string) {
	pointer := unsafe.Pointer(&s.uid)
	atomic.StorePointer(&pointer, unsafe.Pointer(&newUid))
}

// CasUid sets the session uid and return oldUid
func (s *session) CasUid(newUid string) string {
	newValue := unsafe.Pointer(&newUid)
	pointer := unsafe.Pointer(&s.uid)
	for {
		oldValue := atomic.LoadPointer(&pointer)
		if atomic.CompareAndSwapPointer(&pointer, oldValue, newValue) {
			return *(*string)(oldValue)
		}
	}
}
