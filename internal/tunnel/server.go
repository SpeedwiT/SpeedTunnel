package tunnel

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/SpeedwiT/SpeedTunnel/internal/transport"
)

type Server struct {
	addr      string
	secret    string
	transport string
	sni       string
	ln        net.Listener
	onMux     func(*Multiplexer, net.Conn)
}

func NewServer(addr, secret, trans, sni string, onMux func(*Multiplexer, net.Conn)) *Server {
	return &Server{addr: addr, secret: secret, transport: trans, sni: sni, onMux: onMux}
}

func (s *Server) ListenAndServe() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", s.addr, err)
	}
	s.ln = ln
	log.Printf("[server] listening on %s transport=%s", s.addr, s.transport)
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(raw net.Conn) {
	var conn net.Conn
	var err error
	switch s.transport {
	case "speedtls", "speedreverse":
		conn, err = transport.AcceptSpeedTLS(raw, s.secret)
	case "speedhttp":
		conn, err = transport.AcceptSpeedHTTP(raw, s.secret)
	default:
		conn, err = transport.AcceptSpeedTLS(raw, s.secret)
	}
	if err != nil {
		log.Printf("[server] accept fail from %s: %v", raw.RemoteAddr(), err)
		raw.Close()
		return
	}
	log.Printf("[server] new tunnel from %s", raw.RemoteAddr())
	mux := NewMultiplexer(conn)
	// ping keepalive
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := mux.Ping(); err != nil {
				return
			}
		}
	}()
	if s.onMux != nil {
		s.onMux(mux, conn)
	}
	if err := mux.Serve(); err != nil {
		log.Printf("[server] mux closed %s: %v", raw.RemoteAddr(), err)
	}
}

func (s *Server) Close() error {
	if s.ln != nil {
		return s.ln.Close()
	}
	return nil
}
