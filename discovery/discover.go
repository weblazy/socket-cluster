package discovery

type EventType int32

const (
	PUT    EventType = 0
	DELETE EventType = 1
)

type WatchChan chan EventType

// ServiceDiscovery
type ServiceDiscovery interface {
	// SetNodeId sets nodeId
	SetNodeId(nodeId string)
	// WatchService Listens for a new node to start
	WatchService(watchChan WatchChan)
	// UpdateInfo Updates the information for this node
	UpdateInfo([]byte) error
	// Register registers the NodeID and notify other nodes
	Register() error
	// WatchService Listens for a new node to start
	GetServerList() (map[string]string, error)
}
