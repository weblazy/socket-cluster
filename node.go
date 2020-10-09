package websocket_cluster

import (
	"encoding/json"
	"fmt"
	"log"

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
		Host          string
		Path          string
		RedisNodeList []RedisNode
		RedisConf     redis.RedisConf
		RedisMaxCount uint32
		Port          int64
		MasterAddress string //Master address
		Password      string //Password for auth when connect to master
		PingInterval  int64  //Heartbeat interval
	}
	RedisNode struct {
		RedisConf redis.RedisConf
		Position  uint32
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
)

var (
	nodeInfo *NodeInfo
)

// NewPeer creates a new peer.
func NewNodeConf(host, path, masterAddress string, redisConf redis.RedisConf) *NodeConf {
	return &NodeConf{
		Host:          host,
		Path:          path,
		Port:          9528,
		RedisConf:     redisConf,
		MasterAddress: masterAddress,
		Password:      defaultPassword,
		PingInterval:  defaultPingInterval,
	}
}

func (conf *NodeConf) WithPassword(password string) *NodeConf {
	conf.Password = password
	return conf
}

func (conf *NodeConf) WithPort(port int64) *NodeConf {
	conf.Port = port
	return conf
}

func (conf *NodeConf) WithPing(pingInterval int64) *NodeConf {
	conf.PingInterval = pingInterval
	return conf
}

// NewPeer creates a new peer.
func StartNode(cfg *NodeConf) {
	rds := redis.NewRedis(cfg.RedisConf.Host, cfg.RedisConf.Type, cfg.RedisConf.Pass)
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
	userHashRing := unsafehash.NewConsistent(cfg.RedisMaxCount)
	for _, value := range cfg.RedisNodeList {
		rdsObj := redis.NewRedis(value.RedisConf.Host, value.RedisConf.Type, value.RedisConf.Pass)
		userHashRing.Add(unsafehash.NewNode(value.RedisConf.Host, value.Position, rdsObj))
	}
	groupHashRing := unsafehash.NewConsistent(cfg.RedisMaxCount)
	for _, value := range cfg.RedisNodeList {
		rdsObj := redis.NewRedis(value.RedisConf.Host, value.RedisConf.Type, value.RedisConf.Pass)
		groupHashRing.Add(unsafehash.NewNode(value.RedisConf.Host, value.Position, rdsObj))
	}
	nodeInfo = &NodeInfo{
		nodeConf:      cfg,
		bizRedis:      rds,
		uidSessions:   syncx.NewConcurrentDoubleMap(32),
		startTime:     time.Now(),
		timer:         timer,
		userHashRing:  userHashRing,
		groupHashRing: groupHashRing,
		transConns:    goutil.AtomicMap(),
		clientConns:   goutil.AtomicMap(),
	}
	nodeInfo.clientAddress = fmt.Sprintf("%s%s/client", cfg.Host, cfg.Path)
	nodeInfo.transAddress = fmt.Sprintf("%s%s/trans", cfg.Host, cfg.Path)

	e := echo.New()
	e.GET(fmt.Sprintf("%s/trans", cfg.Path), transHandler)
	e.GET(fmt.Sprintf("%s/client", cfg.Path), clientHandler)
	SendPing()
	UpdateRedis()
	go ConnectToMaster(cfg)
	err = e.Start(fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		logx.Info(err)
	}
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
		nodeInfo.OnClientMessage(conn, message)
	}
	nodeInfo.clientConns.Delete(conn.RemoteAddr().String())
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
	conn, _, err := websocket.DefaultDialer.Dial(cfg.MasterAddress, nil)
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
			nodeInfo.OnMasterMessage(message)
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

func (nodeInfo *NodeInfo) OnClientMessage(conn *websocket.Conn, message []byte) {
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
		AuthClient(conn, data["data"].(map[string]interface{}))
	}
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
		nodeInfo.UpdateNodeList(data["data"].([]interface{}))
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
func (nodeInfo *NodeInfo) UpdateNodeList(nodeList []interface{}) error {
	for _, value := range nodeList {
		conn, _, err := websocket.DefaultDialer.Dial(value.(string), nil)
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
