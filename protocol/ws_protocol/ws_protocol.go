package ws_protocol

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/weblazy/core/logx"
	"github.com/weblazy/socket-cluster/protocol"
)

var (
	authTime = 10 * time.Second
	upgrader = websocket.Upgrader{
		ReadBufferSize:    4096,
		WriteBufferSize:   4096,
		EnableCompression: true,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	defaultMasterPort int64 = 9527
)

type WsProtocol struct {
	nodeHandler protocol.Node
}

func (this *WsProtocol) Dial(addr string, protoFunc ...protocol.ProtoFunc) (protocol.Session, error) {
	return protocol.Session{}, nil
}

func (this *WsProtocol) ListenAndServe(port int64, nodeHandler protocol.Node, protoFunc ...protocol.ProtoFunc) error {
	this.nodeHandler = nodeHandler
	e := echo.New()
	e.GET("/trans", this.transHandler)
	e.GET("/client", this.clientHandler)
	go func() {
		err := e.Start(fmt.Sprintf(":%d", port))
		if err != nil {
			panic(err)
		}
	}()
	return nil
}

// transHandler deal node connection
func (this *WsProtocol) transHandler(c echo.Context) error {
	connect, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	defer func() {
		if connect != nil {
			connect.Close()
		}
	}()

	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			logx.Info(err)
		}
		return err
	}
	conn := &WsConnection{Conn: connect}
	// this.timer.SetTimer(connect.RemoteAddr().String(), conn, authTime)
	for {
		_, msg, err := connect.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logx.Info(err)
			}
			break
		}
		logx.Info(string(msg))
		this.nodeHandler.OnTransMsg(conn, msg)
	}
	return nil
}

// clientHandler deal client connection
func (this *WsProtocol) clientHandler(c echo.Context) error {
	connect, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			logx.Info(err)
		}
		return err
	}
	conn := &WsConnection{Conn: connect}
	defer func() {
		this.nodeHandler.OnClose(conn)
		if connect != nil {
			connect.Close()
		}
	}()

	this.nodeHandler.OnConnect(conn)
	for {
		_, msg, err := connect.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logx.Info(err)
			}
			break
		}
		logx.Info(string(msg))
		this.nodeHandler.OnClientMsg(conn, msg)
	}
	return nil
}
