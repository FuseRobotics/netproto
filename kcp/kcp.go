package kcp

import (
	"net"

	base "github.com/fuserobotics/netproto"
	proto "github.com/xtaci/kcp-go"
	"github.com/xtaci/smux"
)

// protocol represents the KCP protocol.
type protocol struct {
	listenOptions *ListenOptions
	smuxConfig    *smux.Config
}

// NewKCP builds the KCP protocol instance. Options are optional.
func NewKCP(opts *ListenOptions, smuxConf *smux.Config) base.Protocol {
	var optsc ListenOptions

	if smuxConf == nil {
		smuxConf = smux.DefaultConfig()
	}

	res := &protocol{smuxConfig: smuxConf}
	if opts != nil {
		optsc = *opts
		res.listenOptions = &optsc
	}

	return res
}

// ListenOptions represents possible options when listening with KCP.
type ListenOptions struct {
	// Block is the BlockCrypt for the listener
	Block proto.BlockCrypt
	// DataShards are the number of data shards to use.
	DataShards int
	// ParityShards are the number of parity shards to use.
	ParityShards int
}

// Listen listens to an address.
func (p *protocol) Listen(addr string) (base.Listener, error) {
	var err error
	var res listener
	res.protocol = p

	if p.listenOptions == nil {
		res.listener, err = proto.Listen(addr)
	} else {
		res.listener, err = proto.ListenWithOptions(addr, p.listenOptions.Block, p.listenOptions.DataShards, p.listenOptions.ParityShards)
	}
	if err != nil {
		return nil, err
	}

	return &res, nil
}

// Dial connects to an address.
func (p *protocol) Dial(addr string) (base.Session, error) {
	var err error
	var res net.Conn

	if p.listenOptions == nil {
		res, err = proto.Dial(addr)
	} else {
		res, err = proto.DialWithOptions(addr, p.listenOptions.Block, p.listenOptions.DataShards, p.listenOptions.ParityShards)
	}
	if err != nil {
		return nil, err
	}

	sess, err := smux.Client(res, p.smuxConfig)
	if err != nil {
		return nil, err
	}

	return &session{conn: res, sm: sess, initiator: true}, nil
}

// listener represents a kcp listener
type listener struct {
	*protocol
	listener net.Listener
}

// AcceptSession accepts a session.
func (l *listener) AcceptSession() (base.Session, error) {
	lst := l.listener

	conn, err := lst.Accept()
	if err != nil {
		return nil, err
	}

	sess, err := smux.Server(conn, l.smuxConfig)
	if err != nil {
		return nil, err
	}

	return &session{conn: conn, sm: sess}, nil
}

// Addr gets the address of the listener.
func (l *listener) Addr() net.Addr {
	return l.listener.Addr()
}

// Close closes the listener.
func (l *listener) Close() error {
	return l.listener.Close()
}

// session represents a smux/kcp session
type session struct {
	conn      net.Conn
	sm        *smux.Session
	initiator bool
}

// Initiator returns if this session was built by Dial
func (s *session) Initiator() bool {
	return s.initiator
}

// stream represents a smux stream
type stream struct {
	*smux.Stream
}

// ID returns the identifier for the stream.
func (s *stream) ID() base.StreamID {
	return base.StreamID(s.Stream.ID())
}

// AcceptStream accepts an incoming stream from the connection.
func (s *session) AcceptStream() (base.Stream, error) {
	stm, err := s.sm.AcceptStream()
	if err != nil {
		return nil, err
	}

	return &stream{Stream: stm}, nil
}

// OpenStream opens a new stream on the session.
func (s *session) OpenStream() (base.Stream, error) {
	stm, err := s.sm.OpenStream()
	if err != nil {
		return nil, err
	}
	return &stream{Stream: stm}, nil
}

// LocalAddr is our local address.
func (s *session) LocalAddr() net.Addr {
	return s.conn.LocalAddr()
}

// RemoteAddr is the address of the peer.
func (s *session) RemoteAddr() net.Addr {
	return s.conn.RemoteAddr()
}

// CloseWithError closes the session, transmitting an error if possible.
// Only applicable for Quic.
func (s *session) CloseWithError(err error) error {
	return s.conn.Close()
}

// Close closes the session.
func (s *session) Close() error {
	return s.conn.Close()
}
