package websocket_cluster

type (
	Msg struct {
		MsgType string      `json:"msg_type"`
		Data    interface{} `json:"data"`
	}
)
