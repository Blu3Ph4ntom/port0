package process

import (
	"net"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/blu3ph4ntom/port0/internal/state"
)

func TestManagerLifecycle(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "port0-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	store := state.NewStore(tempDir)

	exitCh := make(chan string, 1)
	onExit := func(name string, exitCode int, restart string) {
		exitCh <- name
	}

	mgr := NewManager(store, onExit)

	var cmd []string
	if runtime.GOOS == "windows" {
		// Use cmd timeout on Windows
		cmd = []string{"cmd", "/c", "timeout", "10"}
	} else {
		cmd = []string{"sleep", "10"}
	}

	proj := &state.Project{
		Name:    "test-proc",
		Cmd:     cmd,
		Cwd:     tempDir,
		Port:    9001,
		LogFile: filepath.Join(tempDir, "test.log"),
		Restart: "no",
	}

	// Test Spawn
	err = mgr.Spawn(proj)
	if err != nil {
		t.Fatalf("failed to spawn: %v", err)
	}

	if !mgr.IsRunning("test-proc") {
		t.Error("expected process to be running")
	}

	pid := mgr.GetPID("test-proc")
	if pid == 0 {
		t.Error("expected non-zero PID")
	}

	// Verify state persistence
	p, err := store.Get("test-proc")
	if err != nil {
		t.Fatalf("failed to get project from store: %v", err)
	}
	if p.PID != pid {
		t.Errorf("store PID %d != actual PID %d", p.PID, pid)
	}

	// Test Kill
	err = mgr.Kill("test-proc")
	if err != nil {
		t.Fatalf("failed to kill: %v", err)
	}

	select {
	case name := <-exitCh:
		if name != "test-proc" {
			t.Errorf("expected exit name test-proc, got %s", name)
		}
	case <-time.After(5 * time.Second):
		t.Error("timeout waiting for exit callback")
	}

	if mgr.IsRunning("test-proc") {
		t.Error("expected process to be stopped")
	}

	// Verify state updated after exit
	p, _ = store.Get("test-proc")
	if p.PID != 0 {
		t.Errorf("expected store PID to be 0 after exit, got %d", p.PID)
	}
}

func TestManagerKillAll(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "port0-killall-*")
	defer os.RemoveAll(tempDir)
	store := state.NewStore(tempDir)
	mgr := NewManager(store, nil)

	var cmd []string
	if runtime.GOOS == "windows" {
		cmd = []string{"cmd", "/c", "timeout", "10"}
	} else {
		cmd = []string{"sleep", "10"}
	}

	names := []string{"proc1", "proc2", "proc3"}
	for i, name := range names {
		mgr.Spawn(&state.Project{
			Name:    name,
			Cmd:     cmd,
			Cwd:     tempDir,
			Port:    9000 + i,
			LogFile: filepath.Join(tempDir, name+".log"),
		})
	}

	mgr.KillAll()

	for _, name := range names {
		if mgr.IsRunning(name) {
			t.Errorf("expected process %s to be stopped", name)
		}
	}
}

func TestProbe(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port

	if !Probe(port, 2*time.Second) {
		t.Error("Probe failed to detect open port")
	}

	ln.Close()
	// Give the OS a moment to free the port
	time.Sleep(100 * time.Millisecond)

	if Probe(port, 200*time.Millisecond) {
		t.Error("Probe detected closed port as open")
	}
}

func TestPrepareCommand(t *testing.T) {
	cwd, _ := os.Getwd()

	// Test basic command
	cmd := []string{"ls", "-la"}
	got, err := prepareCommand(cmd, cwd)
	if err != nil || len(got) != 2 || got[0] != "ls" {
		t.Errorf("prepareCommand failed for basic cmd: %v", got)
	}

	// Test relative path resolution
	if runtime.GOOS == "windows" {
		cmd = []string{".\\myprog.exe"}
		got, _ = prepareCommand(cmd, cwd)
		abs, _ := filepath.Abs(filepath.Join(cwd, ".\\myprog.exe"))
		if got[0] != abs {
			t.Errorf("expected %s, got %s", abs, got[0])
		}
	} else {
		cmd = []string{"./myprog"}
		got, _ = prepareCommand(cmd, cwd)
		abs, _ := filepath.Abs(filepath.Join(cwd, "./myprog"))
		if got[0] != abs {
			t.Errorf("expected %s, got %s", abs, got[0])
		}
	}
}
