package websocket_cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cast"
	"github.com/weblazy/core/consistenthash/unsafehash"
	"github.com/weblazy/core/mapreduce"
	"github.com/weblazy/core/syncx"
	"github.com/weblazy/core/timingwheel"
	"github.com/weblazy/easy/utils/logx"
	"github.com/weblazy/goutil"

	"time"
)

type (
	RedisNode struct {
		RedisConf RedisConf
		Position  uint32
	}

	NodeInfo struct {
		bizRedis      *redis.Client
		nodeConf      *NodeConf
		uidSessions   *syncx.ConcurrentDoubleMap
		clientConns   goutil.Map               //External communication value is *session
		clientAddress string                   //External communication address
		transAddress  string                   //Internal communication address
		timer         *timingwheel.TimingWheel //Timingwheel
		startTime     time.Time
		userHashRing  *unsafehash.Consistent //UsHash ring storage userId
		groupHashRing *unsafehash.Consistent //UsHash ring storage groupId
		transConns    goutil.Map
	}
)

var ()

// NewPeer creates a new peer.
func StartNode(cfg *NodeConf) (*NodeInfo, error) {
	rds := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisConf.Addr,
		Password: cfg.RedisConf.Password,
		DB:       int(cfg.RedisConf.DB),
	})
	timer, err := timingwheel.NewTimingWheel(time.Second, 300, func(k, v interface{}) {
		logx.Infof("%s auth timeout", k)
		err := v.(*Connection).Conn.Close()
		if err != nil {
			logx.Info(err)
		}
	})
	if err != nil {
		return nil, err
	}
	userHashRing := unsafehash.NewConsistent(cfg.RedisMaxCount)
	for _, value := range cfg.RedisNodeList {
		rdsObj := redis.NewClient(&redis.Options{
			Addr:     value.RedisConf.Addr,
			Password: value.RedisConf.Password,
			DB:       int(value.RedisConf.DB),
		})
		userHashRing.Add(unsafehash.NewNode(value.RedisConf.Addr, value.Position, rdsObj))
	}
	groupHashRing := unsafehash.NewConsistent(cfg.RedisMaxCount)
	for _, value := range cfg.RedisNodeList {
		rdsObj := redis.NewClient(&redis.Options{
			Addr:     value.RedisConf.Addr,
			Password: value.RedisConf.Password,
			DB:       int(value.RedisConf.DB),
		})
		groupHashRing.Add(unsafehash.NewNode(value.RedisConf.Addr, value.Position, rdsObj))
	}
	nodeInfo := &NodeInfo{
		nodeConf:      cfg,
		bizRedis:      rds,
		uidSessions:   syncx.NewConcurrentDoubleMap(32),
		startTime:     time.Now(),
		timer:         timer,
		userHashRing:  userHashRing,
		groupHashRing: groupHashRing,
		clientConns:   goutil.AtomicMap(),
		transConns:    goutil.AtomicMap(),
	}
	nodeInfo.clientAddress = fmt.Sprintf("%s%s/client", cfg.Host, cfg.Path)
	nodeInfo.transAddress = fmt.Sprintf("%s%s/trans", cfg.Host, cfg.Path)

	e := echo.New()
	e.GET(fmt.Sprintf("%s/trans", cfg.Path), nodeInfo.transHandler)
	e.GET(fmt.Sprintf("%s/client", cfg.Path), nodeInfo.clientHandler)
	webGroup := e.Group(fmt.Sprintf("%s/web", cfg.Path), originMiddlewareFunc)
	cfg.router(webGroup)
	go func() {
		err = e.Start(fmt.Sprintf(":%d", cfg.Port))
		if err != nil {
			panic(err)
		}
	}()
	nodeInfo.UpdateRedis()
	nodeInfo.SendPing()
	return nodeInfo, nil
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

	for {
		_, msg, err := connect.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logx.Info(err)
			}
			break
		}
		logx.Info(string(msg))
		nodeInfo.OnTransMsg(&Connection{Conn: connect}, msg)
	}
	return nil
}

func (nodeInfo *NodeInfo) clientHandler(c echo.Context) error {
	connect, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	defer func() {
		if connect != nil {
			connect.Close()
		}
	}()
	conn := &Connection{Conn: connect}
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			logx.Info(err)
		}
		return err
	}
	for {
		_, msg, err := connect.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logx.Info(err)
			}
			break
		}
		logx.Info(string(msg))
		nodeInfo.OnClientMsg(conn, msg)
	}
	nodeInfo.clientConns.Delete(connect.RemoteAddr().String())
	return nil
}

