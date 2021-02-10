package node

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cast"
	"github.com/weblazy/core/consistenthash/unsafehash"
	"github.com/weblazy/core/mapreduce"
	"github.com/weblazy/easy/utils/logx"
	"github.com/weblazy/easy/utils/syncx"
	"github.com/weblazy/easy/utils/timingwheel"
	"github.com/weblazy/goutil"
)

type (

	// Node communication node
	Node struct {
		// bizRedis redis client store node info
		bizRedis         *redis.Client
		nodeConf         *NodeConf
		clientIdSessions *syncx.ConcurrentDoubleMap
		// key: socket address
		// value: *session
		clientConns   goutil.Map
		clientAddress string                   // External communication address
		transAddress  string                   // Internal communication address
		timer         *timingwheel.TimingWheel // Timingwheel
		startTime     time.Time                // start time
		userHashRing  *unsafehash.Consistent   // userHashRing ring storage userId
		// key: node transAddress
		// value: *session
		transConns    goutil.Map //
		nodeTimeout   int64      // node heartbeat timeout time
		clientTimeout int64      // client heartbeat timeout time
	}
)

var ()

// NewPeer creates a new peer.
func StartNode(cfg *NodeConf) (*Node, error) {
	rds := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisConf.Addr,
		Password: cfg.RedisConf.Password,
		DB:       int(cfg.RedisConf.DB),
	})

	timer, err := timingwheel.NewTimingWheel(time.Second, 30, func(k, v interface{}) {
		logx.Infof("%s auth timeout", k)
		if v.(*Connection).Conn != nil {
			err := v.(*Connection).Conn.Close()
			if err != nil {
				logx.Info(err)
			}
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

	this := &Node{
		nodeConf:         cfg,
		bizRedis:         rds,
		clientIdSessions: syncx.NewConcurrentDoubleMap(32),
		startTime:        time.Now(),
		timer:            timer,
		userHashRing:     userHashRing,
		clientConns:      goutil.AtomicMap(),
		transConns:       goutil.AtomicMap(),
		nodeTimeout:      cfg.NodePingInterval * 3,
		clientTimeout:    cfg.ClientPingInterval * 3,
	}
	this.transAddress = fmt.Sprintf("%s%s/trans", cfg.Host, cfg.TransPath)
	this.clientAddress = fmt.Sprintf("%s%s/client", cfg.Host, cfg.ClientPath)
	if cfg.echoObj != nil {
		cfg.echoObj.GET(fmt.Sprintf("%s/trans", cfg.TransPath), this.transHandler)
		cfg.echoObj.GET(fmt.Sprintf("%s/client", cfg.ClientPath), this.clientHandler)
		webGroup := cfg.echoObj.Group(fmt.Sprintf("%s/web", cfg.ClientPath), originMiddlewareFunc)
		cfg.router(webGroup)
	} else {
		e := echo.New()
		e.GET(fmt.Sprintf("%s/trans", cfg.TransPath), this.transHandler)
		e.GET(fmt.Sprintf("%s/client", cfg.ClientPath), this.clientHandler)
		webGroup := e.Group(fmt.Sprintf("%s/web", cfg.ClientPath), originMiddlewareFunc)
		cfg.router(webGroup)
		go func() {
			err = e.Start(fmt.Sprintf(":%d", cfg.Port))
			if err != nil {
				panic(err)
			}
		}()
	}

	this.SendPing()
	if this.nodeConf.Host == "" {
		this.Consumer()
	}
	return this, nil
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

// transHandler deal node connection
func (this *Node) transHandler(c echo.Context) error {
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
	conn := &Connection{Conn: connect}
	this.timer.SetTimer(connect.RemoteAddr().String(), conn, authTime)
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
func (this *Node) clientHandler(c echo.Context) error {
	connect, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	defer func() {
		key := connect.RemoteAddr().String()
		v1, ok := this.clientConns.Load(key)
		if ok {
			clientId := v1.(*Session).ClientId
			if clientId != "" {
				this.clientIdSessions.Delete(clientId, key)
			}
		}
		this.clientConns.Delete(key)
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
	this.timer.SetTimer(connect.RemoteAddr().String(), conn, authTime)
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

// SendPing send node Heartbeat
func (this *Node) SendPing() {
	go func() {
		for {
			time.Sleep(time.Duration(this.nodeConf.NodePingInterval) * time.Second)
			nodeInfo := map[string]interface{}{
				"client_address": this.clientAddress,
				"client_count":   this.clientConns.Len(),
				"timestamp":      time.Now().Unix(),
			}
			nodeInfoByte, err := json.Marshal(nodeInfo)
			if err != nil {
				logx.Info(err)
			}
			err = this.bizRedis.HSet(context.Background(), NodeAddress, this.transAddress, string(nodeInfoByte)).Err()
			if err != nil {
				logx.Info(err)
			}
			nodeMap, err := this.bizRedis.HGetAll(context.Background(), NodeAddress).Result()
			if err != nil {
				logx.Info(err)
			}
			if this.nodeConf.Host != "" {
				err = this.UpdateNodeList(nodeMap)
				if err != nil {
					logx.Info(err)
				}
				this.transConns.Range(func(k, v interface{}) bool {
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
			} else {
				now := time.Now().Unix()
				for k1 := range nodeMap {
					nodeInfo := make(map[string]interface{})
					err := json.Unmarshal([]byte(nodeMap[k1]), &nodeInfo)
					if err != nil {
						logx.Info(err)
					}
					if cast.ToInt64(nodeInfo["timestamp"])+this.nodeTimeout < now {
						err := this.bizRedis.HDel(context.Background(), NodeAddress, k1).Err()
						if err != nil {
							logx.Info(err)
						}
					}
				}
			}

		}
	}()
}

//Update clients num
// func (this *Node) UpdateRedis() {
// 	go func() {
// 		for {
// 			time.Sleep(redisInterval * time.Second)
// 			err := this.bizRedis.ZAdd(context.Background(), NodeAddress, &redis.Z{
// 				Score:  float64(this.clientConns.Len()),
// 				Member: this.clientAddress}).Err()
// 			if err != nil {
// 				logx.Info(err)
// 			}
// 		}
// 	}()
// }

// IsOnline determine if a clientId is online
func (this *Node) IsOnline(clientId string) bool {
	now := time.Now().Unix()
	redisNode := this.userHashRing.Get(clientId)
	addrArr, err := redisNode.Extra.(*redis.Client).ZRangeByScore(context.Background(), clientPrefix+clientId, &redis.ZRangeBy{Min: cast.ToString(now - this.clientTimeout), Max: "+inf"}).Result()
	if err != nil {
		logx.Info(err)
		return false
	}
	if len(addrArr) > 0 {
		return true
	}
	return false
}

// OnClientMsg deal client message
func (this *Node) OnClientMsg(conn *Connection, msg []byte) {
	sid := conn.Conn.RemoteAddr().String()
	session, ok := this.clientConns.Load(sid)
	clientId := ""
	if ok {
		clientId = session.(*Session).ClientId
	}
	this.nodeConf.onMsg(&Context{Conn: conn, Msg: msg, ClientId: clientId})
}

// OnClientPing receive client heartbeat
func (this *Node) OnClientPing(clientId string) error {
	redisNode := this.userHashRing.Get(clientId)
	now := time.Now().Unix()
	err := redisNode.Extra.(*redis.Client).ZAdd(context.Background(), clientPrefix+clientId, &redis.Z{Score: cast.ToFloat64(now), Member: this.transAddress}).Err()
	if err != nil {
		return err
	}
	err = redisNode.Extra.(*redis.Client).Expire(context.Background(), clientPrefix+clientId, time.Duration(this.clientTimeout)*time.Second).Err()
	return err
}

// OnTransMsg handle internal communication node messages
func (this *Node) OnTransMsg(conn *Connection, msg []byte) {
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
		err = this.AuthTrans(conn, data["data"].(map[string]interface{}))
		if err != nil {
			logx.Info(err)
		}
	default:
		if receiveClientId, ok := data["receive_client_id"].(string); ok {
			delete(data, "receive_client_id")
			this.clientIdSessions.RangeNextMap(receiveClientId, func(k1, k2 string, se interface{}) bool {
				err = se.(*Session).Conn.WriteJSON(data)
				if err != nil {
					logx.Info(err)
				}
				return true
			})
		} else if receiveClientIds, ok := data["receive_client_ids"].([]string); ok {
			delete(data, "receive_client_ids")
			for k1 := range receiveClientIds {
				receiveClientId := receiveClientIds[k1]
				this.clientIdSessions.RangeNextMap(receiveClientId, func(k1, k2 string, se interface{}) bool {
					err = se.(*Session).Conn.WriteJSON(data)
					if err != nil {
						logx.Info(err)
					}
					return true
				})
			}
		}
	}
}

// AuthTrans Auth the node
func (this *Node) AuthTrans(conn *Connection, args map[string]interface{}) error {
	sid := args["trans_address"].(string)
	this.timer.RemoveTimer(conn.Conn.RemoteAddr().String()) //Cancel timeingwheel task
	if args["password"].(string) != this.nodeConf.Password {
		logx.Infof("Connect:%s,Wrong password:%s", sid, args["password"].(string))
		conn.Conn.Close()
		return fmt.Errorf("auth faild")
	}
	this.transConns.Store(sid, conn)
	return nil
}

// AuthClient Auth the node
func (this *Node) AuthClient(conn *Connection, clientId string) error {
	sid := conn.Conn.RemoteAddr().String()
	this.timer.RemoveTimer(sid) //Cancel timeingwheel task
	session := &Session{Conn: conn, ClientId: clientId}
	this.clientConns.Store(sid, session)
	return this.BindClientId(clientId, session)
}

// UpdateNodeList Add handles addition request
func (this *Node) UpdateNodeList(nodeMap map[string]string) error {
	now := time.Now().Unix()
	for key := range nodeMap {
		if key == this.transAddress {
			continue
		}
		nodeInfo := make(map[string]interface{})
		err := json.Unmarshal([]byte(nodeMap[key]), &nodeInfo)
		if err != nil {
			logx.Info(err)
		}
		if cast.ToInt64(nodeInfo["timestamp"])+this.nodeTimeout < now {
			continue
		}

		transAddress := key
		// 数字小的连接数字大的
		// if strings.Compare(this.transAddress, transAddress) >= 0 {
		// 	continue
		// }
		//已经建立连接
		_, ok := this.transConns.LoadOrStore(transAddress, "")
		if ok {
			continue
		}

		connect, _, err := websocket.DefaultDialer.Dial(transAddress, nil)
		if err != nil {
			logx.Info("dial:", err)
			this.transConns.Delete(transAddress)
			continue
		}

		auth := &Auth{
			Password:     this.nodeConf.Password,
			TransAddress: this.transAddress,
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
		this.transConns.Store(transAddress, conn)
		go func(transAddress string, conn *Connection) {
			defer func(transAddress string, conn *Connection) {
				this.transConns.Delete(transAddress)
				conn.Conn.Close()
			}(transAddress, conn)
			for {
				_, msg, err := conn.Conn.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						logx.Info(err)
					}
					break
				}
				logx.Info(string(msg))
				this.OnTransMsg(conn, msg)
			}
		}(transAddress, conn)
	}
	return nil
}

// SendToClientIds Sending messages to multiple clients
func (this *Node) SendToClientIds(clientIds []string, req map[string]interface{}) error {
	if req == nil {
		return fmt.Errorf("message is nil")
	}
	localClientIds := make([]string, 0)
	nodes := make(map[string]*NodeMap)
	rangeTime := cast.ToString(time.Now().Unix() - this.clientTimeout)
	for k1 := range clientIds {
		redisNode := this.userHashRing.Get(clientIds[k1])
		if _, ok := nodes[redisNode.Id]; ok {
			nodes[redisNode.Id].clientIds = append(nodes[redisNode.Id].clientIds, clientIds[k1])
		} else {
			nodes[redisNode.Id] = &NodeMap{
				node:      redisNode,
				clientIds: []string{clientIds[k1]},
			}
		}
	}
	otherMap := make(map[string][]string)

	for k1 := range nodes {
		nodeMap := nodes[k1]
		pipe := nodeMap.node.Extra.(*redis.Client).Pipeline()
		for k2 := range nodeMap.clientIds {
			pipe.ZRangeByScore(context.Background(), clientPrefix+nodeMap.clientIds[k2], &redis.ZRangeBy{Min: rangeTime, Max: "+inf"}).Result()
		}
		cmders, err := pipe.Exec(context.Background())
		if err != nil {
			logx.Info(cmders, err)
		}
		for k3, cmder := range cmders {
			cmd := cmder.(*redis.StringSliceCmd)
			strMap, err := cmd.Result()
			if err != nil {
				logx.Info(err)
			} else {
				for k4 := range strMap {
					if strMap[k4] == this.transAddress {
						localClientIds = append(localClientIds, nodeMap.clientIds[k3])
					} else {
						if _, ok := otherMap[strMap[k4]]; ok {
							otherMap[strMap[k4]] = append(otherMap[strMap[k4]], nodeMap.clientIds[k3])
						} else {
							otherMap[strMap[k4]] = []string{nodeMap.clientIds[k3]}
						}
					}
				}

			}
		}
	}
	// Concurrent sends to other nodes
	mapreduce.MapVoid(func(source chan<- interface{}) {
		for k1 := range otherMap {
			source <- &BatchData{ip: k1, clientIds: otherMap[k1]}
		}
	}, func(item interface{}) {
		batchData := item.(*BatchData)
		newMap := make(map[string]interface{})
		for k, v := range req {
			newMap[k] = v
		}
		newMap["receive_client_ids"] = batchData.clientIds
		reqByte, err := json.Marshal(newMap)
		if err != nil {
			logx.Info(err)
		}
		if this.nodeConf.Host == "" {
			err = this.bizRedis.Publish(context.Background(), batchData.ip, string(reqByte)).Err()
			if err != nil {
				logx.Info(err)
			}
		} else {
			connect, ok := this.transConns.Load(batchData.ip)
			if ok {
				conn, ok := connect.(*Connection)
				if !ok {
					return
				}
				err = conn.WriteJSON(newMap)
				if err != nil {
					logx.Info(err)
				}
			} else {
				logx.Info(fmt.Printf("trans:%#v不在线", batchData.ip))
				return
			}
		}
	})
	// Concurrent sends to clients
	mapreduce.MapVoid(func(source chan<- interface{}) {
		for k1 := range localClientIds {
			this.clientIdSessions.RangeNextMap(clientIds[k1], func(key1 string, key2 string, value interface{}) bool {
				source <- value
				return true
			})
		}
	}, func(item interface{}) {
		se := item.(*Session)
		// newMap := make(map[string]interface{})
		// for k, v := range req {
		// 	newMap[k] = v
		// }
		// newMap["receive_client_id"] = se.ClientId
		err := se.Conn.WriteJSON(req)
		if err != nil {
			logx.Info(err)
		}
	})
	return nil
}

type NodeMap struct {
	node      *unsafehash.Node
	clientIds []string
}

type BatchData struct {
	ip        string
	clientIds []string
}

// ClientIdsOnline Get online users in the group
func (this *Node) ClientIdsOnline(clientIds []string) []string {
	// now := time.Now().Unix()
	onlineClientIds := make([]string, 0)
	nodes := make(map[string]*NodeMap)
	for k1 := range clientIds {
		redisNode := this.userHashRing.Get(clientIds[k1])
		if _, ok := nodes[redisNode.Id]; ok {
			nodes[redisNode.Id].clientIds = append(nodes[redisNode.Id].clientIds, clientIds[k1])
		} else {
			nodes[redisNode.Id] = &NodeMap{
				node:      redisNode,
				clientIds: []string{clientIds[k1]},
			}
		}
	}
	rangeTime := cast.ToString(time.Now().Unix() - this.clientTimeout)
	for k1 := range nodes {
		nodeMap := nodes[k1]
		pipe := nodeMap.node.Extra.(*redis.Client).Pipeline()
		for k2 := range nodeMap.clientIds {
			pipe.ZRangeByScore(context.Background(), clientPrefix+nodeMap.clientIds[k2], &redis.ZRangeBy{Min: rangeTime, Max: "+inf"}).Result()
		}
		cmders, err := pipe.Exec(context.Background())
		if err != nil {
			logx.Info(err)
		}
		for k3, cmder := range cmders {
			cmd := cmder.(*redis.StringSliceCmd)
			err := cmd.Err()
			if err != nil {
				logx.Info(err)
			} else {
				onlineClientIds = append(onlineClientIds, nodeMap.clientIds[k3])
			}
		}
	}
	return onlineClientIds
}

//GetSessionsByClientIds Get online users in the group
func (this *Node) GetSessionsByClientIds(clientIds []string) []*Session {
	sessions := make([]*Session, 0)
	for k1 := range clientIds {
		this.clientIdSessions.RangeNextMap(clientIds[k1], func(key1, key2 string, value interface{}) bool {
			sessions = append(sessions, value.(*Session))
			return true
		})
	}
	return sessions
}

//BindClientId bind clientId with session
func (this *Node) BindClientId(clientId string, se *Session) error {
	sid := se.Conn.Conn.RemoteAddr().String()
	this.clientIdSessions.StoreWithPlugin(clientId, sid, se, func() {
		oldClientId := se.CasClientId(clientId)
		if oldClientId != "" && oldClientId != clientId {
			this.clientIdSessions.DeleteWithoutLock(oldClientId, sid)
		}
	})

	now := time.Now().Unix()
	redisNode := this.userHashRing.Get(clientId)
	fmt.Printf("%#v\n", clientPrefix+clientId)
	fmt.Printf("%#v\n", this.transAddress)
	err := redisNode.Extra.(*redis.Client).ZAdd(context.Background(), clientPrefix+clientId, &redis.Z{Score: cast.ToFloat64(now), Member: this.transAddress}).Err()
	if err != nil {
		return err
	}
	err = redisNode.Extra.(*redis.Client).Expire(context.Background(), clientPrefix+clientId, time.Duration(this.clientTimeout)*time.Second).Err()
	if err != nil {
		return err
	}

	return nil
}

//SendToClientId Send message to a clientId
func (this *Node) SendToClientId(clientId string, req map[string]interface{}) error {
	if req == nil {
		return fmt.Errorf("message is nil")
	}
	now := time.Now().Unix()
	redisNode := this.userHashRing.Get(clientId)
	ipArr, err := redisNode.Extra.(*redis.Client).ZRangeByScore(context.Background(), clientPrefix+clientId, &redis.ZRangeBy{Min: cast.ToString(now - this.clientTimeout), Max: "+inf"}).Result()
	if err == nil {
		newMap := make(map[string]interface{})
		for k, v := range req {
			newMap[k] = v
		}
		newMap["receive_client_id"] = clientId
		mapreduce.MapVoid(func(source chan<- interface{}) {
			for key := range ipArr {
				source <- ipArr[key]
			}
		}, func(item interface{}) {
			ip := item.(string)
			if ip == this.transAddress {
				this.clientIdSessions.RangeNextMap(clientId, func(k1, k2 string, se interface{}) bool {
					err = se.(*Session).Conn.WriteJSON(req)
					if err != nil {
						logx.Info(err)
					}
					return true
				})
			} else {
				if this.nodeConf.Host == "" {
					reqByte, err := json.Marshal(newMap)
					if err != nil {
						logx.Info(err)
					}
					err = this.bizRedis.Publish(context.Background(), ip, string(reqByte)).Err()
					if err != nil {
						logx.Info(err)
					}
				} else {
					connect, ok := this.transConns.Load(ip)
					if ok {
						conn, ok := connect.(*Connection)
						if !ok {
							return
						}
						err = conn.WriteJSON(newMap)
						if err != nil {
							logx.Info(err)
						}
					} else {
						logx.Info(fmt.Printf("trans:%#v不在线", ip))
						return
					}
				}

			}

		})
	}
	return nil
}

//SendToTrans Send message to a clientId
func (this *Node) SendToTrans(clientId string, path string, req interface{}) error {
	mapreduce.MapVoid(func(source chan<- interface{}) {
		this.transConns.Range(func(key, value interface{}) bool {
			if key == this.transAddress {
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
		err := conn.WriteJSON(req)
		if err != nil {
			logx.Info(err)
		}
	})
	return nil
}

// GetHosts get node address
func (this *Node) GetHosts() ([]string, error) {
	list := make([]string, 0)
	now := time.Now().Unix()
	addrMap, err := this.bizRedis.HGetAll(context.Background(), NodeAddress).Result()
	if err != nil {
		return nil, err
	}
	sortList := make([]Sort, 0)
	expire := now - this.nodeTimeout
	for k1 := range addrMap {
		addrObj := make(map[string]interface{})
		err := json.Unmarshal([]byte(addrMap[k1]), &addrObj)
		if err != nil {
			return nil, err
		}
		if cast.ToInt64(addrObj["timestamp"]) > expire {
			sortList = append(sortList, Sort{
				Obj:  k1,
				Sort: cast.ToInt64(addrObj["client_count"]),
			})
		}
	}
	sort.Sort(SortList(sortList))
	for k1 := range sortList {
		list = append(list, sortList[k1].Obj.(string))
	}
	return list, nil
}

// Consumer pull message from other node
func (this *Node) Consumer() {
	go func() {
		pb := this.bizRedis.Subscribe(context.Background(), this.transAddress)
		for mg := range pb.Channel() {
			data := make(map[string]interface{})
			err := json.Unmarshal([]byte(mg.Payload), &data)
			if err != nil {
				logx.Info(err)
				return
			}
			if receiveClientId, ok := data["receive_client_id"].(string); ok {
				delete(data, "receive_client_id")
				this.clientIdSessions.RangeNextMap(receiveClientId, func(k1, k2 string, se interface{}) bool {
					err = se.(*Session).Conn.WriteJSON(data)
					if err != nil {
						logx.Info(err)
					}
					return true
				})
			} else if receiveClientIds, ok := data["receive_client_ids"].([]string); ok {
				delete(data, "receive_client_ids")
				for k1 := range receiveClientIds {
					receiveClientId := receiveClientIds[k1]
					this.clientIdSessions.RangeNextMap(receiveClientId, func(k1, k2 string, se interface{}) bool {
						err = se.(*Session).Conn.WriteJSON(data)
						if err != nil {
							logx.Info(err)
						}
						return true
					})
				}
			}
		}
	}()
}
