package websocket_cluster

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
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
	RedisNode struct {
		RedisConf redis.RedisConf
		Position  uint32
	}

	NodeInfo struct {
		bizRedis         *redis.Redis
		nodeConf         *NodeConf
		masterConn       *Connection
		uidSessions      *syncx.ConcurrentDoubleMap
		clientConns      goutil.Map               //External communication value is *session
		clientAddress    string                   //External communication address
		transClientConns goutil.Map               //Internal communication value is *Connection
		transAddress     string                   //Internal communication address
		timer            *timingwheel.TimingWheel //Timingwheel
		startTime        time.Time
		userHashRing     *unsafehash.Consistent //UsHash ring storage userId
		groupHashRing    *unsafehash.Consistent //UsHash ring storage groupId
		transServerConns goutil.Map
	}
)

var ()

// NewPeer creates a new peer.
func StartNode(cfg *NodeConf) {
	rds := redis.NewRedis(cfg.RedisConf.Host, cfg.RedisConf.Type, cfg.RedisConf.Pass)
	timer, err := timingwheel.NewTimingWheel(time.Second, 300, func(k, v interface{}) {
		logx.Infof("%s auth timeout", k)
		err := v.(*Connection).Conn.Close()
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
	nodeInfo := &NodeInfo{
		nodeConf:         cfg,
		bizRedis:         rds,
		uidSessions:      syncx.NewConcurrentDoubleMap(32),
		startTime:        time.Now(),
		timer:            timer,
		userHashRing:     userHashRing,
		groupHashRing:    groupHashRing,
		transClientConns: goutil.AtomicMap(),
		clientConns:      goutil.AtomicMap(),
		transServerConns: goutil.AtomicMap(),
	}
	nodeInfo.clientAddress = fmt.Sprintf("%s%s/client", cfg.Host, cfg.Path)
	nodeInfo.transAddress = fmt.Sprintf("%s%s/trans", cfg.Host, cfg.Path)

	e := echo.New()
	e.GET(fmt.Sprintf("%s/trans", cfg.Path), nodeInfo.transHandler)
	e.GET(fmt.Sprintf("%s/client", cfg.Path), nodeInfo.clientHandler)
	webGroup := e.Group(fmt.Sprintf("%s/web", cfg.Path), originMiddlewareFunc)
	cfg.router(webGroup)
	nodeInfo.SendPing()
	nodeInfo.UpdateRedis()
	go nodeInfo.ConnectToMaster(cfg)
	err = e.Start(fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		logx.Info(err)
	}
}

func OptionHandler(c echo.Context) error {
	c.Response().Header().Set("Access-Control-Allow-Origin", "*")
	c.Response().Header().Set("Access-Control-Allow-Headers", "*")
	c.String(200, "")
	return nil
}
func webHandler(c echo.Context) error {
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

func (nodeInfo *NodeInfo) transHandler(c echo.Context) error {
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
		nodeInfo.OnTransMessage(&Connection{Conn: conn}, message)
	}
	return nil
}

func (nodeInfo *NodeInfo) clientHandler(c echo.Context) error {
	connect, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	conn := &Connection{Conn: connect}
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			logx.Info(err)
		}
		return err
	}
	for {
		_, message, err := connect.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logx.Info(err)
			}
			break
		}
		logx.Info(string(message))
		nodeInfo.OnClientMessage(conn, message)
	}
	nodeInfo.clientConns.Delete(connect.RemoteAddr().String())
	return nil
}

//Heartbeat
func (nodeInfo *NodeInfo) SendPing() {
	go func() {
		for {
			time.Sleep(time.Duration(nodeInfo.nodeConf.PingInterval) * time.Second)
			// nodeInfo.masterConn.WriteMessage()
			nodeInfo.masterConn.WriteMessage(websocket.PingMessage, []byte{})
			nodeInfo.transClientConns.Range(func(k, v interface{}) bool {
				err := v.(*Connection).WriteMessage(websocket.PingMessage, []byte{})
				if err != nil {
					logx.Info(err)
				}
				return true
			})
		}
	}()
}

