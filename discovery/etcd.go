package discovery

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
)

const defaultDialTimeout = 5 * time.Second

//EtcdDiscovery 服务发现
type EtcdDiscovery struct {
	ServiceDiscovery
	cli        *clientv3.Client  //etcd client
	serverList map[string]string //服务列表
	lock       sync.Mutex
	lease      int64
	key        string
	val        string           //value
	leaseID    clientv3.LeaseID //租约ID
	//租约keepalieve相应chan
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
}

//NewEtcdDiscovery  新建发现服务
func NewEtcdDiscovery(conf clientv3.Config) *EtcdDiscovery {
	cli, err := clientv3.New(conf)
	if err != nil {
		log.Fatal(err)
	}
	return &EtcdDiscovery{
		cli:        cli,
		serverList: make(map[string]string),
	}
}

//WatchService 初始化服务列表和监视
func (s *EtcdDiscovery) GetServices(key string) error {
	//根据前缀获取现有的key
	// resp, err := s.cli.Get(context.Background(), key, clientv3.WithPrefix())
	// if err != nil {
	// 	return err
	// }

	// for _, ev := range resp.Kvs {
	// 	s.SetServiceList(string(ev.Key), string(ev.Value))
	// }
	return nil
}

//watcher 监听前缀
func (s *EtcdDiscovery) WatchService(watchChan WatchChan) {
	prefix := s.key
	rch := s.cli.Watch(context.Background(), prefix, clientv3.WithPrefix())
	log.Printf("watching prefix:%s now...", prefix)
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch ev.Type {
			case mvccpb.PUT: //修改或者新增
				watchChan <- PUT
			case mvccpb.DELETE: //删除

			}
		}
	}
}

//Close 关闭服务
func (s *EtcdDiscovery) Close() error {
	return s.cli.Close()
}

//设置租约
func (s *EtcdDiscovery) Register() error {
	//设置租约时间
	resp, err := s.cli.Grant(context.Background(), s.lease)
	if err != nil {
		return err
	}
	//注册服务并绑定租约
	_, err = s.cli.Put(context.Background(), s.key, "string(value)", clientv3.WithLease(resp.ID))
	if err != nil {
		return err
	}
	//设置续租 定期发送需求请求
	leaseRespChan, err := s.cli.KeepAlive(context.Background(), resp.ID)

	if err != nil {
		return err
	}
	s.leaseID = resp.ID
	log.Println(s.leaseID)
	s.keepAliveChan = leaseRespChan
	go s.ListenLeaseRespChan()
	log.Printf("Put key:%s  val:%s  success!", s.key, s.val)
	return nil
}

//ListenLeaseRespChan 监听 续租情况
func (s *EtcdDiscovery) ListenLeaseRespChan() {
	for leaseKeepResp := range s.keepAliveChan {
		log.Println("续约成功", leaseKeepResp)
	}
	log.Println("关闭续租")
}
