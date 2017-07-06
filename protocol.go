package netproto

import (
	"io"
	"net"
)

// StreamID is the ID for a stream.
type StreamID uint32

// Protocol generalizes an available reliable stream-based connection stream protocol.
type Protocol interface {
	// Listen listens to an address.
	Listen(addr string) (Listener, error)
	// Dial dials an address.
	Dial(addr string) (Session, error)
}

// Listener listens for connections.
type Listener interface {
	// AcceptSession accepts a session.
	AcceptSession() (Session, error)
	// Dial dials an address. This is an alias for proto.Dial
	Dial(addr string) (Session, error)
	// Addr returns the local address of the listener.
	Addr() net.Addr
	// Close closes the listener and all attached connections.
	Close() error
}

// Session manages an individual connection.
type Session interface {
	io.Closer

	// AcceptStream accepts an incoming stream from the connection.
	AcceptStream() (Stream, error)
	// OpenStream opens a new stream on the session.
	OpenStream() (Stream, error)
	// LocalAddr is our local address.
	LocalAddr() net.Addr
	// RemoteAddr is the address of the peer.
	RemoteAddr() net.Addr
	// Initiator returns if this session was built by Dial
	Initiator() bool
	// CloseWithError closes the session, transmitting an error if possible.
	// Only applicable for Quic.
	CloseWithError(err error) error
}

// Stream manages an individual stream in a session.
type Stream interface {
	io.Reader
	io.Writer

	// ID returns the ID of the stream.
	ID() StreamID
}
