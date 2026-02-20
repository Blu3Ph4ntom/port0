package process

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/blu3ph4ntom/port0/internal/state"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Child struct {
	Name    string
	Cmd     *exec.Cmd
	LogFile *lumberjack.Logger
	Done    chan struct{}
	Err     error
}

type OnExitFunc func(name string, exitCode int, restart string)

type Manager struct {
	mu       sync.Mutex
	children map[string]*Child
	store    *state.Store
	onExit   OnExitFunc
}

func NewManager(store *state.Store, onExit OnExitFunc) *Manager {
	return &Manager{
		children: make(map[string]*Child),
		store:    store,
		onExit:   onExit,
	}
}

func (m *Manager) Spawn(proj *state.Project) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if existing, ok := m.children[proj.Name]; ok {
		if existing.Cmd.Process != nil {
			select {
			case <-existing.Done:
			default:
				return fmt.Errorf("process: %s is already running (pid %d)", proj.Name, existing.Cmd.Process.Pid)
			}
		}
	}

	if len(proj.Cmd) == 0 {
		return fmt.Errorf("process: no command specified for %s", proj.Name)
	}

	// Resolve relative paths on Windows
	cmdArgs := make([]string, len(proj.Cmd))
	copy(cmdArgs, proj.Cmd)
	for i, arg := range cmdArgs {
		// If arg looks like a relative path (.\file or ..\file on Windows)
		if len(arg) > 2 && (arg[0] == '.' && (arg[1] == '\\' || arg[1] == '/')) {
			// Make it absolute
			if abs, err := filepath.Abs(filepath.Join(proj.Cwd, arg)); err == nil {
				cmdArgs[i] = abs
			}
		}
	}

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Dir = proj.Cwd

	env := os.Environ()
	env = append(env, fmt.Sprintf("PORT=%d", proj.Port))
	cmd.Env = env

	logger := &lumberjack.Logger{
		Filename:   proj.LogFile,
		MaxSize:    10,
		MaxBackups: 3,
	}

	cmd.Stdout = io.MultiWriter(os.Stdout, logger)
	cmd.Stderr = io.MultiWriter(os.Stderr, logger)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("process: start %s: %w", proj.Name, err)
	}

	proj.PID = cmd.Process.Pid
	proj.StartedAt = time.Now()
	if err := m.store.Set(proj); err != nil {
		slog.Error("process: save state", "name", proj.Name, "err", err)
	}

	child := &Child{
		Name:    proj.Name,
		Cmd:     cmd,
		LogFile: logger,
		Done:    make(chan struct{}),
	}
	m.children[proj.Name] = child

	go m.waitForExit(child, proj.Restart)

	return nil
}

func (m *Manager) waitForExit(child *Child, restart string) {
	err := child.Cmd.Wait()
	child.Err = err
	close(child.Done)

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	slog.Info("process exited", "name", child.Name, "exit_code", exitCode)

	if p, err := m.store.Get(child.Name); err == nil {
		p.PID = 0
		m.store.Set(p)
	}

	if m.onExit != nil {
		m.onExit(child.Name, exitCode, restart)
	}
}

func (m *Manager) Kill(name string) error {
	m.mu.Lock()
	child, ok := m.children[name]
	m.mu.Unlock()

	if !ok {
		return fmt.Errorf("process: %s not found", name)
	}

	select {
	case <-child.Done:
		return fmt.Errorf("process: %s is not running", name)
	default:
	}

	if child.Cmd.Process == nil {
		return fmt.Errorf("process: %s has no process", name)
	}

	child.Cmd.Process.Signal(os.Interrupt)

	select {
	case <-child.Done:
		return nil
	case <-time.After(5 * time.Second):
		child.Cmd.Process.Kill()
		<-child.Done
		return nil
	}
}

func (m *Manager) KillAll() {
	m.mu.Lock()
	names := make([]string, 0, len(m.children))
	for name := range m.children {
		names = append(names, name)
	}
	m.mu.Unlock()

	for _, name := range names {
		m.Kill(name)
	}
}

func (m *Manager) IsRunning(name string) bool {
	m.mu.Lock()
	child, ok := m.children[name]
	m.mu.Unlock()

	if !ok {
		return false
	}

	select {
	case <-child.Done:
		return false
	default:
		return true
	}
}

func (m *Manager) GetPID(name string) int {
	m.mu.Lock()
	child, ok := m.children[name]
	m.mu.Unlock()

	if !ok {
		return 0
	}
	if child.Cmd.Process == nil {
		return 0
	}
	return child.Cmd.Process.Pid
}

func Probe(port int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 500*time.Millisecond)
		if err == nil {
			conn.Close()
			return true
		}
		time.Sleep(500 * time.Millisecond)
	}
	return false
}
