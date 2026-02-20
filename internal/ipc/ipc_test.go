package ipc

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
)

func TestSendAndReadResponse(t *testing.T) {
	dir := t.TempDir()
	sockPath := filepath.Join(dir, "test.sock")

	ln, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	defer os.Remove(sockPath)

	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		req, err := ReadRequest(conn)
		if err != nil {
			t.Errorf("server: read request: %v", err)
			return
		}
		if req.Op != "ls" {
			t.Errorf("server: op = %q, want ls", req.Op)
			return
		}

		projects := []map[string]interface{}{
			{"name": "myapp", "port": 4200},
		}
		WriteOK(conn, projects)
	}()

	client, err := net.Dial("unix", sockPath)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	if err := SendRequest(client, "ls", nil); err != nil {
		t.Fatal(err)
	}

	resp, err := ReadResponse(client)
	if err != nil {
		t.Fatal(err)
	}

	if !resp.OK {
		t.Errorf("response not OK: %s", resp.Error)
	}

	var projects []map[string]interface{}
	json.Unmarshal(resp.Data, &projects)
	if len(projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(projects))
	}

	<-done
}

func TestSendErrorResponse(t *testing.T) {
	dir := t.TempDir()
	sockPath := filepath.Join(dir, "test2.sock")

	ln, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	defer os.Remove(sockPath)

	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		ReadRequest(conn)
		WriteError(conn, "project not found")
	}()

	client, err := net.Dial("unix", sockPath)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	SendRequest(client, "kill", &KillRequest{Name: "nonexistent"})
	resp, err := ReadResponse(client)
	if err != nil {
		t.Fatal(err)
	}

	if resp.OK {
		t.Error("expected error response")
	}
	if resp.Error != "project not found" {
		t.Errorf("error = %q, want %q", resp.Error, "project not found")
	}

	<-done
}
