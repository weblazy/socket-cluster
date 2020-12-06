package websocket_cluster

type (
	Msg struct {
		Uid     string      `json:"uid"`
		MsgType string      `json:"msg_type"`
		Data    interface{} `json:"data"`
	}
)
