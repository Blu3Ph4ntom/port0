package ipc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
)

type Request struct {
	Op      string          `json:"op"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type Response struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data,omitempty"`
	Error string          `json:"error,omitempty"`
}

type SpawnRequest struct {
	Name      string   `json:"name"`
	Cmd       []string `json:"cmd"`
	Cwd       string   `json:"cwd"`
	Restart   string   `json:"restart"`
	TLS       bool     `json:"tls"`
	PortRange string   `json:"port_range"`
}

type KillRequest struct {
	Name   string `json:"name"`
	Remove bool   `json:"remove"`
}

type LogsRequest struct {
	Name   string `json:"name"`
	Follow bool   `json:"follow"`
}

type LinkRequest struct {
	Name string `json:"name"`
	Port int    `json:"port"`
	Cwd  string `json:"cwd"`
}

type OpenRequest struct {
	Name string `json:"name"`
}

type RegisterRequest struct {
	Name      string   `json:"name"`
	Cmd       []string `json:"cmd"`
	Cwd       string   `json:"cwd"`
	PortRange string   `json:"port_range"`
}

type LogLine struct {
	Line string `json:"line"`
	TS   string `json:"ts"`
}

func SocketPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".port0", "daemon.sock")
}

func PidPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".port0", "daemon.pid")
}

func Connect() (net.Conn, error) {
	sockPath := SocketPath()
	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		return nil, fmt.Errorf("ipc: connect: %w", err)
	}
	return conn, nil
}

func SendRequest(conn net.Conn, op string, payload interface{}) error {
	var raw json.RawMessage
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("ipc: marshal payload: %w", err)
		}
		raw = b
	}

	req := Request{Op: op, Payload: raw}
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("ipc: marshal request: %w", err)
	}
	data = append(data, '\n')
	_, err = conn.Write(data)
	if err != nil {
		return fmt.Errorf("ipc: write: %w", err)
	}
	return nil
}

func ReadResponse(conn net.Conn) (*Response, error) {
	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("ipc: read: %w", err)
		}
		return nil, fmt.Errorf("ipc: connection closed")
	}

	var resp Response
	if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
		return nil, fmt.Errorf("ipc: unmarshal response: %w", err)
	}
	return &resp, nil
}

func StreamLines(conn net.Conn, fn func(line, ts string) bool) error {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var ll LogLine
		if err := json.Unmarshal(scanner.Bytes(), &ll); err != nil {
			continue
		}
		if !fn(ll.Line, ll.TS) {
			return nil
		}
	}
	return scanner.Err()
}

func WriteResponse(conn net.Conn, resp *Response) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("ipc: marshal response: %w", err)
	}
	data = append(data, '\n')
	_, err = conn.Write(data)
	return err
}

func WriteOK(conn net.Conn, data interface{}) error {
	var raw json.RawMessage
	if data != nil {
		b, err := json.Marshal(data)
		if err != nil {
			return err
		}
		raw = b
	}
	return WriteResponse(conn, &Response{OK: true, Data: raw})
}

func WriteError(conn net.Conn, msg string) error {
	return WriteResponse(conn, &Response{OK: false, Error: msg})
}

func WriteLogLine(conn net.Conn, line, ts string) error {
	ll := LogLine{Line: line, TS: ts}
	data, err := json.Marshal(ll)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = conn.Write(data)
	return err
}

func ReadRequest(conn net.Conn) (*Request, error) {
	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("ipc: read request: %w", err)
		}
		return nil, fmt.Errorf("ipc: connection closed")
	}
	var req Request
	if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
		return nil, fmt.Errorf("ipc: unmarshal request: %w", err)
	}
	return &req, nil
}
