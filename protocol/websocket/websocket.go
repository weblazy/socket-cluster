package websocket

import (
	"fmt"
	"net/http"
	"sync"
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

type Connection struct {
	Conn        *websocket.Conn
	Mutex       sync.Mutex
	OnTransMsg  func(conn *protocol.Connection, msg []byte)
	OnClientMsg func(conn *protocol.Connection, msg []byte)
	OnConnect   func(conn *protocol.Connection)
	OnClose     func(conn *protocol.Connection)
}

func (this *Connection) ListenAndServe(port int64, OnTransMsg func(conn *protocol.Connection, msg []byte), clientHandler func(c echo.Context) error, protoFunc ...protocol.ProtoFunc) error {
	// this.transAddress = fmt.Sprintf("%s%s/trans", cfg.Host, cfg.TransPath)
	// this.clientAddress = fmt.Sprintf("%s%s/client", cfg.Host, cfg.ClientPath)
	this.OnTransMsg = OnTransMsg
	e := echo.New()
	e.GET("/trans", this.transHandler)
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

// transHandler deal node connection
func (this *Connection) transHandler(c echo.Context) error {
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
	conn := &protocol.Connection{Conn: connect}
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
		this.OnTransMsg(conn, msg)
	}
	return nil
}

// clientHandler deal client connection
func (this *Connection) clientHandler(c echo.Context) error {
	connect, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			logx.Info(err)
		}
		return err
	}
	conn := &protocol.Connection{Conn: connect}
	defer func() {
		this.OnClose(conn)
		if connect != nil {
			connect.Close()
		}
	}()

	this.OnConnect(conn)
	// log.Sugar.Info(connect.RemoteAddr().String(), connect.LocalAddr().String(), "start")
	for {
		_, msg, err := connect.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logx.Info(err)
			}
			break
		}
		logx.Info(string(msg))
		this.OnClientMsg(conn, msg)
	}

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
