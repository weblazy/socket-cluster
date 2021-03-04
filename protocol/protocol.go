package protocol

import (
	"io"
)

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
	// Proto pack/unpack protocol scheme of socket message.
	// NOTE: Implementation specifications for Message interface should be complied with.
	Proto interface {
		// Version returns the protocol's id and name.
		Version() (byte, string)
		// Pack writes the Message into the connection.
		// NOTE: Make sure to write only once or there will be package contamination!
		Pack(Message) error
		// Unpack reads bytes from the connection to the Message.
		// NOTE: Concurrent unsafe!
		Unpack(Message) error
	}
	// IOWithReadBuffer implements buffered I/O with buffered reader.
	IOWithReadBuffer interface {
		io.ReadWriter
	}
	// ProtoFunc function used to create a custom Proto interface.
	ProtoFunc func(IOWithReadBuffer) Proto
)

type Protocol interface {
	// ListenAndServe turns on the listening service.
	ListenAndServe(port int64, nodeHandler Node, protoFunc ...ProtoFunc) error
	// Dial connects with the peer of the destination address.
	Dial(addr string, protoFunc ...ProtoFunc) (Connection, error)
	// ServeConn serves the connection and returns a session.
	// NOTE:
	//  Not support automatically redials after disconnection;
	//  Not check TLS;
	//  Execute the PostAcceptPlugin plugins.
	// ServeConn(conn net.Conn, protoFunc ...ProtoFunc) (Session, error)
}
