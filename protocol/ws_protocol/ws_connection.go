package ws_protocol

import (
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/weblazy/socket-cluster/protocol"
)

type WsConnection struct {
	Conn  *websocket.Conn
	Mutex sync.Mutex
	protocol.Connection
}

// WriteJSON send json message
func (conn *WsConnection) WriteJSON(data interface{}) error {
	conn.Mutex.Lock()
	defer conn.Mutex.Unlock()
	return conn.Conn.WriteJSON(data)
}

// WriteMsg send byte array message
func (conn *WsConnection) WriteMsg(data []byte) error {
	conn.Mutex.Lock()
	defer conn.Mutex.Unlock()
	return conn.Conn.WriteMessage(websocket.BinaryMessage, data)
}

func (conn *WsConnection) Addr() string {
	return conn.Conn.RemoteAddr().String()
}

func (conn *WsConnection) Close() error {
	return conn.Conn.Close()
}

func OptionHandler(c echo.Context) error {
	c.Response().Header().Set("Access-Control-Allow-Origin", "*")
	c.Response().Header().Set("Access-Control-Allow-Headers", "*")
	return c.String(200, "")
}
func WebHandler(c echo.Context) error {
	c.Response().Header().Set("Access-Control-Allow-Origin", "*")
	c.Response().Header().Set("Access-Control-Allow-Headers", "*")
	return c.JSON(200, "pong")
}

func originMiddlewareFunc(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Access-Control-Allow-Origin", "*")
		c.Response().Header().Set("Access-Control-Allow-Headers", "*")
		return next(c)
	}
}