//Update clients num
func (nodeInfo *NodeInfo) UpdateRedis() {
	go func() {
		for {
			time.Sleep(redisInterval * time.Second)
			nodeInfo.bizRedis.Zadd(redisZsortKey, int64(nodeInfo.clientConns.Len()), nodeInfo.clientAddress)
		}
	}()
}

//Connect to master
func (nodeInfo *NodeInfo) ConnectToMaster(cfg *NodeConf) {
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

	nodeInfo.masterConn = &Connection{Conn: conn}
	auth := &Auth{
		Password:     nodeInfo.nodeConf.Password,
		TransAddress: nodeInfo.transAddress,
	}
	data := Message{
		MessageType: "auth",
		Data:        auth,
	}
	err = nodeInfo.masterConn.WriteJSON(data)
	if err != nil {
		logx.Info(err)
	}
}

//Determine if a uid is online
func (nodeInfo *NodeInfo) IsOnline(uid string) bool {
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

func (nodeInfo *NodeInfo) OnClientMessage(conn *Connection, message []byte) {
	data := make(map[string]interface{})
	err := json.Unmarshal(message, &data)
	if err != nil {
		logx.Info(err)
	}
	sid := conn.Conn.RemoteAddr().String()
	session, ok := nodeInfo.clientConns.Load(sid)
	uid := ""
	if ok {
		uid = session.(*Session).Uid
	}
	nodeInfo.nodeConf.onMessage(nodeInfo, &Context{Conn: conn, Message: message, Uid: uid})
}

func (nodeInfo *NodeInfo) OnTransMessage(conn *Connection, message []byte) {
	data := make(map[string]interface{})
	err := json.Unmarshal(message, &data)
	if err != nil {
		logx.Info(err)
	}
	v1, ok := data["message_type"]
	if !ok {
		logx.Info("message_type is nil")
	}
	switch v1 {
	case "auth":
		nodeInfo.AuthTrans(conn, data["data"].(map[string]interface{}))
	case "chat_message_list":
		receiveUid := data["receive_uid"].(string)
		nodeInfo.uidSessions.RangeNextMap(receiveUid, func(k1, k2 string, se interface{}) bool {
			err = se.(*Session).Conn.WriteJSON(data)
			return true
		})
	}
}

func (nodeInfo *NodeInfo) OnMasterMessage(message []byte) {
	data := make(map[string]interface{})
	err := json.Unmarshal(message, &data)
	if err != nil {
		logx.Info(err)
	}
	v1, ok := data["message_type"]
	if !ok {
		logx.Info("message_type is nil")
	}
	switch v1 {
	case "UpdateNodeList":
		nodeInfo.UpdateNodeList(data["data"].([]interface{}))
	}
}

// Auth the node
func (nodeInfo *NodeInfo) AuthTrans(conn *Connection, args map[string]interface{}) error {
	sid := args["trans_address"].(string)
	nodeInfo.timer.RemoveTimer(sid) //Cancel timeingwheel task
	if args["password"].(string) != nodeInfo.nodeConf.Password {
		logx.Infof("Connect:%s,Wrong password:%s", sid, args["password"].(string))
		conn.Conn.Close()
		return fmt.Errorf("auth faild")
	}
	nodeInfo.transClientConns.Store(sid, conn)
	return nil
}

// Auth the node
func (nodeInfo *NodeInfo) AuthClient(conn *Connection, uid string) error {
	sid := conn.Conn.RemoteAddr().String()
	nodeInfo.timer.RemoveTimer(sid) //Cancel timeingwheel task
	session := &Session{Conn: conn, Uid: uid}
	nodeInfo.clientConns.Store(sid, session)
	nodeInfo.BindUid(uid, session)
	return nil
}

// Add handles addition request
func (nodeInfo *NodeInfo) UpdateNodeList(nodeList []interface{}) error {
	for _, value := range nodeList {
		if value.(string) == nodeInfo.transAddress {
			continue
		}
		_, ok := nodeInfo.transServerConns.LoadOrStore(value.(string), "")
		if ok {
			continue
		}
		connect, _, err := websocket.DefaultDialer.Dial(value.(string), nil)
		conn := &Connection{Conn: connect}
		if err != nil {
			logx.Info("dial:", err)
			nodeInfo.transServerConns.Delete(value.(string))
			continue
		}

		auth := &Auth{
			Password:     nodeInfo.nodeConf.Password,
			TransAddress: nodeInfo.transAddress,
		}
		data := Message{
			MessageType: "auth",
			Data:        auth,
		}
		err = conn.WriteJSON(data)
		if err != nil {
			logx.Info(err)
		}
		go func(conn *Connection) {
			defer conn.Conn.Close()
			for {
				_, message, err := conn.Conn.ReadMessage()
				if err != nil {
					log.Println("read:", err)

					return
				}
				logx.Infof("recv: %s", message)
				nodeInfo.OnTransMessage(conn, message)
			}
		}(conn)
	}

	return nil
}

// func (nodeInfo *NodeInfo) JoinGroup(gid, uid string) error {
// 	now := time.Now().Unix()
// 	node := nodeInfo.groupHashRing.Get(gid)
// 	err := node.Extra.(*redis.Redis).Hset(groupPrefix+gid, uid, strconv.FormatInt(now, 10))
// 	if err != nil {
// 		logx.Info(err)
// 	}
// 	return nil
// }

// func (nodeInfo *NodeInfo) LeaveGroup(gid, uid string) error {
// 	node := nodeInfo.groupHashRing.Get(gid)
// 	_, err := node.Extra.(*redis.Redis).Hdel(groupPrefix+gid, uid)
// 	if err != nil {
// 		logx.Info(err)
// 	}
// 	return nil
// }

// func (nodeInfo *NodeInfo) SendToGroup(gid string, req interface{}) error {
// 	uids := nodeInfo.GroupOnline(gid)
// 	mapreduce.MapVoid(func(source chan<- interface{}) {
// 		for k1, _ := range uids {
// 			nodeInfo.uidSessions.RangeShard(uids[k1], func(key2 string, value interface{}) bool {
// 				source <- value
// 				return true
// 			})
// 		}
// 	}, func(item interface{}) {
// 		se := item.(*Session)
// 		err := se.Conn.WriteJSON(req)
// 		if err != nil {
// 			logx.Info(err)
// 		}
// 	})
// 	return nil
// }

//Get online users in the group
// func (nodeInfo *NodeInfo) GroupOnline(gid string) []string {
// 	now := time.Now().Unix()
// 	node := nodeInfo.groupHashRing.Get(gid)
// 	uids := make([]string, 0)
// 	addrMap, err := node.Extra.(*redis.Redis).Hgetall(groupPrefix + gid)
// 	if err == nil {
// 		return uids
// 	}
// 	for key, value := range addrMap {
// 		old, _ := strconv.ParseInt(value, 10, 64)
// 		if now < old {
// 			uids = append(uids, key)
// 		}
// 	}
// 	return uids
// }

func (nodeInfo *NodeInfo) SendToUids(uidList []string, req interface{}) error {
	uids := nodeInfo.UidsOnline(uidList)
	mapreduce.MapVoid(func(source chan<- interface{}) {
		for k1, _ := range uids {
			nodeInfo.uidSessions.RangeNextMap(uids[k1], func(key1 string, key2 string, value interface{}) bool {
				source <- value
				return true
			})
		}
	}, func(item interface{}) {
		fmt.Printf("%#v\n", item)
		se := item.(*Session)
		req.(map[string]interface{})["receive_uid"] = se.Uid
		err := se.Conn.WriteJSON(req)
		if err != nil {
			logx.Info(err)
		}
	})
	return nil
}

type NodeMap struct {
	node *unsafehash.Node
	uids []string
}

// Get online users in the group
func (nodeInfo *NodeInfo) UidsOnline(uids []string) []string {
	// now := time.Now().Unix()
	nodes := make(map[string]*NodeMap, 0)
	for k1 := range uids {
		node := nodeInfo.userHashRing.Get(uids[k1])
		if _, ok := nodes[node.Id]; ok {
			nodes[node.Id].uids = append(nodes[node.Id].uids, uids[k1])
		} else {
			nodes[node.Id] = &NodeMap{
				node: node,
				uids: []string{uids[k1]},
			}
		}
	}
	// for k1 := range nodes {
	// 	nodeMap := nodes[k1]
	// 	arr, err := nodeMap.node.Extra.(*redis.Redis).Mget(nodeMap.uids...)

	// }

	// for key, value := range arr {
	// 	old, _ := strconv.ParseInt(value, 10, 64)
	// 	if now < old {
	// 		uids = append(uids, key)
	// 	}
	// }
	return uids
}

//Get online users in the group
func (nodeInfo *NodeInfo) GetSessionsByUids(uids []string) []*Session {
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
func (nodeInfo *NodeInfo) BindUid(uid string, se *Session) error {
	now := time.Now().Unix()
	node := nodeInfo.userHashRing.Get(uid)
	err := node.Extra.(*redis.Redis).Hset(userPrefix+uid, nodeInfo.transAddress, strconv.FormatInt(now, 10))
	if err != nil {
		return err
	}
	sid := se.Conn.Conn.RemoteAddr().String()
	nodeInfo.uidSessions.StoreWithPlugin(uid, sid, se, func() {
		oldUid := se.CasUid(uid)
		if oldUid != "" && oldUid != uid {
			nodeInfo.uidSessions.DeleteWithoutLock(oldUid, sid)
		}
	})
	return nil
}

//Send message to a uid
func (nodeInfo *NodeInfo) SendToUid(uid string, req interface{}) error {
	now := time.Now().Unix()
	node := nodeInfo.userHashRing.Get(uid)
	ipMap, err := node.Extra.(*redis.Redis).Hgetall(userPrefix + uid)
	if err == nil {
		mapreduce.MapVoid(func(source chan<- interface{}) {
			for key, value := range ipMap {
				expir, _ := strconv.ParseInt(value, 10, 64)
				expir += 600
				if now < expir {
					source <- key
				}
			}
		}, func(item interface{}) {
			ip := item.(string)
			if ip == nodeInfo.transAddress {
				// se, ok := nodeInfo.uidSessions.Load(uid, ip)
				nodeInfo.uidSessions.RangeNextMap(uid, func(k1, k2 string, se interface{}) bool {
					err = se.(*Session).Conn.WriteJSON(req)
					return true
				})
				// if ok {
				// 	err = se.(*Session).Conn.WriteJSON(req)
				// 	logx.Info(err)
				// }

			} else {
				connect, ok := nodeInfo.transClientConns.Load(ip)
				if ok {
					err = connect.(*Connection).WriteJSON(req)
					if err != nil {
						logx.Info(err)
					}
				} else {
					fmt.Printf("ip2:%#v\n", ip)
				}

			}

		})
	}
	return nil
}

//Send message to a uid
func (nodeInfo *NodeInfo) SendToTrans(uid string, path string, req interface{}) error {
	mapreduce.MapVoid(func(source chan<- interface{}) {
		nodeInfo.transClientConns.Range(func(key, value interface{}) bool {
			if key == nodeInfo.transAddress {
				return true
			}
			source <- value
			return true
		})
	}, func(item interface{}) {
		item.(*Connection).WriteJSON(req)

	})
	return nil
}
