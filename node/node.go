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
	Node interface {
		AuthClient(conn protocol.Connection, clientId string) error
		OnClientPing(clientId int64) error
		IsOnline(clientId int64) bool
		SendToClientIds(clientIds []string, req []byte) error
		SendToClientId(clientId string, req []byte) error
	}

	// Node communication node
	node struct {
		Node
		nodeConf         *NodeConf
		clientIdSessions *syncx.ConcurrentDoubleMap
		// key: socket address
		// value: *session
		clientConns goutil.Map
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
func NewNode(cfg *NodeConf) (Node, error) {
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
	nodeObj := &node{
		nodeConf:         cfg,
		clientIdSessions: syncx.NewConcurrentDoubleMap(32),
		startTime:        time.Now(),
		timer:            timer,
		clientConns:      goutil.AtomicMap(),
		transClients:     goutil.AtomicMap(),
		transServices:    goutil.AtomicMap(),
		nodeTimeout:      cfg.nodePingInterval * 3,
		clientTimeout:    cfg.clientPingInterval * 3,
	}
	nodeObj.nodeConf.discoveryHandler.SetNodeId(cfg.nodeId)
	nodeObj.nodeConf.sessionStorageHandler.SetNodeId(cfg.nodeId)
	cfg.internalProtocolHandler.ListenAndServe(cfg.internalPort, nodeObj.onTransConnect)
	cfg.protocolHandler.ListenAndServe(cfg.port, nodeObj.onClientConnect)
	nodeObj.sendPing()
	nodeObj.register()
	return nodeObj, nil
}

// AuthClient Auth the node
func (this *node) AuthClient(conn protocol.Connection, clientId string) error {
	addr := conn.Addr()
	this.timer.RemoveTimer(addr) // Cancel timeingwheel task
	session := &Session{Conn: conn, ClientId: clientId}
	this.clientConns.Store(addr, session)
	return this.bindClientId(cast.ToInt64(clientId), session)
}

// BindClientId bind clientId with session
func (this *node) bindClientId(clientId int64, se *Session) error {
	addr := se.Conn.Addr()
	this.clientIdSessions.StoreWithPlugin(cast.ToString(clientId), addr, se, func() {
		oldClientId := se.CasClientId(cast.ToString(clientId))
		oldClientIdInt := cast.ToInt64(oldClientId)
		if oldClientId != "" && oldClientIdInt != clientId {
			this.clientIdSessions.DeleteWithoutLock(oldClientId, addr)
		}
	})
	return this.nodeConf.sessionStorageHandler.BindClientId(clientId)
}

func (this *node) OnClientPing(clientId int64) error {
	return this.nodeConf.sessionStorageHandler.OnClientPing(clientId)
}

// IsOnline determine if a clientId is online
func (this *node) IsOnline(clientId int64) bool {
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

// SendToClientId Send message to a clientId
func (this *node) SendToClientId(clientId string, req []byte) error {
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
			if ip == this.nodeConf.nodeId {
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

type BatchData struct {
	ip        string
	clientIds []string
}

// SendToClientIds Sending messages to multiple clients
func (this *node) SendToClientIds(clientIds []string, req []byte) error {
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
		connect, ok := this.transClients.Load(batchData.ip)
		if ok {
			conn, ok := connect.(protocol.Connection)
			if !ok {
				return
			}
			ids := make([]int64, 0)
			for k1 := range batchData.clientIds {
				ids = append(ids, cast.ToInt64(batchData.clientIds[k1]))
			}
			clientsMsg := ClientsMsg{
				ReceiveClientIds: ids,
				Data:             req,
			}
			clientsMsgBytes, err := proto.Marshal(&clientsMsg)
			if err != nil {
				logx.Info(err)
			}
			msg := Msg{
				MsgType: ClientMsgType,
				Data:    clientsMsgBytes,
			}
			msgBytes, err := proto.Marshal(&msg)
			if err != nil {
				logx.Info(err)
			}
			err = conn.WriteMsg(msgBytes)
			if err != nil {
				logx.Info(err)
			}
		} else {
			logx.Info(fmt.Sprintf("trans:%s不在线", batchData.ip))
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

func (this *node) onClientConnect(connect protocol.Connection) {
	defer func() {
		this.onClientClose(connect)
		if connect != nil {
			connect.Close()
		}
	}()
	this.nodeConf.plugin.OnConnect(connect)
	this.timer.SetTimer(connect.Addr(), connect, authTime)
	for {
		msg, err := connect.ReadMsg()
		if err != nil {
			break
		}
		this.onClientMsg(connect, msg)
	}
}

// OnClientMsg deal client message
func (this *node) onClientMsg(conn protocol.Connection, msg []byte) {
	addr := conn.Addr()
	session, ok := this.clientConns.Load(addr)
	clientId := ""
	if ok {
		clientId = session.(*Session).ClientId
	}
	this.nodeConf.onMsg(&Context{Conn: conn, Msg: msg, ClientId: clientId})
}

func (this *node) onClientClose(connect protocol.Connection) {
	this.nodeConf.plugin.OnClose(connect)
	addr := connect.Addr()
	v1, ok := this.clientConns.Load(addr)
	if ok {
		clientId := v1.(*Session).ClientId
		if clientId != "" {
			this.clientIdSessions.Delete(clientId, addr)
		}
	}
	this.clientConns.Delete(addr)
}

func (this *node) onTransConnect(connect protocol.Connection) {
	defer func() {
		if connect != nil {
			connect.Close()
		}
	}()
	for {
		msg, err := connect.ReadMsg()
		if err != nil {
			break
		}
		this.onTransMsg(connect, msg)
	}
}

// OnTransMsg handle internal communication node messages
func (this *node) onTransMsg(conn protocol.Connection, msg []byte) {
	var transMsg Msg
	err := proto.Unmarshal(msg, &transMsg)
	if err != nil {
		logx.Info(err)
		return
	}
	switch transMsg.MsgType {
	// Authentication message
	case AuthMsgType:
		var authMsg AuthMsg
		err = proto.Unmarshal(transMsg.Data, &authMsg)
		if err != nil {
			logx.Info(err)
		}
		err = this.authTrans(conn, &authMsg)
		if err != nil {
			logx.Info(err)
		}
	// Message forwarded to the client
	case ClientMsgType:
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
	// The heartbeat message
	default:
		logx.Info(transMsg)
	}

}

// AuthTrans Auth the node
func (this *node) authTrans(conn protocol.Connection, authMsg *AuthMsg) error {
	nodeId := authMsg.NodeId
	this.timer.RemoveTimer(conn.Addr()) // Cancel timeingwheel task
	if authMsg.Password != this.nodeConf.password {
		logx.Infof("Connect:%s,Wrong password:%s", nodeId, authMsg.Password)
		conn.Close()
		return fmt.Errorf("auth faild")
	}
	this.transClients.Store(nodeId, conn)
	return nil
}

// SendPing send node Heartbeat
func (this *node) sendPing() {
	go func() {
		for {
			time.Sleep(time.Duration(this.nodeConf.nodePingInterval) * time.Second)
			// update node info
			transClients := make([]interface{}, 0)
			this.transClients.Range(func(k1, v1 interface{}) bool {
				transClients = append(transClients, k1)
				return true
			})
			transServices := make([]interface{}, 0)
			this.transServices.Range(func(k1, v1 interface{}) bool {
				transServices = append(transServices, k1)
				return true
			})

			nodeInfo := map[string]interface{}{
				"node_id":        this.nodeConf.nodeId,
				"trans_clients":  transClients,
				"trans_services": transServices,
				"client_count":   this.clientConns.Len(),
				"timestamp":      time.Now().Unix(),
			}
			nodeInfoByte, err := json.Marshal(nodeInfo)
			if err != nil {
				logx.Info(err)
			}
			err = this.nodeConf.discoveryHandler.UpdateInfo(nodeInfoByte)
			if err != nil {
				logx.Info(err)
			}
			err = this.updateNodeList()
			if err != nil {
				logx.Info(err)
			}
			// send to other node
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
func (this *node) register() {
	this.nodeConf.discoveryHandler.Register()
	watchChan := make(chan discovery.EventType, 1)
	go this.nodeConf.discoveryHandler.WatchService(watchChan)
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
				err := this.updateNodeList()
				if err != nil {
					logx.Info(err)
				}
			}
		}
	}()
	// init
	err := this.updateNodeList()
	if err != nil {
		logx.Info(err)
	}
}

// UpdateNodeList Add handles addition request
func (this *node) updateNodeList() error {
	nodeMap := make(map[string]int)
	for k1 := range this.nodeConf.hostList {
		nodeList, port, err := dns.DnsParse(this.nodeConf.hostList[k1])
		if err != nil {
			logx.Info(err)
			return err
		}
		for k2 := range nodeList {
			nodeMap[nodeList[k2]+":"+port] = 1
		}
	}

	// now := time.Now().Unix()
	for k1 := range nodeMap {
		addr := k1
		if addr == this.nodeConf.nodeId {
			continue
		}

		// 数字小的连接数字大的
		// if strings.Compare(this.transAddress, transAddress) >= 0 {
		// 	continue
		// }
		//已经建立连接
		_, ok := this.transServices.LoadOrStore(addr, "")
		if ok {
			continue
		}
		conn, err := this.nodeConf.internalProtocolHandler.Dial(addr)
		if err != nil {
			logx.Info("dial:", err)
			this.transServices.Delete(addr)
			continue
		}

		auth := AuthMsg{
			Password: this.nodeConf.password,
			NodeId:   this.nodeConf.nodeId,
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
		this.transServices.Store(addr, conn)
		go func(ipAddress string, conn protocol.Connection) {
			defer func(addr string, conn protocol.Connection) {
				this.transServices.Delete(addr)
				if conn != nil {
					conn.Close()
				}
			}(addr, conn)
			for {
				msg, err := conn.ReadMsg()
				if err != nil {
					break
				}
				this.onTransMsg(conn, msg)
			}
		}(addr, conn)
	}
	return nil
}
