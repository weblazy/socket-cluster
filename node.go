package websocket_cluster

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"github.com/weblazy/core/consistenthash/unsafehash"
	"github.com/weblazy/core/database/redis"
	"github.com/weblazy/core/mapreduce"
	"github.com/weblazy/core/syncx"
	"github.com/weblazy/core/timingwheel"
	"github.com/weblazy/easy/utils/logx"
	"github.com/weblazy/goutil"

	"strconv"
	"time"
)

type (
	NodeConf struct {
		RedisConf     redis.RedisConf
		RedisMaxCount uint32
		ClientConf    SocketConfig
		TransConf     SocketConfig
		MasterAddress string //Master address
		Password      string //Password for auth when connect to master
		PingInterval  int64  //Heartbeat interval
	}

	NodeInfo struct {
		bizRedis      *redis.Redis
		nodeConf      *NodeConf
		masterConn    *websocket.Conn
		uidSessions   *syncx.ConcurrentDoubleMap
		clientConns   goutil.Map               //External communication value is *session
		clientAddress string                   //External communication address
		transConns    goutil.Map               //Internal communication value is *websocket.Conn
		transAddress  string                   //Internal communication address
		timer         *timingwheel.TimingWheel //Timingwheel
		startTime     time.Time
		userHashRing  *unsafehash.Consistent //UsHash ring storage userId
		groupHashRing *unsafehash.Consistent //UsHash ring storage groupId
	}

	Message struct {
		uid         string      `json:"uid"`
		MessageType string      `json:"message_type"`
		data        interface{} `json:"data"`
	}

	NodeOption func(opt *NodeConf)
)

var (
	nodeInfo *NodeInfo
)

const (
	PERSISTENCE_CONNECTION_PING_INTERVAL = 25
	redisInterval                        = 10
	redisZsortKey                        = "tpcluster_node"
)

// NewPeer creates a new peer.
func NewNodeConf() *NodeConf {
	return &NodeConf{
		TransConf: SocketConfig{
			Ip:   "127.0.0.1",
			Port: 8080,
		},
		ClientConf: SocketConfig{
			Ip:   "127.0.0.1",
			Port: 8080,
		},
		RedisConf: redis.RedisConf{
			Host: "127.0.0.1:6379",
			Type: "node",
		},
		MasterAddress: defaultMasterAddress,
		Password:      defaultPassword,
		PingInterval:  defaultPingInterval,
	}
}

func (conf *NodeConf) WithMasterAddress(masterAddress string) *NodeConf {
	conf.MasterAddress = masterAddress
	return conf
}

func (conf *NodeConf) WithPassword(password string) *NodeConf {
	conf.Password = password
	return conf
}

func (conf *NodeConf) WithTransConf(transConf SocketConfig) *NodeConf {
	conf.TransConf = transConf
	return conf
}

func (conf *NodeConf) WithClientConf(clientConf SocketConfig) *NodeConf {
	conf.ClientConf = clientConf
	return conf
}

func (conf *NodeConf) WithRedisConf(redisConf redis.RedisConf) *NodeConf {
	conf.RedisConf = redisConf
	return conf
}

func (conf *NodeConf) WithPing(pingInterval int64) *NodeConf {
	conf.PingInterval = pingInterval
	return conf
}

// NewPeer creates a new peer.
func StartNode(cfg *NodeConf) {
	redis := redis.NewRedis(cfg.RedisConf.Host, cfg.RedisConf.Type, cfg.RedisConf.Pass)
	timer, err := timingwheel.NewTimingWheel(time.Second, 300, func(k, v interface{}) {
		logx.Infof("%s auth timeout", k)
		err := v.(*websocket.Conn).Close()
		if err != nil {
			logx.Info(err)
		}
	})
	defer timer.Stop()
	if err != nil {
		logx.Info(err)
	}
	nodeInfo = &NodeInfo{
		nodeConf:      cfg,
		bizRedis:      redis,
		uidSessions:   syncx.NewConcurrentDoubleMap(32),
		startTime:     time.Now(),
		timer:         timer,
		userHashRing:  unsafehash.NewConsistent(cfg.RedisMaxCount),
		groupHashRing: unsafehash.NewConsistent(cfg.RedisMaxCount),
		transConns:    goutil.AtomicMap(),
		clientConns:   goutil.AtomicMap(),
	}
	nodeInfo.clientAddress = fmt.Sprintf("%s:%d", cfg.ClientConf.Ip, cfg.ClientConf.Port)
	nodeInfo.transAddress = fmt.Sprintf("%s:%d", cfg.TransConf.Ip, cfg.TransConf.Port)
	e := echo.New()
	e.GET("/ws", transHandler)
	go func() {
		err := e.Start(nodeInfo.transAddress)
		if err != nil {
			logx.Info(err)
		}
	}()

	SendPing()
	UpdateRedis()
	go ConnectToMaster(cfg)
	e1 := echo.New()
	e1.GET("/ws", clientHandler)
	log.Fatal(e1.Start(nodeInfo.clientAddress))

}

