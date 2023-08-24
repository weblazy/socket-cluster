package redis_discovery

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/spf13/cast"
	"github.com/weblazy/easy/db/eredis"
	"github.com/weblazy/easy/sortx"
	"github.com/weblazy/socket-cluster/discovery"
	"github.com/weblazy/socket-cluster/logx"
	"github.com/weblazy/socket-cluster/node"
)

// RedisDiscovery
type RedisDiscovery struct {
	discovery.ServiceDiscovery
	nodeAddr   string
	adminRedis *eredis.RedisClient
	key        string
	timeout    int64
}

// NewRedisDiscovery return a RedisDiscovery
func NewRedisDiscovery(adminRedis *eredis.RedisClient) *RedisDiscovery {
	return &RedisDiscovery{
		adminRedis: adminRedis,
		timeout:    120,
		key:        node.NodeAddress,
	}
}

// SetNodeAddr sets a addr
func (this *RedisDiscovery) SetNodeAddr(addr string) {
	this.nodeAddr = addr
}

// WatchService Listens for a new node to start
func (this *RedisDiscovery) WatchService(watchChan discovery.WatchChan) {
	go func() {
		pb := this.adminRedis.Subscribe(context.Background(), this.key)
		for mg := range pb.Channel() {
			data := make(map[string]interface{})
			err := json.Unmarshal([]byte(mg.Payload), &data)
			if err != nil {
				logx.LogHandler.Error(err)
				return
			}
			watchChan <- discovery.PUT
		}
	}()
}

// Close closes the redis
func (s *RedisDiscovery) Close() error {
	return s.adminRedis.Close()
}

// Register registers the NodeID and notify other nodes
func (this *RedisDiscovery) Register() error {
	err := this.adminRedis.Publish(context.Background(), this.key, this.nodeAddr).Err()
	if err != nil {
		return err
	}
	return nil
}

// UpdateInfo Update the information for this node
func (this *RedisDiscovery) UpdateInfo(nodeInfoByte []byte) error {
	err := this.adminRedis.HSet(context.Background(), node.NodeAddress, this.nodeAddr, string(nodeInfoByte)).Err()
	if err != nil {
		return err
	}
	err = this.adminRedis.Expire(context.Background(), node.NodeAddress, time.Duration(this.timeout*int64(time.Second))).Err()
	if err != nil {
		return err
	}
	return nil
}

// GetInfo get node information
func (this *RedisDiscovery) GetInfo() ([]string, error) {
	list := make([]string, 0)
	now := time.Now().Unix()
	addrMap, err := this.adminRedis.HGetAll(context.Background(), node.NodeAddress).Result()
	if err != nil {
		return nil, err
	}
	sortList := sortx.NewSortList(sortx.DESC)
	expire := now - this.timeout
	for k1 := range addrMap {
		addrObj := make(map[string]interface{})
		err := json.Unmarshal([]byte(addrMap[k1]), &addrObj)
		if err != nil {
			return nil, err
		}
		if cast.ToInt64(addrObj["timestamp"]) > expire {
			sortList.List = append(sortList.List, sortx.Sort{
				Obj:  k1,
				Sort: cast.ToFloat64(addrObj["client_count"]),
			})
		}
	}
	sort.Sort(sortList)
	for k1 := range sortList.List {
		list = append(list, sortList.List[k1].Obj.(string))
	}
	return list, nil
}

func (this *RedisDiscovery) GetServerList() (map[string]string, error) {
	addrMap, err := this.adminRedis.HGetAll(context.Background(), node.NodeAddress).Result()
	if err != nil {
		return nil, err
	}
	resp := map[string]string{}
	for k, v := range addrMap {
		resp[k] = v
	}
	return resp, nil
}
