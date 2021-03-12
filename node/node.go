package node

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/spf13/cast"
	"github.com/weblazy/core/mapreduce"
	"github.com/weblazy/easy/utils/logx"
	"github.com/weblazy/easy/utils/syncx"
	"github.com/weblazy/easy/utils/timingwheel"
	"github.com/weblazy/goutil"
	"github.com/weblazy/socket-cluster/discovery"
	"github.com/weblazy/socket-cluster/dns"
	"github.com/weblazy/socket-cluster/protocol"
)

type (
	// Node communication node
	Node struct {
		protocol.Node
		nodeConf         *NodeConf
		clientIdSessions *syncx.ConcurrentDoubleMap
		// key: socket address
		// value: *session
		clientConns goutil.Map
		nodeId      string                   // Node unique identification
		timer       *timingwheel.TimingWheel // Timingwheel Close connects that timeout without authentication
		startTime   time.Time                // Start time
		// key: nodeId
		// value: *session
		transClients goutil.Map // Forward the message to another node
		// key: nodeId
		// value: *session
		transServices goutil.Map // Receive messages forwarded by other nodes
		nodeTimeout   int64      // node heartbeat timeout time
		clientTimeout int64      // client heartbeat timeout time
	}
)

// NewNode
func NewNode(cfg *NodeConf) (*Node, error) {
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
	node := &Node{
		nodeConf:         cfg,
		nodeId:           cfg.Key,
		clientIdSessions: syncx.NewConcurrentDoubleMap(32),
		startTime:        time.Now(),
		timer:            timer,
		clientConns:      goutil.AtomicMap(),
		transClients:     goutil.AtomicMap(),
		transServices:    goutil.AtomicMap(),
		nodeTimeout:      cfg.NodePingInterval * 3,
		clientTimeout:    cfg.ClientPingInterval * 3,
	}
	node.nodeConf.discoveryHandler.SetNodeId(cfg.Key)
	node.nodeConf.sessionStorageHandler.SetNodeId(cfg.Key)
	cfg.internalProtocolHandler.SetNodeHandler(node)
	cfg.internalProtocolHandler.ListenAndServe(cfg.InternalPort)
	cfg.protocolHandler.SetNodeHandler(node)
	cfg.protocolHandler.ListenAndServe(cfg.Port)
	node.SendPing()
	node.Register()
	return node, nil
}

func (this *Node) OnConnect(connect protocol.Connection) {
	this.timer.SetTimer(connect.Addr(), connect, authTime)
}