func transHandler(c echo.Context) error {
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			logx.Info(err)
		}
		return err
	}
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logx.Info(err)
			}
			break
		}
		logx.Info(string(message))
	}
	return nil
}

func clientHandler(c echo.Context) error {
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			logx.Info(err)
		}
		return err
	}
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logx.Info(err)
			}
			break
		}
		logx.Info(string(message))
	}
	return nil
}

//Heartbeat
func SendPing() {
	go func() {
		for {
			time.Sleep(time.Duration(nodeInfo.nodeConf.PingInterval) * time.Second)
			// nodeInfo.masterConn.WriteMessage()
			nodeInfo.masterConn.WriteMessage(websocket.PingMessage, []byte{})
			nodeInfo.transConns.Range(func(k, v interface{}) bool {
				v.(*websocket.Conn).WriteMessage(websocket.PingMessage, []byte{})
				return true
			})
		}
	}()
}

//Update clients num
func UpdateRedis() {
	go func() {
		for {
			time.Sleep(redisInterval * time.Second)
			nodeInfo.bizRedis.Zadd(redisZsortKey, int64(nodeInfo.clientConns.Len()), nodeInfo.clientAddress)
		}
	}()
}

//Connect to master
func ConnectToMaster(cfg *NodeConf) {
	u := url.URL{Scheme: "ws", Host: cfg.MasterAddress, Path: "/ws"}
	logx.Infof("connecting to %s", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		logx.Info("dial:", err)
	}
	// defer conn.Close()

	go func() {
		defer conn.Close()
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			logx.Infof("recv: %s", message)
		}
	}()

	nodeInfo.masterConn = conn
	auth := &Auth{
		Password:     nodeInfo.nodeConf.Password,
		TransAddress: nodeInfo.transAddress,
	}
	data := map[string]interface{}{
		"type": "auth",
		"data": auth,
	}
	err = conn.WriteJSON(data)
	if err != nil {
		logx.Info(err)
	}
}

//Determine if a uid is online
func IsOnline(uid string) bool {
	now := time.Now().Unix()
	node := nodeInfo.userHashRing.Get(uid)
	addrMap, err := node.Extra.(*redis.Redis).Hgetall(userPrefix + uid)
	if err == nil {
		return false
	}
	for _, value := range addrMap {
		old, _ := strconv.ParseInt(value, 10, 64)
		if now < old {
			return true
		}
	}
	return false
}

func (nodeInfo *NodeInfo) OnTransMessage(conn *websocket.Conn, message []byte) {
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
		AuthTrans(conn, data["data"].(map[string]interface{}))
	}
}

func (nodeInfo *NodeInfo) OnMasterMessage(message []byte) {
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
	case "UpdateNodeList":
		nodeInfo.UpdateNodeList(data["data"].([]string))
	}
}

// Auth the node
func AuthTrans(conn *websocket.Conn, args map[string]interface{}) error {
	sid := conn.RemoteAddr().String()
	nodeInfo.timer.RemoveTimer(sid) //Cancel timeingwheel task
	if args["password"].(string) != nodeInfo.nodeConf.Password {
		logx.Infof("Connect:%s,Wrong password:%s", sid, args["password"].(string))
		conn.Close()
		return fmt.Errorf("auth faild")
	}
	nodeInfo.transConns.Store(sid, conn)
	return nil
}

// Auth the node
func AuthClient(conn *websocket.Conn, args map[string]interface{}) error {
	sid := conn.RemoteAddr().String()
	nodeInfo.timer.RemoveTimer(sid) //Cancel timeingwheel task
	if args["password"].(string) != nodeInfo.nodeConf.Password {
		logx.Infof("Connect:%s,Wrong password:%s", sid, args["password"].(string))
		conn.Close()
		return fmt.Errorf("auth faild")
	}
	nodeInfo.clientConns.Store(sid, conn)
	return nil
}

