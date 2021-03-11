package redis_storage

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/cast"
	"github.com/weblazy/core/logx"
	"github.com/weblazy/socket-cluster/session_storage"
	"github.com/weblazy/socket-cluster/unsafehash"
)

type RedisStorage struct {
	session_storage.SessionStorage
	clientTimeout int64 // client heartbeat timeout time
	segmentMap    *unsafehash.SegmentMap
	nodeId        string
}

type RedisNode struct {
	RedisConf *redis.Options
	Position  int64 //the position of hash ring
}

func NewRedisStorage(redisNodeList []*RedisNode) *RedisStorage {
	segmentMap := unsafehash.NewSegmentMap()
	for _, value := range redisNodeList {
		rdsObj := redis.NewClient(value.RedisConf)
		segmentMap.Add(unsafehash.NewNode(value.RedisConf.Addr, value.Position, rdsObj))
	}
	return &RedisStorage{segmentMap: segmentMap, clientTimeout: 180}
}

func (this *RedisStorage) SetNodeId(nodeId string) {
	this.nodeId = nodeId
}

func (this *RedisStorage) GetIps(clientId int64) ([]string, error) {
	redisNode := this.segmentMap.Get(clientId)
	now := time.Now().Unix()
	ipArr, err := redisNode.Extra.(*redis.Client).ZRangeByScore(context.Background(), session_storage.ClientPrefix+cast.ToString(clientId), &redis.ZRangeBy{Min: cast.ToString(now - this.clientTimeout), Max: "+inf"}).Result()
	return ipArr, err
}

func (this *RedisStorage) BindClientId(clientId int64) error {
	now := time.Now().Unix()
	redisNode := this.segmentMap.Get(clientId)
	logx.Error(clientId)
	err := redisNode.Extra.(*redis.Client).ZAdd(context.Background(), session_storage.ClientPrefix+cast.ToString(clientId), &redis.Z{Score: cast.ToFloat64(now), Member: this.nodeId}).Err()
	if err != nil {
		logx.Error(err.Error())
		return err
	}

	err = redisNode.Extra.(*redis.Client).Expire(context.Background(), session_storage.ClientPrefix+cast.ToString(clientId), time.Duration(this.clientTimeout)*time.Second).Err()
	if err != nil {
		logx.Error(err.Error())
		return err
	}

	return nil
}

type NodeMap struct {
	node      *unsafehash.Node
	clientIds []int64
}

// ClientIdsOnline Get online users in the group
func (this *RedisStorage) ClientIdsOnline(clientIds []int64) []int64 {
	// now := time.Now().Unix()
	onlineClientIds := make([]int64, 0)
	nodes := make(map[string]*NodeMap)
	for k1 := range clientIds {
		redisNode := this.segmentMap.Get(clientIds[k1])
		if _, ok := nodes[redisNode.Id]; ok {
			nodes[redisNode.Id].clientIds = append(nodes[redisNode.Id].clientIds, clientIds[k1])
		} else {
			nodes[redisNode.Id] = &NodeMap{
				node:      redisNode,
				clientIds: []int64{clientIds[k1]},
			}
		}
	}
	rangeTime := cast.ToString(time.Now().Unix() - this.clientTimeout)
	for k1 := range nodes {
		nodeMap := nodes[k1]
		pipe := nodeMap.node.Extra.(*redis.Client).Pipeline()
		for k2 := range nodeMap.clientIds {
			pipe.ZRangeByScore(context.Background(), session_storage.ClientPrefix+cast.ToString(nodeMap.clientIds[k2]), &redis.ZRangeBy{Min: rangeTime, Max: "+inf"}).Result()
		}
		cmders, err := pipe.Exec(context.Background())
		if err != nil {
			logx.Info(err)
		}
		for k3, cmder := range cmders {
			cmd := cmder.(*redis.StringSliceCmd)
			err := cmd.Err()
			if err != nil {
				logx.Info(err)
			} else {
				onlineClientIds = append(onlineClientIds, nodeMap.clientIds[k3])
			}
		}
	}
	return onlineClientIds
}

// IsOnline determine if a clientId is online
func (this *RedisStorage) IsOnline(clientId int64) bool {
	now := time.Now().Unix()
	redisNode := this.segmentMap.Get(clientId)
	addrArr, err := redisNode.Extra.(*redis.Client).ZRangeByScore(context.Background(), session_storage.ClientPrefix+cast.ToString(clientId), &redis.ZRangeBy{Min: cast.ToString(now - this.clientTimeout), Max: "+inf"}).Result()
	if err != nil {
		logx.Info(err)
		return false
	}
	if len(addrArr) > 0 {
		return true
	}
	return false
}

// OnClientPing receive client heartbeat
func (this *RedisStorage) OnClientPing(clientId int64) error {
	redisNode := this.segmentMap.Get(clientId)
	now := time.Now().Unix()
	err := redisNode.Extra.(*redis.Client).ZAdd(context.Background(), session_storage.ClientPrefix+cast.ToString(clientId), &redis.Z{Score: cast.ToFloat64(now), Member: this.nodeId}).Err()
	if err != nil {
		return err
	}
	err = redisNode.Extra.(*redis.Client).Expire(context.Background(), session_storage.ClientPrefix+cast.ToString(clientId), time.Duration(this.clientTimeout)*time.Second).Err()
	return err
}

func (this *RedisStorage) GetClientsIps(clientIds []string) ([]string, map[string][]string, error) {

	nodes := make(map[string]*NodeMap)
	rangeTime := cast.ToString(time.Now().Unix() - this.clientTimeout)
	for k1 := range clientIds {
		redisNode := this.segmentMap.Get(cast.ToInt64(clientIds[k1]))
		if _, ok := nodes[redisNode.Id]; ok {
			nodes[redisNode.Id].clientIds = append(nodes[redisNode.Id].clientIds, cast.ToInt64(clientIds[k1]))
		} else {
			nodes[redisNode.Id] = &NodeMap{
				node:      redisNode,
				clientIds: []int64{cast.ToInt64(clientIds[k1])},
			}
		}
	}
	otherMap := make(map[string][]string)
	localClientIds := make([]string, 0)
	for k1 := range nodes {
		nodeMap := nodes[k1]
		pipe := nodeMap.node.Extra.(*redis.Client).Pipeline()
		for k2 := range nodeMap.clientIds {
			pipe.ZRangeByScore(context.Background(), session_storage.ClientPrefix+cast.ToString(nodeMap.clientIds[k2]), &redis.ZRangeBy{Min: rangeTime, Max: "+inf"}).Result()
		}
		cmders, err := pipe.Exec(context.Background())
		if err != nil {
			logx.Info(cmders, err)
		}
		for k3, cmder := range cmders {
			cmd := cmder.(*redis.StringSliceCmd)
			strMap, err := cmd.Result()
			if err != nil {
				logx.Info(err)
			} else {
				for k4 := range strMap {
					if strMap[k4] == this.nodeId {
						localClientIds = append(localClientIds, cast.ToString(nodeMap.clientIds[k3]))
					} else {
						if _, ok := otherMap[strMap[k4]]; ok {
							otherMap[strMap[k4]] = append(otherMap[strMap[k4]], cast.ToString(nodeMap.clientIds[k3]))
						} else {
							otherMap[strMap[k4]] = []string{cast.ToString(nodeMap.clientIds[k3])}
						}
					}
				}

			}
		}
	}
	return localClientIds, otherMap, nil
}
