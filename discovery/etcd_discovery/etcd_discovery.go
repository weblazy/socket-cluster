package etcd_discovery

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/weblazy/socket-cluster/discovery"
)

const defaultDialTimeout = 5 * time.Second

// EtcdDiscovery
type EtcdDiscovery struct {
	discovery.ServiceDiscovery
	cli           *clientv3.Client  // etcd client
	serverList    map[string]string // service list
	lock          sync.Mutex
	lease         int64
	nodeId        string
	key           string
	val           string                                  // value
	leaseID       clientv3.LeaseID                        // 租约ID
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse // chan for renewal of lease
}

// NewEtcdDiscovery return a EtcdDiscovery
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

// SetNodeId sets a nodeId
func (this *EtcdDiscovery) SetNodeId(nodeId string) {
	this.nodeId = nodeId
}

// GetServices gets all Node nodes
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

// WatchService Listens for a new node to start
func (s *EtcdDiscovery) WatchService(watchChan discovery.WatchChan) {
	prefix := s.key
	rch := s.cli.Watch(context.Background(), prefix, clientv3.WithPrefix())
	log.Printf("watching prefix:%s now...", prefix)
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch ev.Type {
			case mvccpb.PUT: // Modify or add
				watchChan <- discovery.PUT
			case mvccpb.DELETE: // delete

			}
		}
	}
}

//Close closes the etcd client
func (s *EtcdDiscovery) Close() error {
	return s.cli.Close()
}

// Register registers the nodeId and notify other nodes
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

// ListenLeaseRespChan 监听 续租情况
func (s *EtcdDiscovery) ListenLeaseRespChan() {
	for leaseKeepResp := range s.keepAliveChan {
		log.Println("续约成功", leaseKeepResp)
	}
	log.Println("关闭续租")
}
