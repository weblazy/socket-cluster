package websocket_cluster

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"github.com/weblazy/core/timingwheel"
	"github.com/weblazy/easy/utils/logx"
	"github.com/weblazy/goutil"
)

type (
	MasterInfo struct {
		masterConf *MasterConf
		nodeMap    goutil.Map // V is *Connection
		timer      *timingwheel.TimingWheel
		startTime  time.Time
	}
	nodeConn struct {
		conn    *Connection
		address string //Outside address
	}
)

var (
	masterInfo *MasterInfo
)

// Start master node.
func StartMaster(cfg *MasterConf) {
	timer, err := timingwheel.NewTimingWheel(time.Second, 300, func(k, v interface{}) {
		logx.Info(fmt.Sprintf("%s auth timeout", k))
		err := v.(*Connection).Conn.Close()
		if err != nil {
			logx.Info(err)
		}
	})
	defer timer.Stop()
	if err != nil {
		logx.Info(err)
	}
	masterInfo = &MasterInfo{
		masterConf: cfg,
		nodeMap:    goutil.AtomicMap(),
		startTime:  time.Now(),
		timer:      timer,
	}
	e := echo.New()
	e.GET("/master", masterHandler)
	err = e.Start(fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		logx.Info(err)
	}
}

func masterHandler(c echo.Context) error {
	connect, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	conn := &Connection{
		Conn: connect,
	}
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			logx.Info(err)
		}
		return err
	}
	masterInfo.timer.SetTimer(conn.Conn.RemoteAddr().String(), conn, authTime)
	for {
		_, message, err := conn.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logx.Info(err)
			} else {
				logx.Infof("socket close: %v", err)
			}
			break
		}
		masterInfo.OnMessage(conn, message)
	}
	masterInfo.nodeMap.Delete(connect.RemoteAddr().String())
	return nil
}

func (masterInfo *MasterInfo) OnMessage(conn *Connection, message []byte) {
	data := make(map[string]interface{})
	err := json.Unmarshal(message, &data)
	if err != nil {
		logx.Info(err)
	}
	v1, ok := data["type"]
	if !ok {
		logx.Info("type is nil")
	}
	switch v1 {
	case "auth":
		AuthConn(conn, data["data"].(map[string]interface{}))
	}
}

// Auth the node
func AuthConn(conn *Connection, params map[string]interface{}) error {
	sid := conn.Conn.RemoteAddr().String()
	masterInfo.timer.RemoveTimer(sid) //Cancel timeingwheel task
	if params["password"].(string) != masterInfo.masterConf.Password {
		logx.Infof("Connect:%s,Wrong password:%s", sid, params["password"].(string))
		conn.Conn.Close()
		return fmt.Errorf("auth faild")
	}
	masterInfo.setConn(conn, params["trans_address"].(string))
	masterInfo.broadcastAddresses() //Notify all node nodes that new nodes have joined
	return nil
}

//Notify all node nodes that new nodes have joined
func (mi *MasterInfo) broadcastAddresses() {
	nodeList := make([]string, 0)
	mi.nodeMap.Range(func(k interface{}, v interface{}) bool {
		nodeList = append(nodeList, v.(*nodeConn).address)
		return true
	})
	mi.nodeMap.Range(func(k interface{}, v interface{}) bool {
		data := Message{
			MessageType: "UpdateNodeList",
			Data:        nodeList,
		}

		err := v.(*nodeConn).conn.WriteJSON(data)
		if err != nil {
			logx.Info(err)
		}
		return true
	})
}

// set sets a *conn
func (mi *MasterInfo) setConn(conn *Connection, address string) {
	sid := conn.Conn.RemoteAddr().String()
	node := &nodeConn{
		address: address,
		conn:    conn,
	}
	_node, loaded := mi.nodeMap.LoadOrStore(sid, node)
	if !loaded {
		return
	}
	mi.nodeMap.Store(sid, node)
	if oldConn := _node.(*nodeConn).conn; conn != oldConn {
		oldConn.Conn.Close()
	}
}
