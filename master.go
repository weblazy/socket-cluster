package websocket_cluster

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"github.com/weblazy/core/logx"
	"github.com/weblazy/core/timingwheel"
	"github.com/weblazy/goutil"
)

type (
	MasterConf struct {
		SocketConf *SocketConfig //Socket config
		Password   string        //Password for auth when node connect on
	}
	MasterInfo struct {
		masterConf MasterConf
		nodeMap    goutil.Map // V is *nodeConn
		timer      *timingwheel.TimingWheel
		startTime  time.Time
	}
	nodeConn struct {
		conn    *websocket.Conn
		address string //Outside address
	}
)

var (
	masterInfo MasterInfo
)

// Start master node.
func StartMaster(cfg MasterConf) {
	timer, err := timingwheel.NewTimingWheel(time.Second, 300, func(k, v interface{}) {
		logx.Errorf("%s auth timeout", k)
		err := v.(*websocket.Conn).Close()
		if err != nil {
			logx.Error(err)
		}
	})
	defer timer.Stop()
	if err != nil {
		logx.Fatal(err)
	}
	masterInfo = MasterInfo{
		masterConf: cfg,
		nodeMap:    goutil.AtomicMap(),
		startTime:  time.Now(),
		timer:      timer,
	}
	e := echo.New()
	e.GET("/ws", masterHandler)
	addr := fmt.Sprintf("%s:%d", cfg.SocketConf.Ip, cfg.SocketConf.Port)
	e.Start(addr)
}

func masterHandler(c echo.Context) error {
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return err
	}
	masterInfo.timer.SetTimer(conn.RemoteAddr().String(), conn, authTime)
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			} else {
				log.Printf("socket close: %v", err)
			}
			break
		}
		masterInfo.OnMessage(conn, message)
		fmt.Printf("%#v\n", string(message))
	}
	return nil
}

func (masterInfo *MasterInfo) OnMessage(conn *websocket.Conn, message []byte) {
	data := make(map[string]interface{})
	err := json.Unmarshal(message, &data)
	if err != nil {
		log.Printf("error: %v", err)
	}
	v1, ok := data["type"]
	if !ok {
		log.Printf("type is nil")
	}
	switch v1 {
	case "auth":
		AuthConn(conn, data["data"].(map[string]interface{}))
	}
}

// Auth the node
func AuthConn(conn *websocket.Conn, args map[string]interface{}) error {
	sid := conn.RemoteAddr().String()
	masterInfo.timer.RemoveTimer(sid) //Cancel timeingwheel task
	if args["password"].(string) != masterInfo.masterConf.Password {
		logx.Errorf("Connect:%s,Wrong password:%s", sid, args["password"].(string))
		conn.Close()
		return fmt.Errorf("auth faild")
	}
	masterInfo.setConn(conn, args["trans_address"].(string))
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
		err := v.(*nodeConn).conn.WriteJSON(nodeList)
		if err != nil {
			fmt.Printf("%#v\n", err)
		}
		return true
	})
}

// set sets a *conn
func (mi *MasterInfo) setConn(conn *websocket.Conn, address string) {
	sid := conn.RemoteAddr().String()
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
		oldConn.Close()
	}
}
