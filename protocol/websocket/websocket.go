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

type WsProtocol struct {
	ConnectHandler protocol.Node
}

func (this *WsProtocol) ListenAndServe(port int64, connectHandler protocol.Node, protoFunc ...protocol.ProtoFunc) error {
	// this.transAddress = fmt.Sprintf("%s%s/trans", cfg.Host, cfg.TransPath)
	// this.clientAddress = fmt.Sprintf("%s%s/client", cfg.Host, cfg.ClientPath)
	this.ConnectHandler = connectHandler
	e := echo.New()
	e.GET("/trans", this.transHandler)
	e.GET("/client", this.clientHandler)
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
		this.ConnectHandler.OnTransMsg(conn, msg)
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
		this.ConnectHandler.OnClose(conn)
		if connect != nil {
			connect.Close()
		}
	}()

	this.ConnectHandler.OnConnect(conn)
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
		this.ConnectHandler.OnClientMsg(conn, msg)
	}

	return nil
}

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
func (conn *WsConnection) WriteMsg(msgType int, data []byte) error {
	conn.Mutex.Lock()
	defer conn.Mutex.Unlock()
	return conn.Conn.WriteMessage(msgType, data)
}

func (conn *WsConnection) Addr() string {
	return conn.Conn.RemoteAddr().String()
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
