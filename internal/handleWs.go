package internal

import (
	"bufio"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
)

const (
	opcodeContinuation = 0x0
	opcodeText         = 0x1
	opcodeBinary       = 0x2
	opcodeClose        = 0x8
	opcodePing         = 0x9
	opcodePong         = 0xA
)

type WSConn struct {
	conn net.Conn
	r    *bufio.Reader
}

func ConnectWS(serverAddr string, validateToken string, isLocalhost bool) (*WSConn, error) {
	hostOnly := serverAddr
	if strings.Contains(serverAddr, ":") {
		hostOnly, _, _ = net.SplitHostPort(serverAddr)
	}

	var (
		conn net.Conn
		err  error
	)
	if isLocalhost {
		conn, err = net.Dial("tcp", serverAddr)
	} else {
		config := &tls.Config{ServerName: hostOnly, InsecureSkipVerify: isLocalhost}
		conn, err = tls.Dial("tcp", serverAddr, config)
	}
	if err != nil {
		return nil, err
	}

	key := make([]byte, 16)
	rand.Reader.Read(key)
	strKey := base64.StdEncoding.EncodeToString(key)
	path := fmt.Sprintf("/ws?token=%s", validateToken)
	if _, err := fmt.Fprintf(conn,
		"GET %s HTTP/1.1\r\nHost: %s\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Key: %s\r\nSec-WebSocket-Version: 13\r\nValidate: %s\r\n\r\n",
		path, hostOnly, strKey, validateToken,
	); err != nil {
		conn.Close()
		return nil, err
	}

	r := bufio.NewReader(conn)
	statusLine, err := r.ReadString('\n')
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("error reading response: %w", err)
	}
	if !strings.Contains(statusLine, "HTTP/1.1 101 Switching Protocols") {
		conn.Close()
		return nil, fmt.Errorf("server rejected upgrade: %s", strings.TrimSpace(statusLine))
	}
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("error reading handshake headers: %w", err)
		}
		if line == "\r\n" || line == "\n" {
			break
		}
	}
	fmt.Println("Connect to Websocket")
	return &WSConn{conn: conn, r: r}, nil
}

func (w *WSConn) Close() error {
	if w == nil || w.conn == nil {
		return nil
	}
	return w.conn.Close()
}

func (w *WSConn) Write(frame []byte) error {
	if w == nil || w.conn == nil {
		return fmt.Errorf("ws: not connected")
	}
	_, err := w.conn.Write(frame)
	return err
}

func (w *WSConn) Read() error {
	if w == nil || w.conn == nil {
		return fmt.Errorf("ws: not connected")
	}
	first, err := w.r.ReadByte()
	if err != nil {
		return err
	}
	opcode := first & 0x0F
	second, err := w.r.ReadByte()
	if err != nil {
		return err
	}
	masked := second&0x80 != 0
	length := uint64(second & 0x7F)
	switch length {
	case 126:
		var b [2]byte
		if _, err := io.ReadFull(w.r, b[:]); err != nil {
			return err
		}
		length = uint64(binary.BigEndian.Uint16(b[:]))
	case 127:
		var b [8]byte
		if _, err := io.ReadFull(w.r, b[:]); err != nil {
			return err
		}
		length = binary.BigEndian.Uint64(b[:])
	}
	if length > 1<<20 {
		return fmt.Errorf("ws: frame too large: %d", length)
	}
	var mask [4]byte
	if masked {
		if _, err := io.ReadFull(w.r, mask[:]); err != nil {
			return err
		}
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(w.r, payload); err != nil {
		return err
	}
	if masked {
		for i := range payload {
			payload[i] ^= mask[i%4]
		}
	}
	switch opcode {
	case opcodePing:
		return w.writeControl(opcodePong, payload)
	case opcodeClose:
		return fmt.Errorf("ws: connection closed by server")
	default:
		return nil
	}
}

func (w *WSConn) writeControl(opcode byte, payload []byte) error {
	if w == nil || w.conn == nil {
		return fmt.Errorf("ws: not connected")
	}
	if len(payload) > 125 {
		return fmt.Errorf("ws: control frame too large")
	}
	frame := make([]byte, 2, 2+len(payload))
	frame[0] = 0x80 | opcode
	frame[1] = byte(len(payload))
	frame = append(frame, payload...)
	_, err := w.conn.Write(frame)
	return err
}
