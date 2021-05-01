package business_client

import (
	"time"

	"github.com/weblazy/core/syncx"
	"github.com/weblazy/core/timingwheel"
	"github.com/weblazy/goutil"
	"github.com/weblazy/socket-cluster/logx"
	"github.com/weblazy/socket-cluster/protocol"
)

type (
	BusinessClient interface {
		// IsOnline gets the online status of the clientId
		IsOnline(clientId int64) bool
		// SendToClientId sends a message to a clientId
		SendToClientId(clientId string, req []byte) error
		//  SendToClientId sends a message to multiple clientIds
		SendToClientIds(clientIds []string, req []byte) error
	}

	// Node communication node
	businessClient struct {
		BusinessClient
		businessClientConf *BusinessClientConf
		// clientIdSessions map[key1]map[key2]value
		// Key1: clientId
		// Key2: socket address
		// Value: *Session
		clientIdSessions *syncx.ConcurrentDoubleMap
		// Key: socket address
		// Value: *Session
		clientConns goutil.Map
		timer       *timingwheel.TimingWheel // Timingwheel Close connects that timeout without authentication
		startTime   time.Time                // Start time
		// Key: nodeId
		// Value: *Session
		transClients goutil.Map // Forward the message to another node
		// Key: socket address
		// Value: protocol.Connection
		transServices goutil.Map // Receive messages forwarded by other nodes
		nodeTimeout   int64      // Node heartbeat timeout time
		clientTimeout int64      // Client heartbeat timeout time

		internalProtocolHandler protocol.Protocol // Direct protocol between node and node
	}
)

// NewNode return a Node
func NewBusinessClient(cfg *BusinessClientConf) (BusinessClient, error) {
	timer, err := timingwheel.NewTimingWheel(time.Second, 30, func(k, v interface{}) {
		// TODO Count the number of unauthenticated connections
		logx.LogHandler.Infof("%s auth timeout", k)
		if v.(protocol.Connection) != nil {
			err := v.(protocol.Connection).Close()
			if err != nil {
				logx.LogHandler.Error(err)
			}
		}
	})

	if err != nil {
		return nil, err
	}
	businessClientObj := &businessClient{
		businessClientConf: cfg,
		clientIdSessions:   syncx.NewConcurrentDoubleMap(32),
		startTime:          time.Now(),
		timer:              timer,
		clientConns:        goutil.AtomicMap(),
		transClients:       goutil.AtomicMap(),
		transServices:      goutil.AtomicMap(),
		nodeTimeout:        cfg.nodePingInterval * 3,
	}

	return businessClientObj, nil
}

func IsOnline(clientId int64) bool {
	return false
}

func SendToClientId(clientId string, req []byte) error {

	return nil
}

func SendToClientIds(clientIds []string, req []byte) error {
	return nil
}
