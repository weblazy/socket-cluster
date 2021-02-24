package redis_storage

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/cast"
	"github.com/weblazy/socket-cluster/session_storage"
	"github.com/weblazy/socket-cluster/unsafehash"
)

type RedisStorage struct {
	session_storage.SessionStorage
	clientTimeout int64 // client heartbeat timeout time
	segmentMap    *unsafehash.SegmentMap
	transAddress  string
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
	return &RedisStorage{segmentMap: segmentMap}
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
	err := redisNode.Extra.(*redis.Client).ZAdd(context.Background(), session_storage.ClientPrefix+cast.ToString(clientId), &redis.Z{Score: cast.ToFloat64(now), Member: this.transAddress}).Err()
	if err != nil {
		return err
	}
	err = redisNode.Extra.(*redis.Client).Expire(context.Background(), session_storage.ClientPrefix+cast.ToString(clientId), time.Duration(this.clientTimeout)*time.Second).Err()
	if err != nil {
		return err
	}
	return nil
}
