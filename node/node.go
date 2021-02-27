package node

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/spf13/cast"
	"github.com/weblazy/core/mapreduce"
	"github.com/weblazy/easy/utils/logx"
	"github.com/weblazy/easy/utils/syncx"
	"github.com/weblazy/easy/utils/timingwheel"
	"github.com/weblazy/goutil"
	"github.com/weblazy/socket-cluster/discovery"
	"github.com/weblazy/socket-cluster/dns"
	"github.com/weblazy/socket-cluster/protocol"
	"github.com/weblazy/socket-cluster/session_storage"
	"github.com/weblazy/socket-cluster/unsafehash"
)

type (

	// Node communication node
	Node struct {
		protocol.Node
		// adminRedis redis client store node info
		adminRedis       *redis.Client
		nodeConf         *NodeConf
		clientIdSessions *syncx.ConcurrentDoubleMap
		// key: socket address
		// value: *session
		clientConns    goutil.Map
		clientAddress  string                         // External communication address
		transAddress   string                         // Internal communication address
		timer          *timingwheel.TimingWheel       // Timingwheel
		startTime      time.Time                      // start time
		sessionStorage session_storage.SessionStorage // userHashRing ring storage userId
		// key: node transAddress
		// value: *session
		transConns    goutil.Map //
		transClients  goutil.Map //
		transServices goutil.Map //
		nodeTimeout   int64      // node heartbeat timeout time
		clientTimeout int64      // client heartbeat timeout time
	}
)

var ()

// NewPeer creates a new peer.
func StartNode(cfg *NodeConf) (*Node, error) {
	rds := redis.NewClient(cfg.RedisConf)
	timer, err := timingwheel.NewTimingWheel(time.Second, 30, func(k, v interface{}) {
		logx.Infof("%s auth timeout", k)
		if v.(protocol.Connection) != nil {
			err := v.(protocol.Connection).Close()
			if err != nil {
				logx.Info(err)
			}
		}

	})

	if err != nil {
		return nil, err
	}
	this := &Node{
		nodeConf:         cfg,
		adminRedis:       rds,
		clientIdSessions: syncx.NewConcurrentDoubleMap(32),
		startTime:        time.Now(),
		timer:            timer,
		clientConns:      goutil.AtomicMap(),
		transConns:       goutil.AtomicMap(),
		transServices:    goutil.AtomicMap(),
		nodeTimeout:      cfg.NodePingInterval * 3,
		clientTimeout:    cfg.ClientPingInterval * 3,
	}
	cfg.protocolHandler.ListenAndServe(cfg.Port, this)

	this.SendPing()
	this.Ping()
	this.Register()

	return this, nil
}

func (this *Node) OnConnect(connect protocol.Connection) {
	key := connect.Addr()
	v1, ok := this.clientConns.Load(key)
	if ok {
		clientId := v1.(*Session).ClientId
		if clientId != "" {
			this.clientIdSessions.Delete(clientId, key)
		}
	}
	this.clientConns.Delete(key)
}

func (this *Node) OnClose(connect protocol.Connection) {
	this.timer.SetTimer(connect.Addr(), connect, authTime)
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
			err = this.adminRedis.HSet(context.Background(), NodeAddress, this.transAddress, string(nodeInfoByte)).Err()
			if err != nil {
				logx.Info(err)
			}
			nodeMap, err := this.adminRedis.HGetAll(context.Background(), NodeAddress).Result()
			if err != nil {
				logx.Info(err)
			}
			if this.nodeConf.Host != "" {
				// err = this.UpdateNodeList(nodeMap)
				// if err != nil {
				// 	logx.Info(err)
				// }
				this.transConns.Range(func(k, v interface{}) bool {
					conn, ok := v.(protocol.Connection)
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
						err := this.adminRedis.HDel(context.Background(), NodeAddress, k1).Err()
						if err != nil {
							logx.Info(err)
						}
					}
				}
			}

		}
	}()
}

