package redis_discovery

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/cast"
	"github.com/weblazy/easy/utils/logx"
	easy_sort "github.com/weblazy/easy/utils/sort"
	"github.com/weblazy/socket-cluster/discovery"
	"github.com/weblazy/socket-cluster/node"
)

//RedisDiscovery 服务发现
type RedisDiscovery struct {
	discovery.ServiceDiscovery
	nodeId     string
	adminRedis *redis.Client
	key        string
	timeout    int64
}

//NewRedisDiscovery 新建发现服务
func NewRedisDiscovery(conf *redis.Options) *RedisDiscovery {
	rds := redis.NewClient(conf)
	return &RedisDiscovery{
		adminRedis: rds,
		timeout:    120,
	}
}

// Consumer pull message from other node
func (this *RedisDiscovery) SetNodeId(nodeId string) {
	this.nodeId = nodeId
}

// Consumer pull message from other node
func (this *RedisDiscovery) WatchService(watchChan discovery.WatchChan) {
	go func() {
		pb := this.adminRedis.Subscribe(context.Background(), this.key)
		for mg := range pb.Channel() {
			data := make(map[string]interface{})
			err := json.Unmarshal([]byte(mg.Payload), &data)
			if err != nil {
				logx.Info(err)
				return
			}
			watchChan <- discovery.PUT
		}
	}()
}

//Close 关闭服务
func (s *RedisDiscovery) Close() error {
	return s.adminRedis.Close()
}

//设置租约
func (this *RedisDiscovery) Register() error {
	err := this.adminRedis.Publish(context.Background(), this.key, this.nodeId).Err()
	if err != nil {
		logx.Info(err)
	}
	return nil
}

func (this *RedisDiscovery) UpdateInfo(nodeInfoByte []byte) error {
	err := this.adminRedis.HSet(context.Background(), node.NodeAddress, this.nodeId, string(nodeInfoByte)).Err()
	if err != nil {
		logx.Info(err)
	}
	this.adminRedis.Expire(context.Background(), node.NodeAddress, time.Duration(this.timeout*int64(time.Second)))
	return nil
}

// GetHosts get node address
func (this *RedisDiscovery) GetInfo() ([]string, error) {
	list := make([]string, 0)
	now := time.Now().Unix()
	addrMap, err := this.adminRedis.HGetAll(context.Background(), node.NodeAddress).Result()
	if err != nil {
		return nil, err
	}
	sortList := make([]easy_sort.Sort, 0)
	expire := now - this.timeout
	for k1 := range addrMap {
		addrObj := make(map[string]interface{})
		err := json.Unmarshal([]byte(addrMap[k1]), &addrObj)
		if err != nil {
			return nil, err
		}
		if cast.ToInt64(addrObj["timestamp"]) > expire {
			sortList = append(sortList, easy_sort.Sort{
				Obj:  k1,
				Sort: cast.ToInt64(addrObj["client_count"]),
			})
		}
	}
	sort.Sort(easy_sort.SortList(sortList))
	for k1 := range sortList {
		list = append(list, sortList[k1].Obj.(string))
	}
	return list, nil
}
