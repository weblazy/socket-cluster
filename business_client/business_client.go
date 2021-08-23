package business_client

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/weblazy/core/mapreduce"
	"github.com/weblazy/goutil"
	"github.com/weblazy/socket-cluster/discovery"
	"github.com/weblazy/socket-cluster/dns"
	"github.com/weblazy/socket-cluster/logx"
	"github.com/weblazy/socket-cluster/node"
	"github.com/weblazy/socket-cluster/protocol"
)

type (
	BusinessClient interface {
		// IsOnline gets the online status of the clientId
		IsOnline(clientId string) bool
		// SendToClientId sends a message to a clientId
		SendToClientId(clientId string, req []byte) error
		//  SendToClientId sends a message to multiple clientIds
		SendToClientIds(clientIds []string, req []byte) error
	}

	// Node communication node
	businessClient struct {
		BusinessClient
		businessClientConf *BusinessClientConf
		// Key: socket address
		// Value: protocol.Connection
		nodeIpMap     goutil.Map // Receive messages forwarded by other nodes
		nodeIdMap     goutil.Map // Receive messages forwarded by other nodes
		nodeTimeout   int64      // Node heartbeat timeout time
		clientTimeout int64      // Client heartbeat timeout time

		internalProtocolHandler protocol.Protocol // Direct protocol between node and node
	}
)

// NewNode return a Node
func NewBusinessClient(cfg *BusinessClientConf) (BusinessClient, error) {
	businessClientObj := &businessClient{
		businessClientConf: cfg,
		nodeIpMap:          goutil.AtomicMap(),
		nodeIdMap:          goutil.AtomicMap(),
		nodeTimeout:        cfg.nodePingInterval * 3,
	}
	go businessClientObj.watchService()
	return businessClientObj, nil
}

