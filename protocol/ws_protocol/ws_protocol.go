package ws_protocol

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/weblazy/socket-cluster/protocol"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:    4096,
		WriteBufferSize:   4096,
		EnableCompression: true,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

type WsProtocol struct {
	ClientPath  string
	HandlerFunc func(conn protocol.Connection) echo.HandlerFunc
}

func NewWsProtocol(clientPath string, h func(conn protocol.Connection) echo.HandlerFunc) *WsProtocol {
	return &WsProtocol{
		ClientPath:  clientPath,
		HandlerFunc: h,
	}
}

func (this *WsProtocol) ListenAndServe(port int64, onConnect func(conn protocol.Connection)) error {
	e := echo.New()
	e.GET(this.ClientPath, func(c echo.Context) error {
		connect, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			if _, ok := err.(websocket.HandshakeError); !ok {

			}
			return err
		}
		conn := NewWsConnection(connect)
		handler := this.HandlerFunc(conn)
		err = handler(c)
		if err != nil {
			return err
		}
		onConnect(conn)
		return nil
	})
	go func() {
		err := e.Start(fmt.Sprintf(":%d", port))
		if err != nil {
			panic(err)
		}
	}()
	return nil
}

func (this *WsProtocol) Dial(addr string) (protocol.Connection, error) {
	connect, _, err := websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		return nil, err
	}
	return NewWsConnection(connect), nil
}
