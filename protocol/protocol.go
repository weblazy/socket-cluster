package protocol

type (
	// Message a socket message interface.
	Message interface {
	}

	// Header is an operation interface of required message fields.
	// NOTE: Must be supported by Proto interface.
	Header interface {
	}

	// Body is an operation interface of optional message fields.
	// SUGGEST: For features complete, the protocol interface should support it.
	Body interface{}
)
type (
	// Proto pack/read protocol scheme of socket message.
	Proto interface {
		// Pack writes the Message into the connection.
		// NOTE: Make sure to write only once or there will be package contamination!
		Pack([]byte) ([]byte, error)
		// Read bytes from the connection.
		Read(conn Connection, f func(conn Connection, msg []byte)) error
	}
)

type Protocol interface {
	// ListenAndServe turns on the listening service.
	ListenAndServe(port int64, nodeHandler Node) error
	// Dial connects with the socket of the destination address.
	Dial(addr string) (Connection, error)
	// ServeConn serves the connection and returns a session.
	// NOTE:
	//  Not support automatically redials after disconnection;
	//  Not check TLS;
	//  Execute the PostAcceptPlugin plugins.
	ServeConn(conn Connection, OnTransMsg func(conn Connection, msg []byte)) error
}
