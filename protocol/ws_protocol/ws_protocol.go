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

func (this *WsProtocol) SetNodeHandler(nodeHandler protocol.Node) {
	this.nodeHandler = nodeHandler
}

func (this *WsProtocol) Dial(addr string) (protocol.Connection, error) {
	connect, _, err := websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		return nil, err
	}
	return &WsConnection{Conn: connect}, nil
}

func (this *WsProtocol) ListenAndServe(port int64) error {
	e := echo.New()
	e.GET("/client", this.clientHandler)
	go func() {
		err := e.Start(fmt.Sprintf(":%d", port))
		if err != nil {
			panic(err)
		}
	}()
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

func (this *WsProtocol) ServeConn(conn protocol.Connection, OnTransMsg func(conn protocol.Connection, msg []byte)) error {
	defer func() {
		this.nodeHandler.OnClose(conn)
		if conn != nil {
			conn.Close()
		}
	}()

	this.nodeHandler.OnConnect(conn)
	for {
		_, msg, err := conn.(*WsConnection).Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logx.Info(err)
			}
			return err
		}
		logx.Info(string(msg))
		this.nodeHandler.OnClientMsg(conn, msg)
	}
}
