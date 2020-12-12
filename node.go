package websocket_cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cast"
	"github.com/sunmi-OS/gocore/log"
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

	Node struct {
		bizRedis         *redis.Client
		nodeConf         *NodeConf
		clientIdSessions *syncx.ConcurrentDoubleMap
		clientConns      goutil.Map               //External communication value is *session
		clientAddress    string                   //External communication address
		transAddress     string                   //Internal communication address
		timer            *timingwheel.TimingWheel //Timingwheel
		startTime        time.Time
		userHashRing     *unsafehash.Consistent //UsHash ring storage userId
		groupHashRing    *unsafehash.Consistent //UsHash ring storage groupId
		transConns       goutil.Map
		nodeTimeout      int64
		clientTimeout    int64
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
	groupHashRing := unsafehash.NewConsistent(cfg.RedisMaxCount)
	for _, value := range cfg.RedisNodeList {
		rdsObj := redis.NewClient(&redis.Options{
			Addr:     value.RedisConf.Addr,
			Password: value.RedisConf.Password,
			DB:       int(value.RedisConf.DB),
		})
		groupHashRing.Add(unsafehash.NewNode(value.RedisConf.Addr, value.Position, rdsObj))
	}
	this := &Node{
		nodeConf:         cfg,
		bizRedis:         rds,
		clientIdSessions: syncx.NewConcurrentDoubleMap(32),
		startTime:        time.Now(),
		timer:            timer,
		userHashRing:     userHashRing,
		groupHashRing:    groupHashRing,
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

func (this *Node) clientHandler(c echo.Context) error {
	connect, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	defer func() {
		this.clientConns.Delete(connect.RemoteAddr().String())
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
	log.Sugar.Info(connect.RemoteAddr().String(), connect.LocalAddr().String(), "start")
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

//Heartbeat
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

//Determine if a clientId is online
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

func (this *Node) OnClientMsg(conn *Connection, msg []byte) {
	sid := conn.Conn.RemoteAddr().String()
	session, ok := this.clientConns.Load(sid)
	clientId := ""
	if ok {
		clientId = session.(*Session).ClientId
	}
	this.nodeConf.onMsg(&Context{Conn: conn, Msg: msg, ClientId: clientId})
}

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
		receiveClientId := data["receive_client_id"].(string)
		delete(data, "receive_client_id")
		this.clientIdSessions.RangeNextMap(receiveClientId, func(k1, k2 string, se interface{}) bool {
			err = se.(*Session).Conn.WriteJSON(data)
			if err != nil {
				logx.Info(err)
			}
			return true
		})
	}
}

// Auth the node
func (this *Node) AuthTrans(conn *Connection, args map[string]interface{}) error {
	sid := args["trans_address"].(string)
	this.timer.RemoveTimer(conn.Conn.RemoteAddr()) //Cancel timeingwheel task
	if args["password"].(string) != this.nodeConf.Password {
		logx.Infof("Connect:%s,Wrong password:%s", sid, args["password"].(string))
		conn.Conn.Close()
		return fmt.Errorf("auth faild")
	}
	this.transConns.Store(sid, conn)
	return nil
}

// Auth the node
func (this *Node) AuthClient(conn *Connection, clientId string) error {
	sid := conn.Conn.RemoteAddr().String()
	this.timer.RemoveTimer(sid) //Cancel timeingwheel task
	session := &Session{Conn: conn, ClientId: clientId}
	this.clientConns.Store(sid, session)
	return this.BindClientId(clientId, session)
}

// Add handles addition request
func (this *Node) UpdateNodeList(nodeMap map[string]string) error {
	now := time.Now().Unix()
	for key := range nodeMap {
		nodeInfo := make(map[string]interface{})
		err := json.Unmarshal([]byte(nodeMap[key]), &nodeInfo)
		if err != nil {
			logx.Info(err)
		}
		if cast.ToInt64(nodeInfo["timestamp"])+this.nodeTimeout < now {
			continue
		}
		//数字小的连接数字大的
		transAddress := key
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

func (this *Node) SendToClientIds(clientIdList []string, req map[string]interface{}) error {
	if req == nil {
		return fmt.Errorf("message is nil")
	}
	clientIds := this.ClientIdsOnline(clientIdList)
	mapreduce.MapVoid(func(source chan<- interface{}) {
		for k1 := range clientIds {
			this.clientIdSessions.RangeNextMap(clientIds[k1], func(key1 string, key2 string, value interface{}) bool {
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
		newMap["receive_client_id"] = se.ClientId
		err := se.Conn.WriteJSON(newMap)
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

// Get online users in the group
func (this *Node) ClientIdsOnline(clientIds []string) []string {
	// now := time.Now().Unix()
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
	// for k1 := range nodes {
	// 	nodeMap := nodes[k1]
	// 	arr, err := nodeMap.node.Extra.(*redis.Redis).Mget(nodeMap.clientIds...)

	// }

	// for key, value := range arr {
	// 	old, _ := strconv.ParseInt(value, 10, 64)
	// 	if now < old {
	// 		clientIds = append(clientIds, key)
	// 	}
	// }
	return clientIds
}

//Get online users in the group
func (this *Node) GetSessionsByClientIds(clientIds []string) []*Session {
	sessions := make([]*Session, 0)
	for k1 := range clientIds {
		this.clientIdSessions.RangeShard(clientIds[k1], func(key2 string, value interface{}) bool {
			sessions = append(sessions, value.(*Session))
			return true
		})
	}
	return sessions
}

//Get bind clientId with session
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

//Send message to a clientId
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

//Send message to a clientId
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

func (this *Node) Consumer() {
	go func() {
		pb := this.bizRedis.Subscribe(context.Background(), this.transAddress)
		for mg := range pb.Channel() {
			fmt.Printf("收到订阅消息:%#v\n", mg.Payload)
			data := make(map[string]interface{})
			err := json.Unmarshal([]byte(mg.Payload), &data)
			if err != nil {
				logx.Info(err)
				return
			}
			receiveClientId := data["receive_client_id"].(string)
			delete(data, "receive_client_id")
			this.clientIdSessions.RangeNextMap(receiveClientId, func(k1, k2 string, se interface{}) bool {
				err = se.(*Session).Conn.WriteJSON(data)
				if err != nil {
					logx.Info(err)
				}
				return true
			})
		}

	}()
}
