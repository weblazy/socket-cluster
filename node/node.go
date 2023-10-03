package node

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/weblazy/core/mapreduce"
	"github.com/weblazy/easy/econfig"
	"github.com/weblazy/easy/elog"
	"github.com/weblazy/easy/syncx"
	"github.com/weblazy/easy/timingwheel"
	"github.com/weblazy/goutil"
	"github.com/weblazy/socket-cluster/logx"
	"github.com/weblazy/socket-cluster/protocol"
	"go.uber.org/zap"
)

type (
	Node interface {
		// SetClientIdOnline binds the client ID to the connection
		SetClientIdOnline(conn protocol.Connection, clientId string) error
		// OnClientPing updates the client heartbeat status
		OnClientPing(clientId string) error
		// IsOnline gets the online status of the clientId
		IsOnline(clientId string) bool
		// SendToClientId sends a message to a clientId
		SendToClientId(clientId string, req []byte) error
		//  SendToClientId sends a message to multiple clientIds
		SendToClientIds(clientIds []string, req []byte) error
	}

	// Node communication node
	node struct {
		Node
		nodeConf *NodeConf
		// clientIdSessions map[key1]map[key2]value
		// Key1: clientId
		// Key2: socket address
		// Value: *Session
		clientIdSessions *syncx.ConcurrentDoubleMap
		// Key: socket address
		// Value: *Session
		clientConns goutil.Map
		timer       *timingwheel.TimingWheel // Timingwheel Close connects that timeout without authentication
		startTime   time.Time                // Start time
		// Key: socket address
		// Value: *Session
		businessClients goutil.Map // Send the message to client
		// Key: nodeId
		// Value: *Session
		// transClients goutil.Map // Forward the message to another node
		// Key: socket address
		// Value: protocol.Connection
		// transServices goutil.Map // Receive messages forwarded by other nodes
		nodeTimeout   int64 // Node heartbeat timeout time
		clientTimeout int64 // Client heartbeat timeout time
	}
)

