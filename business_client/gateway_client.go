package business_client

import (
	"context"
	"fmt"

	"github.com/weblazy/core/mapreduce"
	"github.com/weblazy/easy/grpc/grpc_client"
	"github.com/weblazy/easy/grpc/grpc_client/grpc_client_config"
	"github.com/weblazy/goutil"
	"github.com/weblazy/socket-cluster/discovery"
	"github.com/weblazy/socket-cluster/grpcs/socket_cluster_gateway/proto/gateway"
	"github.com/weblazy/socket-cluster/logx"
	"github.com/weblazy/socket-cluster/session_storage"
)

type GatewayClient struct {
	nodeIpMap             goutil.Map                     // Receive messages forwarded by other nodes
	nodeIdMap             goutil.Map                     // Receive messages forwarded by other nodes
	sessionStorageHandler session_storage.SessionStorage // On-line state storage components
	businessClientConf    *BusinessClientConf
}

func NewGatewayClient(cfg *BusinessClientConf) *GatewayClient {
	gatewayClient := GatewayClient{
		businessClientConf: cfg,
		nodeIpMap:          goutil.AtomicMap(),
		nodeIdMap:          goutil.AtomicMap(),
	}
	watchChan := make(chan discovery.EventType, 1)
	go gatewayClient.businessClientConf.discoveryHandler.WatchService(watchChan)
	go func() {
		for {
			select {
			case _, ok := <-watchChan:
				if !ok {
					logx.LogHandler.Infof("channel close\n")
					return
				}
				err := gatewayClient.updateNodeList()
				if err != nil {
					logx.LogHandler.Error(err)
				}
			}
		}
	}()
	return &gatewayClient
}

func (g *GatewayClient) updateNodeList() error {
	nodeMap, err := g.businessClientConf.discoveryHandler.GetServerList()
	if err != nil {
		return err
	}
	for k1 := range nodeMap {
		addr := k1
		// Connection already exists
		_, ok := g.nodeIpMap.LoadOrStore(addr, "")
		if ok {
			continue
		}
		cfg := grpc_client_config.DefaultConfig()
		client := grpc_client.NewGrpcClient(cfg)
		gatewayClient := gateway.NewGatewayServiceClient(client)
		g.nodeIpMap.Store(addr, gatewayClient)
	}
	return nil
}

func (g *GatewayClient) SendToClientId(req *gateway.SendToClientIdRequest) (*gateway.SendToClientIdResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("message is nil")
	}
	ipArr, err := g.sessionStorageHandler.GetIps(req.ClientId)
	if err == nil {
		mapreduce.MapVoid(func(source chan<- interface{}) {
			for key := range ipArr {
				source <- ipArr[key]
			}
		}, func(item interface{}) {
			ip := item.(string)
			connect, ok := g.nodeIpMap.Load(ip)
			if ok {
				conn, ok := connect.(gateway.GatewayServiceClient)
				if !ok {
					return
				}
				clientsMsg := &gateway.SendToClientIdRequest{
					ClientId: req.ClientId,
					Data:     req.Data,
				}
				conn.SendToClientId(context.Background(), clientsMsg)
				if err != nil {
					logx.LogHandler.Error(err)
				}
			} else {
				logx.LogHandler.Errorf("node:%s not online", ip)
				return
			}

		})
	}
	return nil, nil
}

type BatchData struct {
	ip        string
	clientIds []string
}

func (g *GatewayClient) SendToClientIds(req *gateway.SendToClientIdsRequest) (*gateway.SendToClientIdsResponse, error) {
	if req == nil {
		return &gateway.SendToClientIdsResponse{
			Code: -1,
			Msg:  "message is nil",
		}, nil
	}
	clientMap, err := g.businessClientConf.sessionStorageHandler.GetClientsIps(req.ClientIds)
	if err != nil {
		return &gateway.SendToClientIdsResponse{
			Code: -1,
			Msg:  err.Error(),
		}, nil
	}
	// Concurrent sends to other nodes
	mapreduce.MapVoid(func(source chan<- interface{}) {
		for k1 := range clientMap {
			source <- &BatchData{ip: k1, clientIds: clientMap[k1]}
		}
	}, func(item interface{}) {
		batchData := item.(*BatchData)
		connect, ok := g.nodeIpMap.Load(batchData.ip)
		if ok {
			conn, ok := connect.(gateway.GatewayServiceClient)
			if !ok {
				return
			}
			ids := make([]string, 0)
			for k1 := range batchData.clientIds {
				ids = append(ids, batchData.clientIds[k1])
			}
			_, err = conn.SendToClientIds(context.Background(), &gateway.SendToClientIdsRequest{
				ClientIds: ids,
				Data:      req.Data,
			})
			if err != nil {
				logx.LogHandler.Error(err)
			}
		} else {
			logx.LogHandler.Error("node:%s not online", batchData.ip)
			return
		}

	})

	return &gateway.SendToClientIdsResponse{}, nil
}
func (g *GatewayClient) IsOnline(clientId string) (bool, error) {
	return g.sessionStorageHandler.IsOnline(clientId), nil
}
