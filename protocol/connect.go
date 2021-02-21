package protocol

type Connection interface {
	WriteJSON(data interface{})
	WriteMsg(msgType int, data []byte)
}
