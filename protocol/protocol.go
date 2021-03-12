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
	// ListenAndServe turns on the listening service.
	ListenAndServe(port int64) error

	// ListenAndServe turns on the listening service.
	SetNodeHandler(nodeHandler Node)
	// Dial connects with the socket of the destination address.
	Dial(addr string) (Connection, error)
	// ServeConn serves the connection and returns a session.
	// NOTE:
	//  Not support automatically redials after disconnection;
	//  Not check TLS;
	//  Execute the PostAcceptPlugin plugins.
	ServeConn(conn Connection, OnTransMsg func(conn Connection, msg []byte)) error
}
