package websocket_cluster

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"github.com/weblazy/core/logx"
	"github.com/weblazy/core/timingwheel"
	"github.com/weblazy/goutil"
)

type (
	MasterConf struct {
		Addr     string //Socket config
		Password string //Password for auth when node connect on
	}
	MasterInfo struct {
		masterConf MasterConf
		nodeMap    goutil.Map // V is *websocket.Conn
		timer      *timingwheel.TimingWheel
		startTime  time.Time
	}
	nodeConn struct {
		conn    *websocket.Conn
		address string //Outside address
	}

	Auth struct {
		TransAddress string //Node address ip:port
		Password     string //Password for auth when node connect on
	}
)

var (
	masterInfo MasterInfo

	upgrader = websocket.Upgrader{
		ReadBufferSize:    4096,
		WriteBufferSize:   4096,
		EnableCompression: true,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
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
	e.GET("/ws", serveMaster)
	e.Start(cfg.Addr)
}

func serveMaster(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return err
	}
	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		fmt.Printf("%#v\n", string(message))
	}
	return nil
}

func serveTrans(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return err
	}
	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		fmt.Printf("%#v\n", string(message))
	}
	return nil
}

func serveClient(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return err
	}
	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		fmt.Printf("%#v\n", string(message))
	}
	return nil
}

// Auth the node
// func Auth(args *Auth) *tp.Status {
// 	session := m.Session()
// 	sid := conn.RemoteAddr().String()

// 	peer := m.Peer()
// 	psession, ok := peer.GetSession(sid)
// 	masterInfo.timer.RemoveTimer(sid) //Cancel timeingwheel task
// 	if args.Password != masterInfo.masterConf.Password && ok {
// 		logx.Errorf("Connect:%s,Wrong password:%s", sid, args.Password)
// 		psession.Close()
// 		return StatusUnauthorized
// 	}
// 	masterInfo.setSession(psession, args.TransAddress)
// 	masterInfo.broadcastAddresses() //Notify all node nodes that new nodes have joined
// 	return nil
// }

//Notify all node nodes that new nodes have joined
func (mi *MasterInfo) broadcastAddresses() {
	nodeList := make([]string, 0)
	mi.nodeMap.Range(func(k interface{}, v interface{}) bool {
		nodeList = append(nodeList, v.(nodeConn).address)
		return true
	})
	mi.nodeMap.Range(func(k interface{}, v interface{}) bool {
		err := v.(nodeConn).conn.WriteJSON(nodeList)
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