// Add handles addition request
func (nodeInfo *NodeInfo) UpdateNodeList(nodeList []string) error {
	for _, value := range nodeList {
		u := url.URL{Scheme: "ws", Host: value, Path: "/ws"}
		logx.Infof("connecting to %s", u.String())

		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			logx.Info("dial:", err)
		}

		auth := &Auth{
			Password:     nodeInfo.nodeConf.Password,
			TransAddress: nodeInfo.transAddress,
		}
		err = conn.WriteJSON(auth)
		if err != nil {
			logx.Info(err)
		}
	}

	return nil
}

func JoinGroup(gid, uid string) error {
	now := time.Now().Unix()
	node := nodeInfo.groupHashRing.Get(gid)
	err := node.Extra.(*redis.Redis).Hset(groupPrefix+gid, uid, strconv.FormatInt(now, 10))
	if err != nil {
		logx.Info(err)
	}
	return nil
}

func LeaveGroup(gid, uid string) error {
	node := nodeInfo.groupHashRing.Get(gid)
	_, err := node.Extra.(*redis.Redis).Hdel(groupPrefix+gid, uid)
	if err != nil {
		logx.Info(err)
	}
	return nil
}

func SendToGroup(gid string, path string, req interface{}) error {
	uids := GroupOnline(gid)
	mapreduce.MapVoid(func(source chan<- interface{}) {
		for k1, _ := range uids {
			nodeInfo.uidSessions.RangeShard(uids[k1], func(key2 string, value interface{}) bool {
				source <- value
				return true
			})
		}
	}, func(item interface{}) {
		se := item.(*Session)
		err := se.Conn.WriteJSON(req)
		if err != nil {
			logx.Info(err)
		}
	})
	return nil
}

//Get online users in the group
func GroupOnline(gid string) []string {
	now := time.Now().Unix()
	node := nodeInfo.groupHashRing.Get(gid)
	uids := make([]string, 0)
	addrMap, err := node.Extra.(*redis.Redis).Hgetall(groupPrefix + gid)
	if err == nil {
		return uids
	}
	for key, value := range addrMap {
		old, _ := strconv.ParseInt(value, 10, 64)
		if now < old {
			uids = append(uids, key)
		}
	}
	return uids
}

//Get online users in the group
func GetSessionsByUids(uids []string) []*Session {
	sessions := make([]*Session, 0)
	for k1 := range uids {
		nodeInfo.uidSessions.RangeShard(uids[k1], func(key2 string, value interface{}) bool {
			sessions = append(sessions, value.(*Session))
			return true
		})
	}
	return sessions
}

//Get bind uid with session
func BindUid(uid string, se *Session) error {
	now := time.Now().Unix()
	node := nodeInfo.userHashRing.Get(uid)
	err := node.Extra.(*redis.Redis).Hset(userPrefix+uid, nodeInfo.transAddress, strconv.FormatInt(now, 10))
	if err != nil {
		return err
	}
	sid := se.Conn.RemoteAddr().String()
	nodeInfo.uidSessions.StoreWithPlugin(uid, sid, se, func() {
		oldUid := se.CasUid(uid)
		if oldUid != "" && oldUid != uid {
			nodeInfo.uidSessions.DeleteWithoutLock(oldUid, sid)
		}
	})
	return nil
}

//Send message to a uid
func SendToUid(uid string, path string, req interface{}) error {
	now := time.Now().Unix()
	node := nodeInfo.userHashRing.Get(uid)
	ipMap, err := node.Extra.(*redis.Redis).Hgetall(userPrefix + uid)
	if err != nil {
		mapreduce.MapVoid(func(source chan<- interface{}) {
			for key, value := range ipMap {
				expir, _ := strconv.ParseInt(value, 10, 64)
				if now > expir {
					source <- key
				}
			}
		}, func(item interface{}) {
			sid := item.(string)
			conn, ok := nodeInfo.uidSessions.Load(uid, sid)
			if ok {
				conn.(*websocket.Conn).WriteJSON(req)
			}

		})
	}
	return nil
}

//Send message to a uid
func SendToTrans(uid string, path string, req interface{}) error {
	mapreduce.MapVoid(func(source chan<- interface{}) {
		nodeInfo.transConns.Range(func(key, value interface{}) bool {
			if key == nodeInfo.transAddress {
				return true
			}
			source <- value
			return true
		})
	}, func(item interface{}) {
		item.(*websocket.Conn).WriteJSON(req)

	})
	return nil
}
