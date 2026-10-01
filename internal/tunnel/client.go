package tunnel

import (
	"log"
	"net"
	"time"

	"github.com/SpeedwiT/SpeedTunnel/internal/transport"
)

type Client struct {
	addr      string
	secret    string
	transport string
	sni       string
	mux       *Multiplexer
	onMux     func(*Multiplexer)
}

func NewClient(addr, secret, trans, sni string, onMux func(*Multiplexer)) *Client {
	return &Client{addr: addr, secret: secret, transport: trans, sni: sni, onMux: onMux}
}

func (c *Client) ConnectWithRetry() {
	backoff := time.Second
	maxBackoff := 30 * time.Second
	for {
		if err := c.connectOnce(); err != nil {
			log.Printf("[client] connect failed: %v retry in %s", err, backoff)
			time.Sleep(backoff)
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			continue
		}
		backoff = time.Second
	}
}

func (c *Client) connectOnce() error {
	var conn net.Conn
	var err error
	timeout := 10 * time.Second
	switch c.transport {
	case "speedtls":
		conn, err = transport.DialSpeedTLS(c.addr, c.sni, c.secret, timeout)
	case "speedhttp":
		conn, err = transport.DialSpeedHTTP(c.addr, c.sni, c.secret, timeout)
	case "speedreverse":
		conn, err = transport.DialSpeedReverse(c.addr, c.secret, timeout)
	default:
		conn, err = transport.DialSpeedTLS(c.addr, c.sni, c.secret, timeout)
	}
	if err != nil {
		return err
	}
	log.Printf("[client] connected to %s via %s", c.addr, c.transport)
	mux := NewMultiplexer(conn)
	c.mux = mux
	if c.onMux != nil {
		c.onMux(mux)
	}
	// keepalive
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := mux.Ping(); err != nil {
				return
			}
		}
	}()
	return mux.Serve()
}

func (c *Client) GetMux() *Multiplexer { return c.mux }