// NewNode return a Node
func NewNode(cfg *NodeConf) (Node, error) {
	timer, err := timingwheel.NewTimingWheel(time.Second, 30, func(k, v interface{}) {
		// TODO Count the number of unauthenticated connections
		logx.LogHandler.Infof("%s auth timeout", k)
		if v.(protocol.Connection) != nil {
			err := v.(protocol.Connection).Close()
			if err != nil {
				logx.LogHandler.Error(err)
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
		businessClients:  goutil.AtomicMap(),
		nodeTimeout:      cfg.nodePingInterval * 3,   // The timeout is three times as long as the interval between heartbeats
		clientTimeout:    cfg.clientPingInterval * 3, // The timeout is three times as long as the interval between heartbeats
	}
	nodeObj.nodeConf.discoveryHandler.SetNodeAddr(cfg.addr)
	// cfg.internalProtocolHandler.ListenAndServe(cfg.internalPort, nodeObj.onTransConnect)

	cfg.protocolHandler.ListenAndServe(cfg.port, nodeObj.onClientConnect)
	nodeObj.sendPing()
	nodeObj.register()
	return nodeObj, nil
}

// SetOnline sets the clientId to online
func (this *node) SetClientIdOnline(conn protocol.Connection, clientId string) error {
	addr := conn.Addr()
	this.timer.RemoveTimer(addr) // Cancel timeingwheel task
	session := &Session{Conn: conn, ClientId: clientId}
	this.clientConns.Store(addr, session)
	this.clientIdSessions.StoreWithPlugin(clientId, addr, session, func() {
		oldClientId := session.CasClientId(clientId)
		if oldClientId != "" && oldClientId != clientId {
			this.clientIdSessions.DeleteWithoutLock(oldClientId, addr)
		}
	})
	return this.nodeConf.sessionStorageHandler.BindClientId(this.nodeConf.addr, clientId)
}

// OnClientPing updates the client heartbeat status
func (this *node) OnClientPing(clientId string) error {
	return this.nodeConf.sessionStorageHandler.OnClientPing(this.nodeConf.addr, clientId)
}

// IsOnline determine if a clientId is online
func (this *node) IsOnline(clientId string) bool {
	addrArr, err := this.nodeConf.sessionStorageHandler.GetIps(clientId)
	if err != nil {
		logx.LogHandler.Error(err)
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
	ipArr, err := this.nodeConf.sessionStorageHandler.GetIps(clientId)
	if econfig.GlobalViper.GetBool("BaseConfig.Debug") {
		elog.InfoCtx(context.Background(), "SendToClientId", zap.String("clientId", clientId), zap.String("req", string(req)), zap.Any("ipArr", ipArr))
	}
	if err == nil {
		mapreduce.MapVoid(func(source chan<- interface{}) {
			for key := range ipArr {
				source <- ipArr[key]
			}
		}, func(item interface{}) {
			// ip := item.(string)
			// if ip == this.nodeConf.nodeId {
			this.clientIdSessions.RangeNextMap(clientId, func(k1, k2 string, se interface{}) bool {
				err = se.(*Session).Conn.WriteMsg(req)
				if err != nil {
					elog.InfoCtx(context.Background(), "SendToClientId", zap.String("clientId", clientId), zap.String("req", string(req)), zap.Error(err))
				}
				return true
			})
			// } else {
			// 	connect, ok := this.transClients.Load(ip)
			// 	if ok {
			// 		conn, ok := connect.(protocol.Connection)
			// 		if !ok {
			// 			return
			// 		}
			// 		clientsMsg := ClientsMsg{
			// 			ReceiveClientIds: []string{clientId},
			// 			Data:             req,
			// 		}
			// 		clientsMsgBytes, err := proto.Marshal(&clientsMsg)
			// 		if err != nil {
			// 			logx.LogHandler.Error(err)
			// 		}
			// 		transReq := Msg{
			// 			MsgType: ClientMsgType,
			// 			Data:    clientsMsgBytes,
			// 		}
			// 		reqBytes, err := proto.Marshal(&transReq)
			// 		err = conn.WriteMsg(reqBytes)
			// 		if err != nil {
			// 			logx.LogHandler.Error(err)
			// 		}
			// 	} else {
			// 		logx.LogHandler.Errorf("node:%s not online", ip)
			// 		return
			// 	}
			// }
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
	clientMap, err := this.nodeConf.sessionStorageHandler.GetClientsIps(clientIds)
	if err != nil {
		return err
	}
	localClientIds, _ := clientMap[this.nodeConf.addr]
	delete(clientMap, this.nodeConf.addr)
	// Concurrent sends to other nodes
	// mapreduce.MapVoid(func(source chan<- interface{}) {
	// 	for k1 := range clientMap {

	// 		source <- &BatchData{ip: k1, clientIds: clientMap[k1]}
	// 	}
	// }, func(item interface{}) {
	// 	batchData := item.(*BatchData)
	// 	connect, ok := this.transClients.Load(batchData.ip)
	// 	if ok {
	// 		conn, ok := connect.(protocol.Connection)
	// 		if !ok {
	// 			return
	// 		}
	// 		ids := make([]string, 0)
	// 		for k1 := range batchData.clientIds {
	// 			ids = append(ids, batchData.clientIds[k1])
	// 		}
	// 		clientsMsg := ClientsMsg{
	// 			ReceiveClientIds: ids,
	// 			Data:             req,
	// 		}
	// 		clientsMsgBytes, err := proto.Marshal(&clientsMsg)
	// 		if err != nil {
	// 			logx.LogHandler.Error(err)
	// 		}
	// 		msg := Msg{
	// 			MsgType: ClientMsgType,
	// 			Data:    clientsMsgBytes,
	// 		}
	// 		msgBytes, err := proto.Marshal(&msg)
	// 		if err != nil {
	// 			logx.LogHandler.Error(err)
	// 		}
	// 		err = conn.WriteMsg(msgBytes)
	// 		if err != nil {
	// 			logx.LogHandler.Error(err)
	// 		}
	// 	} else {
	// 		logx.LogHandler.Error("node:%s not online", batchData.ip)
	// 		return
	// 	}

	// })
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
			logx.LogHandler.Error(err)
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
	// The timeout unbound clientId connect will be closed
	this.timer.SetTimer(connect.Addr(), connect, authTime)
	for {
		msg, err := connect.ReadMsg()
		if err != nil {
			logx.LogHandler.Error(err)
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

// func (this *node) onTransConnect(connect protocol.Connection) {
// 	defer func() {
// 		if connect != nil {
// 			connect.Close()
// 		}
// 	}()
// 	this.timer.SetTimer(connect.Addr(), connect, authTime)
// 	for {
// 		msg, err := connect.ReadMsg()
// 		if err != nil {
// 			logx.LogHandler.Error(err)
// 			break
// 		}
// 		this.onTransClientMsg(connect, msg)
// 	}
// }

// OnTransMsg handle internal communication node messages
// func (this *node) onTransServerMsg(conn protocol.Connection, msg []byte) {
// 	var transMsg Msg
// 	err := proto.Unmarshal(msg, &transMsg)
// 	if err != nil {
// 		logx.LogHandler.Error(err, string(msg))
// 		return
// 	}
// 	switch transMsg.MsgType {

// 	case ClientMsgType: // Message forwarded to the client
// 		var clientsMsg ClientsMsg
// 		err = proto.Unmarshal(transMsg.Data, &clientsMsg)
// 		if err != nil {
// 			logx.LogHandler.Error(err)
// 		}
// 		for k1 := range clientsMsg.ReceiveClientIds {
// 			receiveClientId := clientsMsg.ReceiveClientIds[k1]
// 			this.clientIdSessions.RangeNextMap(receiveClientId, func(k1, k2 string, se interface{}) bool {
// 				err = se.(*Session).Conn.WriteMsg(clientsMsg.Data)
// 				if err != nil {
// 					logx.LogHandler.Error(err)
// 				}
// 				return true
// 			})
// 		}

// 	case PingMsgType: // The heartbeat message

// 	default: // The unknow message
// 		logx.LogHandler.Info(transMsg)
// 	}

// }

// OnTransMsg handle internal communication node messages
// func (this *node) onTransClientMsg(conn protocol.Connection, msg []byte) {
// 	var transMsg Msg
// 	err := proto.Unmarshal(msg, &transMsg)
// 	if err != nil {
// 		logx.LogHandler.Error(err, string(msg))
// 		return
// 	}
// 	switch transMsg.MsgType {
// 	case AuthNodeMsgType: // Authentication message
// 		var authMsg AuthMsg
// 		err = proto.Unmarshal(transMsg.Data, &authMsg)
// 		if err != nil {
// 			logx.LogHandler.Error(err)
// 		}
// 		err = this.authTrans(conn, &authMsg)
// 		if err != nil {
// 			logx.LogHandler.Error(err)
// 		}

// 	case PingMsgType: // The heartbeat message
// 	case AuthBusinessClientMsgType: // Authentication message
// 		var authMsg AuthMsg
// 		err = proto.Unmarshal(transMsg.Data, &authMsg)
// 		if err != nil {
// 			logx.LogHandler.Error(err)
// 		}
// 		err = this.authBusinessClient(conn, &authMsg)
// 		if err != nil {
// 			logx.LogHandler.Error(err)
// 			return
// 		}
// 		bindNodeIdMsg := BindNodeIdMsg{NodeId: this.nodeConf.nodeId}

// 		bindNodeIdMsgBytes, err := proto.Marshal(&bindNodeIdMsg)
// 		if err != nil {
// 			logx.LogHandler.Error(err)
// 		}
// 		bindNodeIdReq := Msg{
// 			MsgType: BindNodeIdMsgType,
// 			Data:    bindNodeIdMsgBytes,
// 		}
// 		reqBytes, err := proto.Marshal(&bindNodeIdReq)

// 		err = conn.WriteMsg(reqBytes)
// 		if err != nil {
// 			logx.LogHandler.Error(err)
// 		}
// 	case ClientMsgType: // Message forwarded to the client
// 		addr := conn.Addr()
// 		_, ok := this.businessClients.Load(addr)
// 		if !ok {
// 			logx.LogHandler.Error(errors.New("un auth connect"))
// 			return
// 		}
// 		var clientsMsg ClientsMsg
// 		err = proto.Unmarshal(transMsg.Data, &clientsMsg)
// 		if err != nil {
// 			logx.LogHandler.Error(err)
// 		}
// 		for k1 := range clientsMsg.ReceiveClientIds {
// 			receiveClientId := clientsMsg.ReceiveClientIds[k1]
// 			this.clientIdSessions.RangeNextMap(receiveClientId, func(k1, k2 string, se interface{}) bool {
// 				err = se.(*Session).Conn.WriteMsg(clientsMsg.Data)
// 				if err != nil {
// 					logx.LogHandler.Error(err)
// 				}
// 				return true
// 			})
// 		}

// 	default: // The unknow message
// 		logx.LogHandler.Info(transMsg)
// 	}

// }

// AuthTrans Auth the node
// func (this *node) authTrans(conn protocol.Connection, authMsg *AuthMsg) error {
// 	nodeId := authMsg.NodeId
// 	this.timer.RemoveTimer(conn.Addr()) // Cancel timeingwheel task
// 	if authMsg.Password != this.nodeConf.password {
// 		logx.LogHandler.Infof("Connect:%s,Wrong password:%s", nodeId, authMsg.Password)
// 		conn.Close()
// 		return fmt.Errorf("auth faild")
// 	}
// 	this.transClients.Store(nodeId, conn)
// 	return nil
// }

// AuthTrans Auth the node
func (this *node) authBusinessClient(conn protocol.Connection, authMsg *AuthMsg) error {
	addr := conn.Addr()
	this.timer.RemoveTimer(addr) // Cancel timeingwheel task
	if authMsg.Password != this.nodeConf.password {
		logx.LogHandler.Infof("Connect:%s,Wrong password:%s", addr, authMsg.Password)
		conn.Close()
		return fmt.Errorf("auth faild")
	}
	this.businessClients.Store(addr, conn)
	return nil
}

// SendPing send node Heartbeat
func (this *node) sendPing() {
	go func() {
		for {
			time.Sleep(time.Duration(this.nodeConf.nodePingInterval) * time.Second)
			// update node info
			nodeInfo := map[string]interface{}{
				"node_addr":    this.nodeConf.addr,
				"client_count": this.clientConns.Len(),
				"timestamp":    time.Now().Unix(),
			}
			nodeInfoByte, err := json.Marshal(nodeInfo)
			if err != nil {
				logx.LogHandler.Error(err)
			}
			err = this.nodeConf.discoveryHandler.UpdateInfo(nodeInfoByte)
			if err != nil {
				logx.LogHandler.Error(err)
			}
			// err = this.updateNodeList()
			if err != nil {
				logx.LogHandler.Error(err)
			}
			// send to other node
			// this.transServices.Range(func(k, v interface{}) bool {
			// 	conn, ok := v.(protocol.Connection)
			// 	if !ok {
			// 		return true
			// 	}

			// 	err := conn.WriteMsg(PingMsg)
			// 	if err != nil {
			// 		logx.LogHandler.Error(err)
			// 	}
			// 	return true
			// })
		}
	}()
}

// Register
func (this *node) register() {
	this.nodeConf.discoveryHandler.Register()
	// watchChan := make(chan discovery.EventType, 1)
	// go this.nodeConf.discoveryHandler.WatchService(watchChan)
	// go func() {
	// 	for {
	// 		select {
	// 		case _, ok := <-watchChan:
	// 			if !ok {
	// 				logx.LogHandler.Infof("channel close\n")
	// 				return
	// 			}
	// 			err := this.updateNodeList()
	// 			if err != nil {
	// 				logx.LogHandler.Error(err)
	// 			}
	// 		}
	// 	}
	// }()
	// init
	// err := this.updateNodeList()
	// if err != nil {
	// 	logx.LogHandler.Error(err)
	// }
}

// UpdateNodeList Add handles addition request
// func (this *node) updateNodeList() error {
// 	nodeMap := make(map[string]int)
// 	for k1 := range this.nodeConf.hostList {
// 		nodeList, port, err := dns.DnsParse(this.nodeConf.hostList[k1])
// 		if err != nil {
// 			logx.LogHandler.Error(err)
// 			return err
// 		}
// 		for k2 := range nodeList {
// 			nodeMap[nodeList[k2]+":"+port] = 1
// 		}
// 	}

// 	for k1 := range nodeMap {
// 		addr := k1
// 		if addr == this.nodeConf.nodeId {
// 			continue
// 		}

// 		// Connection already exists
// 		_, ok := this.transServices.LoadOrStore(addr, "")
// 		if ok {
// 			continue
// 		}
// 		conn, err := this.nodeConf.internalProtocolHandler.Dial(addr)
// 		if err != nil {
// 			logx.LogHandler.Errorf("dial:%s", err.Error())
// 			this.transServices.Delete(addr)
// 			continue
// 		}

// 		auth := AuthMsg{
// 			Password: this.nodeConf.password,
// 			NodeId:   this.nodeConf.nodeId,
// 		}
// 		authBytes, err := proto.Marshal(&auth)
// 		if err != nil {
// 			logx.LogHandler.Error(err)
// 		}
// 		data := Msg{
// 			MsgType: AuthNodeMsgType,
// 			Data:    authBytes,
// 		}
// 		reqBytes, err := proto.Marshal(&data)
// 		if err != nil {
// 			logx.LogHandler.Error(err)
// 		}
// 		err = conn.WriteMsg(reqBytes)
// 		if err != nil {
// 			logx.LogHandler.Info(err)
// 		}
// 		this.transServices.Store(addr, conn)
// 		go func(addr string, conn protocol.Connection) {
// 			defer func(addr string, conn protocol.Connection) {
// 				this.transServices.Delete(addr)
// 				if conn != nil {
// 					conn.Close()
// 				}
// 			}(addr, conn)
// 			for {
// 				msg, err := conn.ReadMsg()
// 				if err != nil {
// 					break
// 				}
// 				this.onTransServerMsg(conn, msg)
// 			}
// 		}(addr, conn)
// 	}
// 	return nil
// }
