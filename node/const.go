package node

import (
	"time"

	proto "github.com/golang/protobuf/proto"
)

const (
	clientPrefix              = "client#"
	DefaultPassword           = "password"
	defaultClientPingInterval = 120
	DefaultNodePingInterval   = 10
	defaultInternalPort       = 9528
	defaultPort               = 9527
	defaultAddr               = "127.0.0.1:9527"
	NodeAddress               = "node_address"
	authTime                  = 10 * time.Second
)

const (
	AuthNodeMsgType           int32 = 1
	ClientMsgType             int32 = 2
	PingMsgType               int32 = 3
	AuthBusinessClientMsgType int32 = 4
	BindNodeIdMsgType         int32 = 5
)

var PingMsg []byte

func init() {
	msg := Msg{MsgType: PingMsgType}
	PingMsg, _ = proto.Marshal(&msg)
}
