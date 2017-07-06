package quic

import (
	base "github.com/fuserobotics/netproto"
	proto "github.com/lucas-clemente/quic-go"
)

// protocol represents the KCP protocol.
type protocol struct {
	config *proto.Config
}

// NewQuic builds the Quic protocol instance.
func NewQuic(config *proto.Config) base.Protocol {
	if config == nil {
		config = &proto.Config{}
	}

	return &protocol{config: config}
}

// Listen listens to an address.
func (p *protocol) Listen(addr string) (base.Listener, error) {
	var err error
	var res listener
	res.protocol = p

	res.Listener, err = proto.ListenAddr(addr, p.config)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

// Dial connects to an address.
func (p *protocol) Dial(addr string) (base.Session, error) {
	res, err := proto.DialAddr(addr, p.config)
	if err != nil {
		return nil, err
	}

	return &session{Session: res, initiator: true}, nil
}

// listener represents a kcp listener
type listener struct {
	proto.Listener
	*protocol
}

// AcceptSession accepts a session.
func (l *listener) AcceptSession() (base.Session, error) {
	lst := l.Listener

	conn, err := lst.Accept()
	if err != nil {
		return nil, err
	}

	return &session{Session: conn}, nil
}

// session represents a quic session
type session struct {
	proto.Session
	initiator bool
}

// Initiator returns if this session was built by Dial
func (s *session) Initiator() bool {
	return s.initiator
}

// stream represents a quic stream
type stream struct {
	proto.Stream
}

// ID returns the identifier for the stream.
func (s *stream) ID() base.StreamID {
	return base.StreamID(s.Stream.StreamID())
}

// AcceptStream accepts an incoming stream from the connection.
func (s *session) AcceptStream() (base.Stream, error) {
	stm, err := s.Session.AcceptStream()
	if err != nil {
		return nil, err
	}

	return &stream{Stream: stm}, nil
}

// OpenStream opens a new stream on the session.
func (s *session) OpenStream() (base.Stream, error) {
	stm, err := s.Session.OpenStream()
	if err != nil {
		return nil, err
	}
	return &stream{Stream: stm}, nil
}

// CloseWithError closes the session, transmitting an error if possible.
// Only applicable for Quic.
func (s *session) CloseWithError(err error) error {
	return s.Session.Close(err)
}

// Close closes the session.
func (s *session) Close() error {
	return s.CloseWithError(nil)
}
