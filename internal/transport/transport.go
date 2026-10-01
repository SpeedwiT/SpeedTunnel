package transport

import (
	"bufio"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"math/big"
	"net"
	"strings"
	"time"
)

// Obfuscate wraps conn with optional padding/fragmentation

// FragmentedConn fragments writes into random chunks to evade DPI
type FragmentedConn struct {
	net.Conn
	minChunk int
	maxChunk int
	delayMin time.Duration
	delayMax time.Duration
}

func NewFragmentedConn(c net.Conn) *FragmentedConn {
	return &FragmentedConn{Conn: c, minChunk: 400, maxChunk: 1200, delayMin: 0, delayMax: 5 * time.Millisecond}
}

func (f *FragmentedConn) Write(p []byte) (int, error) {
	total := len(p)
	offset := 0
	for offset < total {
		chunkSize := f.minChunk
		if f.maxChunk > f.minChunk {
			n, _ := rand.Int(rand.Reader, big.NewInt(int64(f.maxChunk-f.minChunk)))
			chunkSize += int(n.Int64())
		}
		if chunkSize > total-offset {
			chunkSize = total - offset
		}
		if _, err := f.Conn.Write(p[offset : offset+chunkSize]); err != nil {
			return offset, err
		}
		offset += chunkSize
		if offset < total && f.delayMax > 0 {
			d := f.delayMin
			if f.delayMax > f.delayMin {
				n, _ := rand.Int(rand.Reader, big.NewInt(int64(f.delayMax-f.delayMin)))
				d += time.Duration(n.Int64())
			}
			time.Sleep(d)
		}
	}
	return total, nil
}

// PaddingConn adds random padding length prefix (for obfuscation layer)
type PaddingConn struct {
	net.Conn
	key []byte
}

func NewPaddingConn(c net.Conn, key []byte) *PaddingConn {
	return &PaddingConn{Conn: c, key: key}
}

// Dial helpers for each transport

func DialSpeedTLS(addr, sni, secret string, timeout time.Duration) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, err
	}
	// Send fake TLS ClientHello with SNI spoof
	if err := sendFakeTLSHello(conn, sni); err != nil {
		conn.Close()
		return nil, fmt.Errorf("fake hello: %w", err)
	}
	// Send auth header
	if err := sendAuth(conn, secret); err != nil {
		conn.Close()
		return nil, err
	}
	// Expect AuthOK (simple 3 byte response)
	if err := expectAuthOK(conn); err != nil {
		conn.Close()
		return nil, err
	}
	return NewFragmentedConn(conn), nil
}

func DialSpeedHTTP(addr, host, secret string, timeout time.Duration) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, err
	}
	// Fake HTTP/2-like upgrade request with Host spoof
	fakeReq := fmt.Sprintf("GET / HTTP/1.1\r\nHost: %s\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\nSec-WebSocket-Version: 13\r\nUser-Agent: Mozilla/5.0\r\n\r\n", host)
	if _, err := conn.Write([]byte(fakeReq)); err != nil {
		conn.Close()
		return nil, err
	}
	// Read fake 101 response (optional, not strict)
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	br := bufio.NewReader(conn)
	line, _ := br.ReadString('\n')
	_ = line
	_ = conn.SetReadDeadline(time.Time{})
	if err := sendAuth(conn, secret); err != nil {
		conn.Close()
		return nil, err
	}
	if err := expectAuthOK(conn); err != nil {
		conn.Close()
		return nil, err
	}
	return NewFragmentedConn(conn), nil
}

func DialSpeedReverse(addr, secret string, timeout time.Duration) (net.Conn, error) {
	// Same as TLS but role reversed – caller is IRAN connecting outward
	return DialSpeedTLS(addr, "www.cloudflare.com", secret, timeout)
}

// Server side accept helpers

func AcceptSpeedTLS(conn net.Conn, expectedSecret string) (net.Conn, error) {
	// Consume fake TLS hello using buffered reader so we don't lose auth bytes
	_ = conn.SetReadDeadline(time.Now().Add(4 * time.Second))
	br := bufio.NewReader(conn)
	peek, err := br.Peek(5)
	if err == nil && len(peek) >= 5 && peek[0] == 0x16 && peek[1] == 0x03 {
		// TLS record: 5 byte header + payload
		recLen := int(binary.BigEndian.Uint16(peek[3:5]))
		need := 5 + recLen
		tmp := make([]byte, need)
		if _, err := io.ReadFull(br, tmp); err != nil {
			// fallback: continue
		}
	}
	_ = conn.SetReadDeadline(time.Time{})
	// Auth may already be buffered in br, so use bufferedConn wrapper
	wrapped := &bufferedConn{Conn: conn, br: br}
	if err := verifyAuth(wrapped, expectedSecret); err != nil {
		return nil, err
	}
	if _, err := conn.Write([]byte{0x11, 0x00, 0x00}); err != nil {
		return nil, err
	}
	// Return conn that drains buffered reader first, then raw conn, with fragmentation
	return &bufferedFragmentedConn{bufferedConn: bufferedConn{Conn: conn, br: br}, frag: NewFragmentedConn(conn)}, nil
}

