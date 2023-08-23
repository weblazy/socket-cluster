package redis_storage

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/cast"
	"github.com/weblazy/easy/db/eredis"
	"github.com/weblazy/socket-cluster/logx"
	"github.com/weblazy/socket-cluster/session_storage"
)

type RedisStorage struct {
	session_storage.SessionStorage
	clientTimeout int64 // client heartbeat timeout time
	redisClient   *eredis.RedisClient
}

type RedisNode struct {
	RedisConf *redis.Options
	Position  int64 //the position of hash ring
}

// NewRedisStorage return a RedisStorage
func NewRedisStorage(redisClient *eredis.RedisClient) *RedisStorage {
	return &RedisStorage{redisClient: redisClient, clientTimeout: 360}
}

// GetIps get ip list by clientId
func (this *RedisStorage) GetIps(clientId string) ([]string, error) {
	now := time.Now().Unix()
	ipArr, err := this.redisClient.ZRangeByScore(context.Background(), session_storage.ClientPrefix+clientId, &redis.ZRangeBy{Min: cast.ToString(now - this.clientTimeout), Max: "+inf"}).Result()
	return ipArr, err
}

// BindClientId set online with clientId
func (this *RedisStorage) BindClientId(nodeId string, clientId string) error {
	now := time.Now().Unix()
	err := this.redisClient.ZAdd(context.Background(), session_storage.ClientPrefix+clientId, &redis.Z{Score: cast.ToFloat64(now), Member: nodeId}).Err()
	if err != nil {
		return err
	}

	err = this.redisClient.Expire(context.Background(), session_storage.ClientPrefix+clientId, time.Duration(this.clientTimeout)*time.Second).Err()
	if err != nil {
		return err
	}

	return nil
}

// ClientIdsOnline Get online users in the group
func (this *RedisStorage) ClientIdsOnline(clientIds []string) []string {
	onlineClientIds := make([]string, 0)
	rangeTime := cast.ToString(time.Now().Unix() - this.clientTimeout)

	pipe := this.redisClient.Pipeline()
	for k2 := range clientIds {
		pipe.ZRangeByScore(context.Background(), session_storage.ClientPrefix+clientIds[k2], &redis.ZRangeBy{Min: rangeTime, Max: "+inf"}).Result()
	}
	cmders, err := pipe.Exec(context.Background())
	if err != nil {
		logx.LogHandler.Error(err)
	}
	for k3, cmder := range cmders {
		cmd := cmder.(*redis.StringSliceCmd)
		err := cmd.Err()
		if err != nil {
			logx.LogHandler.Error(err)
		} else {
			onlineClientIds = append(onlineClientIds, clientIds[k3])
		}
	}

	return onlineClientIds
}

// ClientIdsOnlineWithLua Get online users in the group
func (this *RedisStorage) ClientIdsOnlineWithLua(clientIds []string) []string {
	onlineClientIds := make([]string, 0)
	return onlineClientIds
}

// IsOnline determine if a clientId is online
func (this *RedisStorage) IsOnline(clientId string) bool {
	now := time.Now().Unix()
	addrArr, err := this.redisClient.ZRangeByScore(context.Background(), session_storage.ClientPrefix+clientId, &redis.ZRangeBy{Min: cast.ToString(now - this.clientTimeout), Max: "+inf"}).Result()
	if err != nil {
		logx.LogHandler.Error(err)
		return false
	}
	if len(addrArr) > 0 {
		return true
	}
	return false
}

// OnClientPing receive client heartbeat
func (this *RedisStorage) OnClientPing(nodeId string, clientId string) error {
	now := time.Now().Unix()
	err := this.redisClient.ZAdd(context.Background(), session_storage.ClientPrefix+clientId, &redis.Z{Score: cast.ToFloat64(now), Member: nodeId}).Err()
	if err != nil {
		return err
	}
	err = this.redisClient.Expire(context.Background(), session_storage.ClientPrefix+clientId, time.Duration(this.clientTimeout)*time.Second).Err()
	return err
}

// GetIps get ip list by clientId list
func (this *RedisStorage) GetClientsIps(clientIds []string) (map[string][]string, error) {
	rangeTime := cast.ToString(time.Now().Unix() - this.clientTimeout)

	pipe := this.redisClient.Pipeline()
	for k2 := range clientIds {
		pipe.ZRangeByScore(context.Background(), session_storage.ClientPrefix+clientIds[k2], &redis.ZRangeBy{Min: rangeTime, Max: "+inf"}).Result()
	}
	cmders, err := pipe.Exec(context.Background())
	if err != nil {
		logx.LogHandler.Error(cmders, err)
	}
	otherMap := make(map[string][]string)
	for k3, cmder := range cmders {
		cmd := cmder.(*redis.StringSliceCmd)
		strMap, err := cmd.Result()
		if err != nil {
			logx.LogHandler.Error(err)
		} else {
			for k4 := range strMap {
				if _, ok := otherMap[strMap[k4]]; ok {
					otherMap[strMap[k4]] = append(otherMap[strMap[k4]], clientIds[k3])
				} else {
					otherMap[strMap[k4]] = []string{clientIds[k3]}
				}

			}

		}
	}

	return otherMap, nil
}