func (this *Node) OnClose(connect protocol.Connection) {
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

// SendPing send node Heartbeat
func (this *Node) SendPing() {
	go func() {
		for {
			time.Sleep(time.Duration(this.nodeConf.NodePingInterval) * time.Second)
			nodeInfo := map[string]interface{}{
				"node_id":      this.nodeId,
				"client_count": this.clientConns.Len(),
				"timestamp":    time.Now().Unix(),
			}
			nodeInfoByte, err := json.Marshal(nodeInfo)
			if err != nil {
				logx.Info(err)
			}
			err = this.nodeConf.discoveryHandler.UpdateInfo(nodeInfoByte)
			if err != nil {
				logx.Info(err)
			}
			err = this.UpdateNodeList()
			if err != nil {
				logx.Info(err)
			}
			this.transServices.Range(func(k, v interface{}) bool {
				conn, ok := v.(protocol.Connection)
				if !ok {
					return true
				}
				err := conn.WriteMsg([]byte{})
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
					fmt.Printf("channel关闭退出\n")
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
	addrArr, err := this.nodeConf.sessionStorageHandler.GetIps(clientId)
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
	// data := make(map[string]interface{})
	// err := json.Unmarshal(msg, &data)
	var transMsg Msg
	err := proto.Unmarshal(msg, &transMsg)
	if err != nil {
		logx.Info(err)
		return
	}
	switch transMsg.MsgType {
	case AuthMsgType:
		var authMsg AuthMsg
		err = proto.Unmarshal(transMsg.Data, &authMsg)
		if err != nil {
			logx.Info(err)
		}
		err = this.AuthTrans(conn, &authMsg)
		if err != nil {
			logx.Info(err)
		}
	default:
		var clientsMsg ClientsMsg
		err = proto.Unmarshal(transMsg.Data, &clientsMsg)
		if err != nil {
			logx.Info(err)
		}
		for k1 := range clientsMsg.ReceiveClientIds {
			receiveClientId := clientsMsg.ReceiveClientIds[k1]
			this.clientIdSessions.RangeNextMap(cast.ToString(receiveClientId), func(k1, k2 string, se interface{}) bool {
				err = se.(*Session).Conn.WriteMsg(clientsMsg.Data)
				if err != nil {
					logx.Info(err)
				}
				return true
			})
		}
	}

}

// AuthTrans Auth the node
func (this *Node) AuthTrans(conn protocol.Connection, authMsg *AuthMsg) error {
	nodeId := authMsg.NodeId
	this.timer.RemoveTimer(conn.Addr()) //Cancel timeingwheel task
	if authMsg.Password != this.nodeConf.Password {
		logx.Infof("Connect:%s,Wrong password:%s", nodeId, authMsg.Password)
		conn.Close()
		return fmt.Errorf("auth faild")
	}
	this.transClients.Store(nodeId, conn)
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
	nodeMap := make(map[string]int)
	for k1 := range this.nodeConf.HostList {
		nodeList, err := dns.DnsParse(this.nodeConf.HostList[k1])
		if err != nil {
			logx.Info(err)
			return err
		}
		for k2 := range nodeList {
			nodeMap[nodeList[k2]] = 1
		}
	}

	// now := time.Now().Unix()
	for k1 := range nodeMap {
		ipAddress := k1
		if ipAddress == this.nodeId {
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
		conn, err := this.nodeConf.internalProtocolHandler.Dial(fmt.Sprintf("%s:%d", ipAddress, this.nodeConf.InternalPort))
		if err != nil {
			logx.Info("dial:", err)
			this.transServices.Delete(ipAddress)
			continue
		}

		auth := AuthMsg{
			Password: this.nodeConf.Password,
			NodeId:   this.nodeId,
		}
		authBytes, err := proto.Marshal(&auth)
		if err != nil {
			logx.Info(err)
		}
		data := Msg{
			MsgType: AuthMsgType,
			Data:    authBytes,
		}
		reqBytes, err := proto.Marshal(&data)
		if err != nil {
			logx.Info(err)
		}
		err = conn.WriteMsg(reqBytes)
		if err != nil {
			logx.Info(err)
		}
		this.transServices.Store(ipAddress, conn)
		go func(ipAddress string, conn protocol.Connection) {
			defer func(ipAddress string, conn protocol.Connection) {
				this.transServices.Delete(ipAddress)
				conn.Close()
			}(ipAddress, conn)
			err := this.nodeConf.internalProtocolHandler.ServeConn(conn, this.OnTransMsg)
			if err != nil {
				return
			}

		}(ipAddress, conn)
	}
	return nil
}

// SendToClientIds Sending messages to multiple clients
func (this *Node) SendToClientIds(clientIds []string, req []byte) error {
	if req == nil {
		return fmt.Errorf("message is nil")
	}
	localClientIds, clientMap, err := this.nodeConf.sessionStorageHandler.GetClientsIps(clientIds)
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
		msg := Msg{
			MsgType: ClientMsgType,
			Data:    req,
		}
		msgBytes, err := proto.Marshal(&msg)
		if err != nil {
			logx.Info(err)
		}
		ids := make([]int64, 0)
		for k1 := range batchData.clientIds {
			ids = append(ids, cast.ToInt64(batchData.clientIds[k1]))
		}
		clientsMsg := ClientsMsg{
			ReceiveClientIds: ids,
			Data:             msgBytes,
		}
		connect, ok := this.transClients.Load(batchData.ip)
		if ok {
			conn, ok := connect.(protocol.Connection)
			if !ok {
				return
			}
			clientsMsgBytes, err := proto.Marshal(&clientsMsg)
			if err != nil {
				logx.Info(err)
			}
			err = conn.WriteMsg(clientsMsgBytes)
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
			this.clientIdSessions.RangeNextMap(localClientIds[k1], func(key1 string, key2 string, value interface{}) bool {
				source <- value
				return true
			})
		}
	}, func(item interface{}) {
		se := item.(*Session)
		err := se.Conn.WriteMsg(req)
		if err != nil {
			logx.Info(err)
		}
	})
	return nil
}

type BatchData struct {
	ip        string
	clientIds []string
}

// GetSessionsByClientIds Get online users in the group
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

// BindClientId bind clientId with session
func (this *Node) BindClientId(clientId int64, se *Session) error {
	sid := se.Conn.Addr()
	this.clientIdSessions.StoreWithPlugin(cast.ToString(clientId), sid, se, func() {
		oldClientId := se.CasClientId(cast.ToString(clientId))
		oldClientIdInt := cast.ToInt64(oldClientId)
		if oldClientId != "" && oldClientIdInt != clientId {
			this.clientIdSessions.DeleteWithoutLock(oldClientId, sid)
		}
	})

	return this.nodeConf.sessionStorageHandler.BindClientId(clientId)

}

// SendToClientId Send message to a clientId
func (this *Node) SendToClientId(clientId string, req []byte) error {
	if req == nil {
		return fmt.Errorf("message is nil")
	}
	ipArr, err := this.nodeConf.sessionStorageHandler.GetIps(cast.ToInt64(clientId))
	if err == nil {
		mapreduce.MapVoid(func(source chan<- interface{}) {
			for key := range ipArr {
				source <- ipArr[key]
			}
		}, func(item interface{}) {
			ip := item.(string)
			if ip == this.nodeId {
				this.clientIdSessions.RangeNextMap(clientId, func(k1, k2 string, se interface{}) bool {
					err = se.(*Session).Conn.WriteMsg(req)
					if err != nil {
						logx.Info(err)
					}
					return true
				})
			} else {
				connect, ok := this.transClients.Load(ip)
				if ok {
					conn, ok := connect.(protocol.Connection)
					if !ok {
						return
					}
					clientsMsg := ClientsMsg{
						ReceiveClientIds: []int64{cast.ToInt64(clientId)},
						Data:             req,
					}
					clientsMsgBytes, err := proto.Marshal(&clientsMsg)
					transReq := Msg{
						MsgType: ClientMsgType,
						Data:    clientsMsgBytes,
					}
					reqBytes, err := proto.Marshal(&transReq)
					err = conn.WriteMsg(reqBytes)
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

// SendToTrans Send message to a clientId
func (this *Node) SendToTrans(clientId string, path string, req []byte) error {
	mapreduce.MapVoid(func(source chan<- interface{}) {
		this.transServices.Range(func(key, value interface{}) bool {
			if key == this.nodeId {
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
		err := conn.WriteMsg(req)
		if err != nil {
			logx.Info(err)
		}
	})
	return nil
}
