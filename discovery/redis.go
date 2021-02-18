package discovery

import (
	"context"
	"encoding/json"

	"github.com/go-redis/redis/v8"
	"github.com/weblazy/core/logx"
)

//RedisDiscovery 服务发现
type RedisDiscovery struct {
	ServiceDiscovery
	transAddress string
	adminRedis   *redis.Client
	key          string
}

//NewRedisDiscovery 新建发现服务
func NewRedisDiscovery(conf *redis.Options) *RedisDiscovery {
	rds := redis.NewClient(conf)
	return &RedisDiscovery{
		adminRedis: rds,
	}
}

// Consumer pull message from other node
func (this *RedisDiscovery) WatchService(watchChan WatchChan) {
	go func() {
		pb := this.adminRedis.Subscribe(context.Background(), this.key)
		for mg := range pb.Channel() {
			data := make(map[string]interface{})
			err := json.Unmarshal([]byte(mg.Payload), &data)
			if err != nil {
				logx.Info(err)
				return
			}
			watchChan <- PUT

			if _, ok := data["receive_client_id"].(string); ok {
				delete(data, "receive_client_id")
			}
		}
	}()
}

//GetServices 获取服务地址
func (s *RedisDiscovery) GetServices() ([]string, error) {
	return []string{}, nil

}

//Close 关闭服务
func (s *RedisDiscovery) Close() error {
	return s.adminRedis.Close()
}

// func main() {
// 	var endpoints = []string{"42.192.166.82:2379"}
// 	ser := NewRedisDiscovery(endpoints)
// 	defer ser.Close()
// 	ser.WatchService("/web/")
// 	ser.WatchService("/gRPC/")
// 	for {
// 		select {
// 		case <-time.Tick(10 * time.Second):
// 			log.Println(ser.GetServices())
// 		}
// 	}
// }

//设置租约
func (this *RedisDiscovery) Register() error {
	err := this.adminRedis.HSet(context.Background(), this.key, this.transAddress, "value").Err()
	if err != nil {
		logx.Info(err)
	}
	return nil
}
