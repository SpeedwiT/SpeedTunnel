package tunnel

import (
	"encoding/binary"
	"fmt"
	"io"
	"sync"
)

const (
	MsgNewConn  = 0x01 // new forwarded connection
	MsgData     = 0x02 // data frame
	MsgClose    = 0x03 // close stream
	MsgPing     = 0x04
	MsgPong     = 0x05
	MsgAuth     = 0x10
	MsgAuthOK   = 0x11
	MsgAuthFail = 0x12
)

const MaxFrameSize = 1 << 20 // 1MB

type Frame struct {
	Type     byte
	StreamID uint32
	Payload  []byte
}

func WriteFrame(w io.Writer, f Frame) error {
	hdr := make([]byte, 9)
	hdr[0] = f.Type
	binary.BigEndian.PutUint32(hdr[1:5], f.StreamID)
	binary.BigEndian.PutUint32(hdr[5:9], uint32(len(f.Payload)))
	if _, err := w.Write(hdr); err != nil {
		return err
	}
	if len(f.Payload) > 0 {
		if _, err := w.Write(f.Payload); err != nil {
			return err
		}
	}
	return nil
}

func (m *Multiplexer) writeFrame(f Frame) error {
	m.wmu.Lock()
	defer m.wmu.Unlock()
	return WriteFrame(m.conn, f)
}

func ReadFrame(r io.Reader) (Frame, error) {
	hdr := make([]byte, 9)
	if _, err := io.ReadFull(r, hdr); err != nil {
		return Frame{}, err
	}
	typ := hdr[0]
	sid := binary.BigEndian.Uint32(hdr[1:5])
	length := binary.BigEndian.Uint32(hdr[5:9])
	if length > MaxFrameSize {
		return Frame{}, fmt.Errorf("frame too large: %d", length)
	}
	var payload []byte
	if length > 0 {
		payload = make([]byte, length)
		if _, err := io.ReadFull(r, payload); err != nil {
			return Frame{}, err
		}
	}
	return Frame{Type: typ, StreamID: sid, Payload: payload}, nil
}

// Conn wraps a multiplexed stream
type Stream struct {
	ID     uint32
	RecvCh chan []byte
	ctrl   *Multiplexer
	closed bool
	mu     sync.Mutex
}

func (s *Stream) Write(p []byte) error {
	s.mu.Lock()
	closed := s.closed
	s.mu.Unlock()
	if closed {
		return fmt.Errorf("stream closed")
	}
	return s.ctrl.writeFrame(Frame{Type: MsgData, StreamID: s.ID, Payload: p})
}

func (s *Stream) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	s.mu.Unlock()
	return s.ctrl.writeFrame(Frame{Type: MsgClose, StreamID: s.ID})
}

// Multiplexer handles multiple streams over one control connection
type Multiplexer struct {
	conn    io.ReadWriteCloser
	streams map[uint32]*Stream
	mu      sync.RWMutex
	wmu     sync.Mutex
	nextID  uint32
	onNew   func(*Stream)
	onFrame func(Frame)
}

func NewMultiplexer(conn io.ReadWriteCloser) *Multiplexer {
	return &Multiplexer{
		conn:    conn,
		streams: make(map[uint32]*Stream),
		nextID:  1,
	}
}

func (m *Multiplexer) SetOnNewConn(fn func(*Stream)) { m.onNew = fn }

func (m *Multiplexer) OpenStream() (*Stream, error) {
	m.mu.Lock()
	id := m.nextID
	m.nextID += 2
	s := &Stream{ID: id, RecvCh: make(chan []byte, 64), ctrl: m}
	m.streams[id] = s
	m.mu.Unlock()
	if err := m.writeFrame(Frame{Type: MsgNewConn, StreamID: id}); err != nil {
		m.mu.Lock()
		delete(m.streams, id)
		m.mu.Unlock()
		return nil, err
	}
	return s, nil
}

func (m *Multiplexer) loop() error {
	for {
		f, err := ReadFrame(m.conn)
		if err != nil {
			return err
		}
		switch f.Type {
		case MsgPing:
			_ = m.writeFrame(Frame{Type: MsgPong})
		case MsgPong:
			// ignore
		case MsgNewConn:
			m.mu.Lock()
			if _, exists := m.streams[f.StreamID]; !exists {
				s := &Stream{ID: f.StreamID, RecvCh: make(chan []byte, 64), ctrl: m}
				m.streams[f.StreamID] = s
				m.mu.Unlock()
				if m.onNew != nil {
					go m.onNew(s)
				}
			} else {
				m.mu.Unlock()
			}
		case MsgData:
			m.mu.RLock()
			s, ok := m.streams[f.StreamID]
			m.mu.RUnlock()
			if ok {
				select {
				case s.RecvCh <- f.Payload:
				default:
				}
			}
		case MsgClose:
			m.mu.Lock()
			if s, ok := m.streams[f.StreamID]; ok {
				close(s.RecvCh)
				delete(m.streams, f.StreamID)
				s.mu.Lock()
				s.closed = true
				s.mu.Unlock()
			}
			m.mu.Unlock()
		}
		if m.onFrame != nil {
			m.onFrame(f)
		}
	}
}

func (m *Multiplexer) Ping() error { return m.writeFrame(Frame{Type: MsgPing}) }

func (m *Multiplexer) Serve() error { return m.loop() }

func (m *Multiplexer) Close() error { return m.conn.Close() }
