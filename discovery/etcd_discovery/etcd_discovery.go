package etcd_discovery

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/weblazy/socket-cluster/discovery"
	"github.com/weblazy/socket-cluster/logx"
	"github.com/weblazy/socket-cluster/node"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const defaultDialTimeout = 5 * time.Second

// EtcdDiscovery
type EtcdDiscovery struct {
	discovery.ServiceDiscovery
	cli           *clientv3.Client  // etcd client
	serverList    map[string]string // service list
	lock          sync.Mutex
	lease         int64
	nodeAddr      string
	key           string
	val           string // value
	leaseID       clientv3.LeaseID
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
		key:        node.NodeAddress,
	}
}

// SetNodeAddr sets a nodeAddr
func (this *EtcdDiscovery) SetNodeAddr(nodeAddr string) {
	this.nodeAddr = nodeAddr
}

// WatchService Listens for a new node to start
func (s *EtcdDiscovery) WatchService(watchChan discovery.WatchChan) {
	prefix := s.key
	rch := s.cli.Watch(context.Background(), prefix, clientv3.WithPrefix())
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

// Close closes the etcd client
func (s *EtcdDiscovery) Close() error {
	return s.cli.Close()
}

// Register registers the nodeId and notify other nodes
func (s *EtcdDiscovery) Register() error {
	//sets the lease time
	resp, err := s.cli.Grant(context.Background(), s.lease)
	if err != nil {
		return err
	}
	// register and bind lease
	_, err = s.cli.Put(context.Background(), s.key, "string(value)", clientv3.WithLease(resp.ID))
	if err != nil {
		return err
	}
	// Set up renewal time to send renewal request
	leaseRespChan, err := s.cli.KeepAlive(context.Background(), resp.ID)

	if err != nil {
		return err
	}
	s.leaseID = resp.ID
	s.keepAliveChan = leaseRespChan
	go s.ListenLeaseRespChan()
	return nil
}

func (s *EtcdDiscovery) GetServerList() (map[string]string, error) {
	resp, err := s.cli.Get(context.Background(), s.key)
	if err != nil {
		return nil, err
	}
	for _, obj := range resp.Kvs {
		s.serverList[string(obj.Key)] = string(obj.Value)
	}
	return s.serverList, nil
}

// ListenLeaseRespChan Monitor lease renewals
func (s *EtcdDiscovery) ListenLeaseRespChan() {
	for leaseKeepResp := range s.keepAliveChan {
		logx.LogHandler.Info("续约成功", leaseKeepResp)
	}
	logx.LogHandler.Infof("关闭续租")
}