// SendPing send node Heartbeat
func (this *Node) Ping() {
	go func() {
		for {
			time.Sleep(time.Duration(this.nodeConf.NodePingInterval) * time.Second)
			// nodeInfo := NodeInfo{
			// 	ClientAddress: this.clientAddress,
			// 	ClientCount:   int64(this.clientConns.Len()),
			// 	Timestamp:     time.Now().Unix(),
			// }
			// nodeInfoByte, err := proto.Marshal(nodeInfo)
			// if err != nil {
			// 	logx.Info(err)
			// }
			// discoveryHandler.Ping(nodeInfoByte)
			// nodeMap, err := discoveryHandler.GetServices()

			// if err != nil {
			// 	logx.Info(err)
			// }

			// err = this.UpdateNodeList(nodeMap)
			// if err != nil {
			// 	logx.Info(err)
			// }
			this.transConns.Range(func(k, v interface{}) bool {
				conn, ok := v.(protocol.Connection)
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

// Register
func (this *Node) Register() {
	this.nodeConf.discoveryHandler.Register()
	watchChan := make(chan discovery.EventType, 1)
	go this.nodeConf.discoveryHandler.WatchService(watchChan)
	// nodeMap, err := discoveryHandler.GetService()
	go func() {
		for {
			select {
			case value, ok := <-watchChan:
				fmt.Printf("%#v\n", value)
				fmt.Printf("%#v\n", ok)
				if !ok {
					fmt.Printf("监听失败退出\n")
					return
				}
				err := this.UpdateNodeList()
				if err != nil {
					logx.Info(err)
				}

			}

		}
	}()
	// init
	err := this.UpdateNodeList()
	if err != nil {
		logx.Info(err)
	}
}

//Update clients num
// func (this *Node) UpdateRedis() {
// 	go func() {
// 		for {
// 			time.Sleep(redisInterval * time.Second)
// 			err := this.adminRedis.ZAdd(context.Background(), NodeAddress, &redis.Z{
// 				Score:  float64(this.clientConns.Len()),
// 				Member: this.clientAddress}).Err()
// 			if err != nil {
// 				logx.Info(err) f
// 			}
// 		}
// 	}()
// }

// IsOnline determine if a clientId is online
func (this *Node) IsOnline(clientId int64) bool {
	addrArr, err := this.sessionStorage.GetIps(clientId)
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
func (this *Node) OnClientMsg(conn protocol.Connection, msg []byte) {
	sid := conn.Addr()
	session, ok := this.clientConns.Load(sid)
	clientId := ""
	if ok {
		clientId = session.(*Session).ClientId
	}
	this.nodeConf.onMsg(&Context{Conn: conn, Msg: msg, ClientId: clientId})
}

// OnTransMsg handle internal communication node messages
func (this *Node) OnTransMsg(conn protocol.Connection, msg []byte) {
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
func (this *Node) AuthTrans(conn protocol.Connection, args map[string]interface{}) error {
	sid := args["trans_address"].(string)
	this.timer.RemoveTimer(conn.Addr()) //Cancel timeingwheel task
	if args["password"].(string) != this.nodeConf.Password {
		logx.Infof("Connect:%s,Wrong password:%s", sid, args["password"].(string))
		conn.Close()
		return fmt.Errorf("auth faild")
	}
	this.transConns.Store(sid, conn)
	return nil
}

// AuthClient Auth the node
func (this *Node) AuthClient(conn protocol.Connection, clientId string) error {
	sid := conn.Addr()
	this.timer.RemoveTimer(sid) //Cancel timeingwheel task
	session := &Session{Conn: conn, ClientId: clientId}
	this.clientConns.Store(sid, session)
	return this.BindClientId(cast.ToInt64(clientId), session)
}

// UpdateNodeList Add handles addition request
func (this *Node) UpdateNodeList() error {
	nodeMap, err := dns.DnsParse(this.nodeConf.Host)
	if err != nil {
		logx.Info(err)
		return err
	}
	// now := time.Now().Unix()
	for key := range nodeMap {
		ipAddress := nodeMap[key]
		if ipAddress == this.transAddress {
			continue
		}
		// nodeInfo := make(map[string]interface{})
		// err := json.Unmarshal([]byte(nodeMap[key]), &nodeInfo)
		// if err != nil {
		// 	logx.Info(err)
		// }
		// if cast.ToInt64(nodeInfo["timestamp"])+this.nodeTimeout < now {
		// 	continue
		// }

		// 数字小的连接数字大的
		// if strings.Compare(this.transAddress, transAddress) >= 0 {
		// 	continue
		// }
		//已经建立连接
		_, ok := this.transServices.LoadOrStore(ipAddress, "")
		if ok {
			continue
		}

		session, err := this.nodeConf.protocolHandler.Dial(ipAddress)
		if err != nil {
			logx.Info("dial:", err)
			this.transServices.Delete(ipAddress)
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
		conn := session.Conn
		err = conn.WriteJSON(data)
		if err != nil {
			logx.Info(err)
		}
		this.transServices.Store(ipAddress, conn)
		go func(ipAddress string, conn protocol.Connection) {
			defer func(ipAddress string, conn protocol.Connection) {
				this.transServices.Delete(ipAddress)
				conn.Close()
			}(ipAddress, conn)
			for {
				msg, err := conn.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						logx.Info(err)
					}
					break
				}
				logx.Info(string(msg))
				this.OnTransMsg(conn, msg)
			}
		}(ipAddress, conn)
	}
	return nil
}

// SendToClientIds Sending messages to multiple clients
func (this *Node) SendToClientIds(clientIds []string, req map[string]interface{}) error {
	if req == nil {
		return fmt.Errorf("message is nil")
	}
	localClientIds, clientMap, err := this.sessionStorage.GetClientsIps(clientIds)
	if err != nil {
		return err
	}
	// Concurrent sends to other nodes
	mapreduce.MapVoid(func(source chan<- interface{}) {
		for k1 := range clientMap {
			source <- &BatchData{ip: k1, clientIds: clientMap[k1]}
		}
	}, func(item interface{}) {
		batchData := item.(*BatchData)
		newMap := make(map[string]interface{})
		for k, v := range req {
			newMap[k] = v
		}
		newMap["receive_client_ids"] = batchData.clientIds
		connect, ok := this.transConns.Load(batchData.ip)
		if ok {
			conn, ok := connect.(protocol.Connection)
			if !ok {
				return
			}
			err := conn.WriteJSON(newMap)
			if err != nil {
				logx.Info(err)
			}
		} else {
			logx.Info(fmt.Printf("trans:%#v不在线", batchData.ip))
			return
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
func (this *Node) BindClientId(clientId int64, se *Session) error {
	sid := se.Conn.Addr()
	this.clientIdSessions.StoreWithPlugin(cast.ToString(clientId), sid, se, func() {
		oldClientId := se.CasClientId(cast.ToString(clientId))
		oldClientIdInt := cast.ToInt64(oldClientId)
		if oldClientId != "" && oldClientIdInt != clientId {
			this.clientIdSessions.DeleteWithoutLock(oldClientId, sid)
		}
	})

	return this.sessionStorage.BindClientId(clientId)

}

// SendToClientId Send message to a clientId
func (this *Node) SendToClientId(clientId string, req map[string]interface{}) error {
	if req == nil {
		return fmt.Errorf("message is nil")
	}
	ipArr, err := this.sessionStorage.GetIps(cast.ToInt64(clientId))
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
				connect, ok := this.transConns.Load(ip)
				if ok {
					conn, ok := connect.(protocol.Connection)
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
		conn, ok := item.(protocol.Connection)
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
	addrMap, err := this.adminRedis.HGetAll(context.Background(), NodeAddress).Result()
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