// bufferedFragmentedConn combines buffered read with fragmented write
type bufferedFragmentedConn struct {
	bufferedConn
	frag *FragmentedConn
}

func (b *bufferedFragmentedConn) Write(p []byte) (int, error) { return b.frag.Write(p) }

func AcceptSpeedHTTP(conn net.Conn, expectedSecret string) (net.Conn, error) {
	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	br := bufio.NewReader(conn)
	// consume HTTP headers until empty line
	for {
		line, err := br.ReadString('\n')
		if err != nil || line == "\r\n" || line == "\n" || strings.TrimSpace(line) == "" {
			break
		}
	}
	// send fake 101
	_, _ = conn.Write([]byte("HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n\r\n"))
	_ = conn.SetReadDeadline(time.Time{})
	// Buffered reader may have extra bytes; need to handle auth that may already be buffered
	// For simplicity we read auth via raw conn – but buffered data might contain it.
	// So we create a wrapper that first drains buffered reader.
	wrapped := &bufferedConn{Conn: conn, br: br}
	if err := verifyAuth(wrapped, expectedSecret); err != nil {
		return nil, err
	}
	if _, err := conn.Write([]byte{0x11, 0x00, 0x00}); err != nil {
		return nil, err
	}
	return NewFragmentedConn(conn), nil
}

type bufferedConn struct {
	net.Conn
	br *bufio.Reader
}

func (b *bufferedConn) Read(p []byte) (int, error) {
	if b.br.Buffered() > 0 {
		return b.br.Read(p)
	}
	return b.Conn.Read(p)
}

// Auth helpers

func sendAuth(conn net.Conn, secret string) error {
	// Simple: 1 byte type 0x10 + 4 byte len + secret bytes
	data := []byte(secret)
	hdr := make([]byte, 5)
	hdr[0] = 0x10
	binary.BigEndian.PutUint32(hdr[1:5], uint32(len(data)))
	if _, err := conn.Write(hdr); err != nil {
		return err
	}
	_, err := conn.Write(data)
	return err
}

func verifyAuth(conn net.Conn, expected string) error {
	hdr := make([]byte, 5)
	if _, err := io.ReadFull(conn, hdr); err != nil {
		return fmt.Errorf("read auth header: %w", err)
	}
	if hdr[0] != 0x10 {
		return fmt.Errorf("invalid auth type %x", hdr[0])
	}
	length := binary.BigEndian.Uint32(hdr[1:5])
	if length > 256 {
		return fmt.Errorf("auth too long")
	}
	buf := make([]byte, length)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return err
	}
	if string(buf) != expected {
		_, _ = conn.Write([]byte{0x12, 0x00, 0x00})
		return fmt.Errorf("auth failed")
	}
	return nil
}

func expectAuthOK(conn net.Conn) error {
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	defer conn.SetReadDeadline(time.Time{})
	hdr := make([]byte, 3)
	if _, err := io.ReadFull(conn, hdr); err != nil {
		return fmt.Errorf("read auth response: %w", err)
	}
	if hdr[0] != 0x11 {
		return fmt.Errorf("auth rejected by server")
	}
	return nil
}

func sendFakeTLSHello(conn net.Conn, sni string) error {
	// Minimal fake TLS ClientHello: record header 0x16 0x03 0x03 + length + payload with SNI
	// This is NOT a valid TLS, just enough to look like TLS to DPI
	sniBytes := []byte(sni)
	payload := make([]byte, 0, 100+len(sniBytes))
	payload = append(payload, 0x01, 0x00) // handshake type ClientHello + length placeholder
	payload = append(payload, 0x03, 0x03) // version TLS 1.2
	// 32 bytes random
	rb := make([]byte, 32)
	_, _ = rand.Read(rb)
	payload = append(payload, rb...)
	payload = append(payload, 0x00) // session id len
	// cipher suites
	payload = append(payload, 0x00, 0x04, 0x13, 0x01, 0x13, 0x02)
	payload = append(payload, 0x01, 0x00) // compression
	// extensions: SNI
	extLen := 2 + 2 + 2 + 1 + 2 + len(sniBytes)
	payload = append(payload, byte(extLen>>8), byte(extLen))
	payload = append(payload, 0x00, 0x00) // SNI extension type
	sniExtLen := 2 + 1 + 2 + len(sniBytes)
	payload = append(payload, byte(sniExtLen>>8), byte(sniExtLen))
	payload = append(payload, byte((len(sniBytes)+3)>>8), byte(len(sniBytes)+3))
	payload = append(payload, 0x00)
	payload = append(payload, byte(len(sniBytes)>>8), byte(len(sniBytes)))
	payload = append(payload, sniBytes...)
	// fix handshake length
	hsLen := len(payload) - 4
	payload[1] = byte(hsLen >> 16)
	payload[2] = byte(hsLen >> 8)
	payload[3] = byte(hsLen)
	// record layer
	rec := make([]byte, 5+len(payload))
	rec[0] = 0x16
	rec[1] = 0x03
	rec[2] = 0x03
	binary.BigEndian.PutUint16(rec[3:5], uint16(len(payload)))
	copy(rec[5:], payload)
	_, err := conn.Write(rec)
	return err
}
