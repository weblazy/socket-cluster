package websocket

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/weblazy/socket-cluster/protocol"
)

type Connection struct {
	Conn  *websocket.Conn
	Mutex sync.Mutex
}

func (this *Connection) ListenAndServe(port int64, transHandler, clientHandler func(c echo.Context) error, protoFunc ...protocol.ProtoFunc) error {
	// this.transAddress = fmt.Sprintf("%s%s/trans", cfg.Host, cfg.TransPath)
	// this.clientAddress = fmt.Sprintf("%s%s/client", cfg.Host, cfg.ClientPath)

	e := echo.New()
	e.GET("/trans", transHandler)
	e.GET("/client", clientHandler)
	// webGroup := e.Group(fmt.Sprintf("%s/web", cfg.ClientPath))
	// cfg.router(webGroup)
	go func() {
		err := e.Start(fmt.Sprintf(":%d", port))
		if err != nil {
			panic(err)
		}
	}()

	return nil
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
