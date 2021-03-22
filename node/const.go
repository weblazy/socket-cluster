package node

import (
	"time"

	proto "github.com/golang/protobuf/proto"
)

const (
	clientPrefix              = "client#"
	defaultPassword           = "password"
	defaultClientPingInterval = 120
	defaultNodePingInterval   = 10
	defaultInternalPort       = 9527
	defaultPort               = 9528
	NodeAddress               = "node_address"
	authTime                  = 10 * time.Second
)

const (
	AuthMsgType   int32 = 1
	ClientMsgType int32 = 2
	PingMsgType   int32 = 3
)

var PingMsg []byte

func init() {
	msg := Msg{MsgType: PingMsgType}
	PingMsg, _ = proto.Marshal(&msg)
}
