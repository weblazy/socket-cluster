package protocol

import (
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/gorilla/websocket"
)

type Connection struct {
	Conn        *websocket.Conn
	Mutex       sync.Mutex
	OnTransMsg  func(conn *Connection, msg []byte)
	OnClientMsg func(conn *Connection, msg []byte)
	OnConnect   func(conn *Connection)
	OnClose     func(conn *Connection)
}

type Connect interface {
	OnTransMsg(conn *Connection, msg []byte)
	OnClientMsg(conn *Connection, msg []byte)
	OnConnect(conn *Connection)
	OnClose(conn *Connection)
}

// WriteJSON send json message
func (conn *Connection) WriteJSON(data interface{}) error {
	conn.Mutex.Lock()
	defer conn.Mutex.Unlock()
	return conn.Conn.WriteJSON(data)
}

// WriteMsg send byte array message
func (conn *Connection) WriteMsg(msgType int, data []byte) error {
	conn.Mutex.Lock()
	defer conn.Mutex.Unlock()
	return conn.Conn.WriteMessage(msgType, data)
}

type Session struct {
	Conn     *Connection
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