// SendToClientId Send message to a clientId
func (this *businessClient) SendToClientId(clientId string, req []byte) error {
	if req == nil {
		return fmt.Errorf("message is nil")
	}
	ipArr, err := this.businessClientConf.sessionStorageHandler.GetIps(clientId)
	if err == nil {
		mapreduce.MapVoid(func(source chan<- interface{}) {
			for key := range ipArr {
				source <- ipArr[key]
			}
		}, func(item interface{}) {
			ip := item.(string)

			connect, ok := this.nodeIdMap.Load(ip)
			if ok {
				conn, ok := connect.(protocol.Connection)
				if !ok {
					return
				}
				clientsMsg := node.ClientsMsg{
					ReceiveClientIds: []string{clientId},
					Data:             req,
				}
				clientsMsgBytes, err := proto.Marshal(&clientsMsg)
				transReq := node.Msg{
					MsgType: node.ClientMsgType,
					Data:    clientsMsgBytes,
				}
				reqBytes, err := proto.Marshal(&transReq)
				err = conn.WriteMsg(reqBytes)
				if err != nil {
					logx.LogHandler.Error(err)
				}
			} else {
				logx.LogHandler.Errorf("node:%s not online", ip)
				return
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
func (this *businessClient) SendToClientIds(clientIds []string, req []byte) error {
	if req == nil {
		return fmt.Errorf("message is nil")
	}
	clientMap, err := this.businessClientConf.sessionStorageHandler.GetClientsIps(clientIds)
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
		connect, ok := this.nodeIdMap.Load(batchData.ip)
		if ok {
			conn, ok := connect.(protocol.Connection)
			if !ok {
				return
			}
			ids := make([]string, 0)
			for k1 := range batchData.clientIds {
				ids = append(ids, batchData.clientIds[k1])
			}
			clientsMsg := node.ClientsMsg{
				ReceiveClientIds: ids,
				Data:             req,
			}
			clientsMsgBytes, err := proto.Marshal(&clientsMsg)
			if err != nil {
				logx.LogHandler.Error(err)
			}
			msg := node.Msg{
				MsgType: node.ClientMsgType,
				Data:    clientsMsgBytes,
			}
			msgBytes, err := proto.Marshal(&msg)
			if err != nil {
				logx.LogHandler.Error(err)
			}
			err = conn.WriteMsg(msgBytes)
			if err != nil {
				logx.LogHandler.Error(err)
			}
		} else {
			logx.LogHandler.Error("node:%s not online", batchData.ip)
			return
		}

	})

	return nil
}

// IsOnline determine if a clientId is online
func (this *businessClient) IsOnline(clientId string) bool {
	addrArr, err := this.businessClientConf.sessionStorageHandler.GetIps(clientId)
	if err != nil {
		logx.LogHandler.Error(err)
		return false
	}
	if len(addrArr) > 0 {
		return true
	}
	return false
}

// watchService
func (this *businessClient) watchService() {
	watchChan := make(chan discovery.EventType, 1)
	go this.businessClientConf.discoveryHandler.WatchService(watchChan)
	go func() {
		for {
			select {
			case _, ok := <-watchChan:
				if !ok {
					logx.LogHandler.Infof("channel close\n")
					return
				}
				err := this.updateNodeList()
				if err != nil {
					logx.LogHandler.Error(err)
				}
			}
		}
	}()
	// init
	err := this.updateNodeList()
	if err != nil {
		logx.LogHandler.Error(err)
	}
}

// UpdateNodeList Add handles addition request
func (this *businessClient) updateNodeList() error {
	nodeMap := make(map[string]int)
	for k1 := range this.businessClientConf.hostList {
		nodeList, port, err := dns.DnsParse(this.businessClientConf.hostList[k1])
		if err != nil {
			logx.LogHandler.Error(err)
			return err
		}
		for k2 := range nodeList {
			nodeMap[nodeList[k2]+":"+port] = 1
		}
	}

	for k1 := range nodeMap {
		addr := k1
		// Connection already exists
		_, ok := this.nodeIpMap.LoadOrStore(addr, "")
		if ok {
			continue
		}
		conn, err := this.businessClientConf.internalProtocolHandler.Dial(addr)
		if err != nil {
			logx.LogHandler.Errorf("dial:%s", err.Error())
			this.nodeIpMap.Delete(addr)
			continue
		}

		auth := node.AuthMsg{
			Password: this.businessClientConf.password,
		}
		authBytes, err := proto.Marshal(&auth)
		if err != nil {
			logx.LogHandler.Error(err)
		}
		data := node.Msg{
			MsgType: node.AuthBusinessClientMsgType,
			Data:    authBytes,
		}
		reqBytes, err := proto.Marshal(&data)
		if err != nil {
			logx.LogHandler.Error(err)
		}
		err = conn.WriteMsg(reqBytes)
		if err != nil {
			logx.LogHandler.Info(err)
		}
		this.nodeIpMap.Store(addr, conn)
		go func(addr string, conn protocol.Connection) {
			defer func(addr string, conn protocol.Connection) {
				nodeId, _ := this.nodeIpMap.Load(addr)
				if nodeId != "" {
					this.nodeIdMap.Delete(nodeId)
				}
				this.nodeIpMap.Delete(addr)
				if conn != nil {
					conn.Close()
				}
			}(addr, conn)
			for {
				msg, err := conn.ReadMsg()
				if err != nil {
					break
				}
				this.onTransServerMsg(addr, conn, msg)
			}
		}(addr, conn)
	}
	return nil
}

// OnTransMsg handle internal communication node messages
func (this *businessClient) onTransServerMsg(addr string, conn protocol.Connection, msg []byte) {
	var transMsg node.Msg
	err := proto.Unmarshal(msg, &transMsg)
	if err != nil {
		logx.LogHandler.Error(err)
		return
	}
	switch transMsg.MsgType {

	case node.PingMsgType: // The heartbeat message
	case node.BindNodeIdMsgType:
		var bindNodeIdMsg node.BindNodeIdMsg
		err = proto.Unmarshal(transMsg.Data, &bindNodeIdMsg)
		if err != nil {
			logx.LogHandler.Error(err)
		}
		this.nodeIdMap.Store(bindNodeIdMsg.NodeId, conn)
		this.nodeIpMap.Store(addr, bindNodeIdMsg.NodeId)
	default: // The unknow message
		logx.LogHandler.Info(transMsg)
	}

}

// SendPing send node Heartbeat
func (this *businessClient) sendPing() {
	go func() {
		for {
			time.Sleep(time.Duration(this.businessClientConf.nodePingInterval) * time.Second)

			// send to other node
			this.nodeIpMap.Range(func(k, v interface{}) bool {
				conn, ok := v.(protocol.Connection)
				if !ok {
					return true
				}

				err := conn.WriteMsg(node.PingMsg)
				if err != nil {
					logx.LogHandler.Error(err)
				}
				return true
			})
		}
	}()
}
