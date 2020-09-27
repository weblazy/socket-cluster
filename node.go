package websocket_cluster

import (
	"fmt"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"github.com/weblazy/core/consistenthash/unsafehash"
	"github.com/weblazy/core/database/redis"
	"github.com/weblazy/core/logx"
	"github.com/weblazy/core/syncx"
	"github.com/weblazy/core/timingwheel"
	tp "github.com/weblazy/teleport"

	"strconv"
	"time"
)

type (
	NodeConf struct {
		RedisConf     redis.RedisConf
		RedisMaxCount uint32
		ClientConf    SocketConfig
		TransConf     SocketConfig
		TransPort     int64  //Internal communication port
		MasterAddress string //Master address
		Password      string //Password for auth when connect to master
		PingInterval  int    //Heartbeat interval
	}

	SocketConfig struct {
		Addr string
		Port int64
	}
	Test struct {
		Reloadable             bool
		PingInterval           int64
		PingNotResPonseLimit   int64
		PingData               string
		CecreteKey             string
		Router                 func()
		SendToWorkerBufferSize int64
		SendToClientBufferSize int64
		nodeSessions           map[string]tp.Session
		startTime              time.Time
		// gatewayConnections map[string]string
		// businessConnections map[string]tp.Session
	}
	NodeInfo struct {
		bizRedis       *redis.Redis
		nodeConf       NodeConf
		masterConn     *websocket.Conn
		nodeConns      map[string]*websocket.Conn
		clientMapConns map[string]*websocket.Conn
		uidSessions    *syncx.ConcurrentDoubleMap
		groupSessions  *syncx.ConcurrentDoubleMap
		clientConns    []*websocket.Conn        //External communication peer
		clientAddress  string                   //External communication address
		transConns     []*websocket.Conn        //Internal communication peer
		transAddress   string                   //Internal communication address
		timer          *timingwheel.TimingWheel //Timingwheel
		startTime      time.Time
		userHashRing   *unsafehash.Consistent //UsHash ring storage userId
		groupHashRing  *unsafehash.Consistent //UsHash ring storage groupId
	}

	Message struct {
		uid  string
		path string
		data interface{}
	}
)

var (
	nodeInfo NodeInfo
)

const (
	PERSISTENCE_CONNECTION_PING_INTERVAL = 25
	redisInterval                        = 10
	redisZsortKey                        = "tpcluster_node"
)

// NewPeer creates a new peer.
func StartNode(cfg NodeConf) {
	redis := redis.NewRedis(cfg.RedisConf.Host, cfg.RedisConf.Type, cfg.RedisConf.Pass)
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
	if cfg.PingInterval == 0 {
		cfg.PingInterval = 1
	}
	nodeInfo = NodeInfo{
		nodeConf:       cfg,
		nodeConns:      make(map[string]*websocket.Conn),
		clientMapConns: make(map[string]*websocket.Conn),
		bizRedis:       redis,
		uidSessions:    syncx.NewConcurrentDoubleMap(32),
		groupSessions:  syncx.NewConcurrentDoubleMap(32),
		startTime:      time.Now(),
		timer:          timer,
		userHashRing:   unsafehash.NewConsistent(cfg.RedisMaxCount),
		groupHashRing:  unsafehash.NewConsistent(cfg.RedisMaxCount),
	}
	port := strconv.FormatInt(int64(cfg.ClientConf.Port), 10)
	nodeInfo.clientAddress = cfg.ClientConf.Addr + ":" + port

	e := echo.New()
	e.GET("/ws", serveTrans)
	go e.Start(nodeInfo.transAddress)

	SendPing()
	UpdateRedis()
	go ConnectToMaster(cfg)
	e1 := echo.New()
	e1.GET("/ws", serveClient)
	log.Fatal(e1.Start(nodeInfo.clientAddress))

}

//Heartbeat
func SendPing() {
	go func() {
		for {
			time.Sleep(time.Duration(nodeInfo.nodeConf.PingInterval) * time.Second)
			// nodeInfo.masterConn.WriteMessage()
			nodeInfo.masterConn.WriteMessage(websocket.PingMessage, []byte{})
			for k1 := range nodeInfo.transConns {
				nodeInfo.transConns[k1].WriteMessage(websocket.PingMessage, []byte{})
			}
		}
	}()
}

//Update clients num
func UpdateRedis() {
	go func() {
		for {
			time.Sleep(redisInterval * time.Second)
			nodeInfo.bizRedis.Zadd(redisZsortKey, int64(len(nodeInfo.clientConns)), nodeInfo.clientAddress)
		}
	}()
}

//Connect to master
func ConnectToMaster(cfg NodeConf) {
	u := url.URL{Scheme: "ws", Host: cfg.MasterAddress, Path: "/ws"}
	log.Printf("connecting to %s", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	// defer conn.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	nodeInfo.masterConn = conn
	port := strconv.FormatInt(int64(cfg.TransConf.Port), 10)
	nodeInfo.transAddress = cfg.TransConf.Addr + ":" + port
	auth := &Auth{
		Password:     nodeInfo.nodeConf.Password,
		TransAddress: nodeInfo.transAddress,
	}
	err = conn.WriteJSON(auth)
	if err != nil {
		fmt.Printf("%#v\n", err)
	}
}
