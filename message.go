package websocket_cluster

type (
	Message struct {
		Uid         string      `json:"uid"`
		MessageType string      `json:"message_type"`
		Data        interface{} `json:"data"`
	}
)
