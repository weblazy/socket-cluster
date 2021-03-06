package discovery

type EventType int32

const (
	PUT    EventType = 0
	DELETE EventType = 1
)

type WatchChan chan EventType

//ServiceDiscovery 服务发现
type ServiceDiscovery interface {
	//设置租约
	// GetServices() ([]string, error)
	SetNodeId(nodeId string)
	WatchService(watchChan WatchChan)
	UpdateInfo([]byte) error
	Register() error
	// func (s *ServiceDiscovery) Ping(value []byte)
	// func (s *ServiceDiscovery) Close() error
}
