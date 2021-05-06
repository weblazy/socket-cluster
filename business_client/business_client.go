package business_client

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/spf13/cast"
	"github.com/weblazy/core/mapreduce"
	"github.com/weblazy/core/timingwheel"
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
		IsOnline(clientId int64) bool
		// SendToClientId sends a message to a clientId
		SendToClientId(clientId string, req []byte) error
		//  SendToClientId sends a message to multiple clientIds
		SendToClientIds(clientIds []string, req []byte) error
	}

	// Node communication node
	businessClient struct {
		BusinessClient
		businessClientConf *BusinessClientConf
		timer              *timingwheel.TimingWheel // Timingwheel Close connects that timeout without authentication
		startTime          time.Time                // Start time
		// Key: socket address
		// Value: protocol.Connection
		nodeServices  goutil.Map // Receive messages forwarded by other nodes
		nodeTimeout   int64      // Node heartbeat timeout time
		clientTimeout int64      // Client heartbeat timeout time

		internalProtocolHandler protocol.Protocol // Direct protocol between node and node
	}
)

// NewNode return a Node
func NewBusinessClient(cfg *BusinessClientConf) (BusinessClient, error) {
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
	businessClientObj := &businessClient{
		businessClientConf: cfg,
		startTime:          time.Now(),
		timer:              timer,
		nodeServices:       goutil.AtomicMap(),
		nodeTimeout:        cfg.nodePingInterval * 3,
	}

	return businessClientObj, nil
}

// SendToClientId Send message to a clientId
func (this *businessClient) SendToClientId(clientId string, req []byte) error {
	if req == nil {
		return fmt.Errorf("message is nil")
	}
	ipArr, err := this.businessClientConf.sessionStorageHandler.GetIps(cast.ToInt64(clientId))
	if err == nil {
		mapreduce.MapVoid(func(source chan<- interface{}) {
			for key := range ipArr {
				source <- ipArr[key]
			}
		}, func(item interface{}) {
			ip := item.(string)

			connect, ok := this.nodeServices.Load(ip)
			if ok {
				conn, ok := connect.(protocol.Connection)
				if !ok {
					return
				}
				clientsMsg := node.ClientsMsg{
					ReceiveClientIds: []int64{cast.ToInt64(clientId)},
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
		_, ok := this.nodeServices.LoadOrStore(addr, "")
		if ok {
			continue
		}
		conn, err := this.businessClientConf.internalProtocolHandler.Dial(addr)
		if err != nil {
			logx.LogHandler.Errorf("dial:%s", err.Error())
			this.nodeServices.Delete(addr)
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
		this.nodeServices.Store(addr, conn)
		go func(ipAddress string, conn protocol.Connection) {
			defer func(addr string, conn protocol.Connection) {
				this.nodeServices.Delete(addr)
				if conn != nil {
					conn.Close()
				}
			}(addr, conn)
			for {
				msg, err := conn.ReadMsg()
				if err != nil {
					break
				}
				this.onTransServerMsg(conn, msg)
			}
		}(addr, conn)
	}
	return nil
}

// Register
func (this *businessClient) register() {
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

// OnTransMsg handle internal communication node messages
func (this *businessClient) onTransServerMsg(conn protocol.Connection, msg []byte) {
	var transMsg node.Msg
	err := proto.Unmarshal(msg, &transMsg)
	if err != nil {
		logx.LogHandler.Error(err)
		return
	}
	switch transMsg.MsgType {

	case node.PingMsgType: // The heartbeat message

	default: // The unknow message
		logx.LogHandler.Info(transMsg)
	}

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
		connect, ok := this.nodeServices.Load(batchData.ip)
		if ok {
			conn, ok := connect.(protocol.Connection)
			if !ok {
				return
			}
			ids := make([]int64, 0)
			for k1 := range batchData.clientIds {
				ids = append(ids, cast.ToInt64(batchData.clientIds[k1]))
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
func (this *businessClient) IsOnline(clientId int64) bool {
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
