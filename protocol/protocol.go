package protocol

type (
	// Proto pack/read protocol scheme of socket message.
	Proto interface {
		// Pack writes the Message into the connection.
		// NOTE: Make sure to write only once or there will be package contamination!
		Pack([]byte) ([]byte, error)
		// Read bytes from the connection.
		Read(conn FlowConnection, f func(conn Connection, msg []byte)) error
	}
)

type Protocol interface {
	// SetNodeHandler sets nodeHandler.
	SetNodeHandler(nodeHandler Node)
	// ListenAndServe turns on the listening service.
	ListenAndServe(port int64) error
	// Dial connects with the socket of the destination address.
	Dial(addr string) (Connection, error)
	// ServeConn serves the connection.
	ServeConn(conn Connection, handler func(conn Connection, msg []byte)) error
}
