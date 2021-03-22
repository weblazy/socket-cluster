package protocol

type (
	// Proto pack/read protocol scheme of socket message.
	Proto interface {
		// Pack writes the Message into the connection.
		// NOTE: Make sure to write only once or there will be package contamination!
		Pack([]byte) ([]byte, error)
		// Read bytes from the connection.
		ReadMsg() ([]byte, error)
	}
)

type Protocol interface {
	// ListenAndServe turns on the listening service.
	ListenAndServe(port int64, onConnect func(conn Connection)) error
	// Dial connects with the socket of the destination address.
	Dial(addr string) (Connection, error)
}
