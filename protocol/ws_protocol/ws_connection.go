package ws_protocol

import (
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/weblazy/easy/utils/logx"
	"github.com/weblazy/socket-cluster/protocol"
)

type WsConnection struct {
	Conn  *websocket.Conn
	Mutex sync.Mutex
	protocol.Connection
}

func NewWsConnection(conn *websocket.Conn) *WsConnection {
	return &WsConnection{
		Conn: conn,
	}
}

// WriteMsg send byte array message
func (this *WsConnection) WriteMsg(data []byte) error {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	return this.Conn.WriteMessage(websocket.TextMessage, data)
}

func (this *WsConnection) ReadMsg() ([]byte, error) {
	_, msg, err := this.Conn.ReadMessage()
	if err != nil {
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			logx.Info(err)
		}
		return nil, err
	}
	return msg, err
}

func (this *WsConnection) Addr() string {
	return this.Conn.RemoteAddr().String()
}

func (this *WsConnection) Close() error {
	if this.Conn == nil {
		return nil
	}
	return this.Conn.Close()
}

func OptionHandler(c echo.Context) error {
	c.Response().Header().Set("Access-Control-Allow-Origin", "*")
	c.Response().Header().Set("Access-Control-Allow-Headers", "*")
	return c.String(200, "")
}

func OriginMiddlewareFunc(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Access-Control-Allow-Origin", "*")
		c.Response().Header().Set("Access-Control-Allow-Headers", "*")
		return next(c)
	}
}