//Heartbeat
func (nodeInfo *NodeInfo) SendPing() {
	go func() {
		for {
			time.Sleep(time.Duration(nodeInfo.nodeConf.PingInterval) * time.Second)

			err := nodeInfo.bizRedis.HSet(context.Background(), nodeInfo.transAddress, cast.ToString(time.Now().Unix())).Err()
			if err != nil {
				logx.Info(err)
			}
			nodeMap, err := nodeInfo.bizRedis.HGetAll(context.Background(), "transAddress").Result()
			if err != nil {
				logx.Info(err)
			}
			err = nodeInfo.UpdateNodeList(nodeMap)
			if err != nil {
				logx.Info(err)
			}
			nodeInfo.transConns.Range(func(k, v interface{}) bool {
				conn, ok := v.(*Connection)
				if !ok {
					return true
				}
				err := conn.WriteMsg(websocket.PingMessage, []byte{})
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
			err := nodeInfo.bizRedis.ZAdd(context.Background(), redisZsortKey, &redis.Z{
				Score:  float64(nodeInfo.clientConns.Len()),
				Member: nodeInfo.clientAddress}).Err()
			if err != nil {
				logx.Info(err)
			}
		}
	}()
}

//Determine if a uid is online
func (nodeInfo *NodeInfo) IsOnline(uid string) bool {
	now := time.Now().Unix()
	node := nodeInfo.userHashRing.Get(uid)
	addrArr, err := node.Extra.(*redis.Client).ZRangeByScore(context.Background(), userPrefix+uid, &redis.ZRangeBy{Min: cast.ToString(now - 600)}).Result()
	if err != nil {
		logx.Info(err)
		return false
	}
	if len(addrArr) > 0 {
		return true
	}
	return false
}

func (nodeInfo *NodeInfo) OnClientMsg(conn *Connection, msg []byte) {
	sid := conn.Conn.RemoteAddr().String()
	session, ok := nodeInfo.clientConns.Load(sid)
	uid := ""
	if ok {
		uid = session.(*Session).Uid
	}
	nodeInfo.nodeConf.onMsg(nodeInfo, &Context{Conn: conn, Msg: msg, Uid: uid})
}

func (nodeInfo *NodeInfo) OnClientPing(uid string, now int64) error {
	node := nodeInfo.userHashRing.Get(uid)
	err := node.Extra.(*redis.Client).ZAdd(context.Background(), userPrefix+uid, &redis.Z{Score: cast.ToFloat64(now), Member: nodeInfo.transAddress}).Err()
	return err
}

func (nodeInfo *NodeInfo) OnTransMsg(conn *Connection, msg []byte) {
	data := make(map[string]interface{})
	err := json.Unmarshal(msg, &data)
	if err != nil {
		logx.Info(err)
		return
	}
	v1, ok := data["msg_type"]
	if !ok {
		logx.Info("msg_type is nil")
		return
	}
	switch v1 {
	case "auth":
		nodeInfo.AuthTrans(conn, data["data"].(map[string]interface{}))
	default:
		receiveUid := data["receive_uid"].(string)
		delete(data, "receive_uid")
		nodeInfo.uidSessions.RangeNextMap(receiveUid, func(k1, k2 string, se interface{}) bool {
			err = se.(*Session).Conn.WriteJSON(data)
			if err != nil {
				logx.Info(err)
			}
			return true
		})
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
	nodeInfo.transConns.Store(sid, conn)
	return nil
}

// Auth the node
func (nodeInfo *NodeInfo) AuthClient(conn *Connection, uid string) error {
	sid := conn.Conn.RemoteAddr().String()
	nodeInfo.timer.RemoveTimer(sid) //Cancel timeingwheel task
	session := &Session{Conn: conn, Uid: uid}
	nodeInfo.clientConns.Store(sid, session)
	return nodeInfo.BindUid(uid, session)
}

// Add handles addition request
func (nodeInfo *NodeInfo) UpdateNodeList(nodeMap map[string]string) error {
	for key, value := range nodeMap {
		if cast.ToInt64(value)+600 < time.Now().Unix() {
			continue
		}
		//数字小的连接数字大的
		if strings.Compare(nodeInfo.transAddress, key) >= 0 {
			continue
		}
		//已经建立连接
		_, ok := nodeInfo.transConns.LoadOrStore(key, "")
		if ok {
			continue
		}

		connect, _, err := websocket.DefaultDialer.Dial(key, nil)
		if err != nil {
			logx.Info("dial:", err)
			nodeInfo.transConns.Delete(key)
			continue
		}

		auth := &Auth{
			Password:     nodeInfo.nodeConf.Password,
			TransAddress: nodeInfo.transAddress,
		}
		data := Msg{
			MsgType: "auth",
			Data:    auth,
		}
		conn := &Connection{Conn: connect}
		err = conn.WriteJSON(data)
		if err != nil {
			logx.Info(err)
		}
		nodeInfo.transConns.Store(key, conn)
		go func(key string, conn *Connection) {
			defer func(key string, conn *Connection) {
				nodeInfo.transConns.Delete(key)
				conn.Conn.Close()
			}(key, conn)
			for {
				_, msg, err := conn.Conn.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						logx.Info(err)
					}
					break
				}
				logx.Info(string(msg))
				nodeInfo.OnTransMsg(conn, msg)
			}
		}(key, conn)
	}
	return nil
}

func (nodeInfo *NodeInfo) SendToUids(uidList []string, req map[string]interface{}) error {
	if req == nil {
		return fmt.Errorf("message is nil")
	}
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
		newMap := make(map[string]interface{})
		for k, v := range req {
			newMap[k] = v
		}
		newMap["receive_uid"] = se.Uid
		err := se.Conn.WriteJSON(newMap)
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
	sid := se.Conn.Conn.RemoteAddr().String()
	nodeInfo.uidSessions.StoreWithPlugin(uid, sid, se, func() {
		oldUid := se.CasUid(uid)
		if oldUid != "" && oldUid != uid {
			nodeInfo.uidSessions.DeleteWithoutLock(oldUid, sid)
		}
	})

	now := time.Now().Unix()
	node := nodeInfo.userHashRing.Get(uid)
	fmt.Printf("%#v\n", userPrefix+uid)
	fmt.Printf("%#v\n", nodeInfo.transAddress)
	err := node.Extra.(*redis.Client).ZAdd(context.Background(), userPrefix+uid, &redis.Z{Score: cast.ToFloat64(now), Member: nodeInfo.transAddress}).Err()
	if err != nil {
		return err
	}

	return nil
}

//Send message to a uid
func (nodeInfo *NodeInfo) SendToUid(uid string, req map[string]interface{}) error {
	if req == nil {
		return fmt.Errorf("message is nil")
	}
	now := time.Now().Unix()
	node := nodeInfo.userHashRing.Get(uid)
	ipArr, err := node.Extra.(*redis.Client).ZRangeByScore(context.Background(), userPrefix+uid, &redis.ZRangeBy{Min: cast.ToString(now - 600)}).Result()
	if err == nil {
		mapreduce.MapVoid(func(source chan<- interface{}) {
			for key, _ := range ipArr {
				source <- ipArr[key]
			}
		}, func(item interface{}) {
			ip := item.(string)
			if ip == nodeInfo.transAddress {
				nodeInfo.uidSessions.RangeNextMap(uid, func(k1, k2 string, se interface{}) bool {
					err = se.(*Session).Conn.WriteJSON(req)
					if err != nil {
						logx.Info(err)
					}
					return true
				})
			} else {
				connect, ok := nodeInfo.transConns.Load(ip)
				if ok {
					conn, ok := connect.(*Connection)
					if !ok {
						return
					}
					req["receive_uid"] = uid
					err = conn.WriteJSON(req)
					if err != nil {
						logx.Info(err)
					}
				} else {
					logx.Info(fmt.Printf("trans:%#v不在线", ip))
					return

				}

			}

		})
	}
	return nil
}

//Send message to a uid
func (nodeInfo *NodeInfo) SendToTrans(uid string, path string, req interface{}) error {
	mapreduce.MapVoid(func(source chan<- interface{}) {
		nodeInfo.transConns.Range(func(key, value interface{}) bool {
			if key == nodeInfo.transAddress {
				return true
			}
			source <- value
			return true
		})
	}, func(item interface{}) {
		conn, ok := item.(*Connection)
		if !ok {
			return
		}
		conn.WriteJSON(req)
	})
	return nil
}
