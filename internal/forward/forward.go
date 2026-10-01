package forward

import (
	"io"
	"net"
	"sync"
	"time"

	"github.com/SpeedwiT/SpeedTunnel/internal/tunnel"
)

// Forwarder handles port forwarding via multiplexer

type Forwarder struct {
	ln   net.Listener
	mux  *tunnel.Multiplexer
	mu   sync.Mutex
	done chan struct{}
}

func NewForwarder(listenAddr string, mux *tunnel.Multiplexer) (*Forwarder, error) {
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, err
	}
	return &Forwarder{ln: ln, mux: mux, done: make(chan struct{})}, nil
}

func (f *Forwarder) Serve() error {
	for {
		conn, err := f.ln.Accept()
		if err != nil {
			select {
			case <-f.done:
				return nil
			default:
				time.Sleep(100 * time.Millisecond)
				continue
			}
		}
		go f.handleClient(conn)
	}
}

func (f *Forwarder) handleClient(client net.Conn) {
	stream, err := f.mux.OpenStream()
	if err != nil {
		client.Close()
		return
	}
	// pipe client <-> stream (via channels + WriteFrame)
	// client -> stream
	go func() {
		buf := make([]byte, 32*1024)
		for {
			n, err := client.Read(buf)
			if err != nil {
				stream.Close()
				client.Close()
				return
			}
			if err := stream.Write(buf[:n]); err != nil {
				client.Close()
				return
			}
		}
	}()
	// stream -> client
	go func() {
		for data := range stream.RecvCh {
			if _, err := client.Write(data); err != nil {
				stream.Close()
				client.Close()
				return
			}
		}
		client.Close()
	}()
}

func (f *Forwarder) Close() error {
	close(f.done)
	return f.ln.Close()
}

// LocalConnector on IRAN side: when new stream arrives, dial local service
func HandleRemoteStream(s *tunnel.Stream, localAddr string) {
	local, err := net.DialTimeout("tcp", localAddr, 5*time.Second)
	if err != nil {
		s.Close()
		return
	}
	// stream -> local
	go func() {
		for data := range s.RecvCh {
			if _, err := local.Write(data); err != nil {
				local.Close()
				s.Close()
				return
			}
		}
		local.Close()
	}()
	// local -> stream
	go func() {
		buf := make([]byte, 32*1024)
		for {
			n, err := local.Read(buf)
			if err != nil {
				if err != io.EOF {
				}
				s.Close()
				local.Close()
				return
			}
			if err := s.Write(buf[:n]); err != nil {
				local.Close()
				return
			}
		}
	}()
}
