package daemon

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/blu3ph4ntom/port0/internal/allocator"
	"github.com/blu3ph4ntom/port0/internal/cert"
	pdns "github.com/blu3ph4ntom/port0/internal/dns"
	"github.com/blu3ph4ntom/port0/internal/ipc"
	"github.com/blu3ph4ntom/port0/internal/process"
	"github.com/blu3ph4ntom/port0/internal/proxy"
	"github.com/blu3ph4ntom/port0/internal/state"
	"github.com/blu3ph4ntom/port0/internal/util"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Daemon struct {
	store    *state.Store
	proxy    *proxy.Proxy
	dns      *pdns.Server
	manager  *process.Manager
	listener net.Listener
	logger   *slog.Logger
}

func Run() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("daemon: get home dir: %w", err)
	}
	base := filepath.Join(home, ".port0")
	os.MkdirAll(base, 0755)
	os.MkdirAll(filepath.Join(base, "logs"), 0755)
	os.MkdirAll(filepath.Join(base, "certs"), 0755)

	logWriter := &lumberjack.Logger{
		Filename:   filepath.Join(base, "daemon.log"),
		MaxSize:    10,
		MaxBackups: 3,
	}
	logger := slog.New(slog.NewJSONHandler(logWriter, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	store := state.DefaultStore()
	p := proxy.New()
	dnsServer := pdns.New()

	d := &Daemon{
		store:  store,
		proxy:  p,
		dns:    dnsServer,
		logger: logger,
	}

	d.manager = process.NewManager(store, d.onProcessExit)

	// Reconcile stale PIDs from previous daemon runs (e.g. crash/kill).
	// We treat "port not listening" as stale and clear PID to avoid lockouts.
	d.reconcileState()
	d.syncState()

	if err := p.StartHTTP(":80"); err != nil {
		logger.Error("failed to start HTTP proxy", "err", err)
		return fmt.Errorf("daemon: start http proxy: %w", err)
	}
	logger.Info("HTTP proxy listening", "addr", ":80")

	dnsAddr, err := dnsServer.StartWithFallback("127.0.0.1:53", "127.0.0.1:5353")
	if err != nil {
		logger.Warn("DNS server failed to start", "err", err)
	} else {
		logger.Info("DNS server listening", "addr", dnsAddr)
	}

	sockPath := ipc.SocketPath()
	os.Remove(sockPath)
	ln, err := net.Listen("unix", sockPath)
	if err != nil {
		return fmt.Errorf("daemon: listen unix: %w", err)
	}
	d.listener = ln
	logger.Info("IPC socket listening", "path", sockPath)

	pidPath := ipc.PidPath()
	os.WriteFile(pidPath, []byte(fmt.Sprintf("%d", os.Getpid())), 0644)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	go d.acceptLoop()

	sig := <-sigCh
	logger.Info("received signal, shutting down", "signal", sig)
	d.shutdown()
	return nil
}

func (d *Daemon) acceptLoop() {
	for {
		conn, err := d.listener.Accept()
		if err != nil {
			return
		}
		go d.handleConn(conn)
	}
}

func (d *Daemon) handleConn(conn net.Conn) {
	defer conn.Close()

	req, err := ipc.ReadRequest(conn)
	if err != nil {
		d.logger.Error("ipc: read request", "err", err)
		return
	}

	d.logger.Info("ipc request", "op", req.Op)

	switch req.Op {
	case "spawn":
		d.handleSpawn(conn, req.Payload)
	case "register":
		d.handleRegister(conn, req.Payload)
	case "unregister":
		d.handleUnregister(conn, req.Payload)
	case "kill":
		d.handleKill(conn, req.Payload)
	case "ls":
		d.handleLs(conn)
	case "logs":
		d.handleLogs(conn, req.Payload)
	case "link":
		d.handleLink(conn, req.Payload)
	case "open":
		d.handleOpen(conn, req.Payload)
	case "status":
		d.handleStatus(conn)
	default:
		ipc.WriteError(conn, fmt.Sprintf("unknown op: %s", req.Op))
	}
}

func (d *Daemon) handleSpawn(conn net.Conn, payload json.RawMessage) {
	var req ipc.SpawnRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		ipc.WriteError(conn, fmt.Sprintf("invalid spawn request: %v", err))
		return
	}

	name := req.Name
	if name == "" {
		name = util.FromCwd(req.Cwd)
	}

	projects, _ := d.store.All()
	takenNames := make(map[string]bool)
	takenPorts := make(map[int]bool)
	for n, p := range projects {
		takenNames[n] = true
		takenPorts[p.Port] = true
	}

	if existing, ok := projects[name]; ok {
		if d.manager.IsRunning(name) {
			ipc.WriteError(conn, fmt.Sprintf("%s is already running at http://%s.localhost (pid %d)", name, name, existing.PID))
			return
		}
		if existing.Cwd != req.Cwd {
			name = util.Deconflict(name, takenNames)
			d.logger.Warn("name collision, deconflicted", "original", req.Name, "resolved", name)
		}
	}

	portRange := "4000-4999"
	if req.PortRange != "" {
		portRange = req.PortRange
	}
	alloc, err := allocator.ParseRange(portRange)
	if err != nil {
		ipc.WriteError(conn, fmt.Sprintf("invalid port range: %v", err))
		return
	}

	port, err := alloc.Pick(takenPorts)
	if err != nil {
		ipc.WriteError(conn, fmt.Sprintf("port allocation failed: %v", err))
		return
	}

	home, err := os.UserHomeDir()
	if err != nil {
		ipc.WriteError(conn, fmt.Sprintf("failed to get home directory: %v", err))
		return
	}
	proj := &state.Project{
		Name:    name,
		Port:    port,
		Cmd:     req.Cmd,
		Cwd:     req.Cwd,
		Restart: req.Restart,
		LogFile: filepath.Join(home, ".port0", "logs", name+".log"),
		Domain:  req.Domain,
	}

	if req.TLS {
		if err := cert.EnsureGenerated(name); err != nil {
			ipc.WriteError(conn, fmt.Sprintf("cert generation failed: %v", err))
			return
		}
	}

	if err := d.manager.Spawn(proj); err != nil {
		ipc.WriteError(conn, fmt.Sprintf("spawn failed: %v", err))
		return
	}

	d.syncState()

	go func() {
		if !process.Probe(port, 10*time.Second) {
			d.logger.Warn("process may not honor PORT env var", "name", name, "port", port)
		}
	}()

	// Build URLs - if domain is set, use subdomain format
	baseHost := name
	if req.Domain != "" {
		baseHost = name + "." + req.Domain
	}

	result := map[string]interface{}{
		"name":      name,
		"port":      port,
		"url":       fmt.Sprintf("http://%s.localhost", baseHost),
		"url_web":   fmt.Sprintf("http://%s.web", baseHost),
		"url_local": fmt.Sprintf("http://%s.local", baseHost),
		"cmd":       req.Cmd,
		"pid":       proj.PID,
	}
	if req.Domain != "" {
		result["domain"] = req.Domain
	}
	if req.TLS {
		result["url"] = fmt.Sprintf("https://%s.localhost", baseHost)
		result["url_web"] = fmt.Sprintf("https://%s.web", baseHost)
		result["url_local"] = fmt.Sprintf("https://%s.local", baseHost)
		result["tls"] = true
	}

	ipc.WriteOK(conn, result)
}

func (d *Daemon) handleKill(conn net.Conn, payload json.RawMessage) {
	var req ipc.KillRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		ipc.WriteError(conn, fmt.Sprintf("invalid kill request: %v", err))
		return
	}

	pid := d.manager.GetPID(req.Name)

	// Try to kill the process, but don't fail if it's already stopped
	// when we're just trying to remove it
	err := d.manager.Kill(req.Name)
	if err != nil && !req.Remove {
		ipc.WriteError(conn, err.Error())
		return
	}

	if req.Remove {
		d.store.Delete(req.Name)
		d.syncState()
		ipc.WriteOK(conn, map[string]interface{}{
			"name":    req.Name,
			"removed": true,
		})
		return
	}

	d.syncState()
	ipc.WriteOK(conn, map[string]interface{}{
		"name": req.Name,
		"pid":  pid,
	})
}

func (d *Daemon) handleLs(conn net.Conn) {
	projects, err := d.store.All()
	if err != nil {
		ipc.WriteError(conn, fmt.Sprintf("failed to list projects: %v", err))
		return
	}

	list := make([]map[string]interface{}, 0, len(projects))
	for _, p := range projects {
		status := "stopped"
		if d.manager.IsRunning(p.Name) {
			status = "running"
		}
		list = append(list, map[string]interface{}{
			"name":       p.Name,
			"port":       p.Port,
			"url":        fmt.Sprintf("http://%s.localhost", p.Name),
			"url_web":    fmt.Sprintf("http://%s.web", p.Name),
			"url_local":  fmt.Sprintf("http://%s.local", p.Name),
			"pid":        p.PID,
			"status":     status,
			"started_at": p.StartedAt,
			"cmd":        p.Cmd,
		})
	}

	ipc.WriteOK(conn, list)
}

func (d *Daemon) handleLogs(conn net.Conn, payload json.RawMessage) {
	var req ipc.LogsRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		ipc.WriteError(conn, fmt.Sprintf("invalid logs request: %v", err))
		return
	}

	proj, err := d.store.Get(req.Name)
	if err != nil {
		ipc.WriteError(conn, err.Error())
		return
	}

	if !req.Follow {
		data, err := os.ReadFile(proj.LogFile)
		if err != nil {
			if os.IsNotExist(err) {
				ipc.WriteOK(conn, map[string]interface{}{"lines": []string{}})
				return
			}
			ipc.WriteError(conn, fmt.Sprintf("read logs: %v", err))
			return
		}
		lines := tailLines(string(data), 100)
		ipc.WriteOK(conn, map[string]interface{}{"lines": lines})
		return
	}

	f, err := os.Open(proj.LogFile)
	if err != nil {
		ipc.WriteError(conn, fmt.Sprintf("open log file: %v", err))
		return
	}
	defer f.Close()

	f.Seek(0, io.SeekEnd)
	reader := bufio.NewReader(f)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		ts := time.Now().Format(time.RFC3339)
		if writeErr := ipc.WriteLogLine(conn, line, ts); writeErr != nil {
			return
		}
	}
}

func (d *Daemon) handleLink(conn net.Conn, payload json.RawMessage) {
	var req ipc.LinkRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		ipc.WriteError(conn, fmt.Sprintf("invalid link request: %v", err))
		return
	}

	home, err := os.UserHomeDir()
	if err != nil {
		ipc.WriteError(conn, fmt.Sprintf("failed to get home directory: %v", err))
		return
	}
	proj := &state.Project{
		Name:      req.Name,
		Port:      req.Port,
		Cwd:       req.Cwd,
		Restart:   "no",
		StartedAt: time.Now(),
		LogFile:   filepath.Join(home, ".port0", "logs", req.Name+".log"),
	}

	if err := d.store.Set(proj); err != nil {
		ipc.WriteError(conn, fmt.Sprintf("link failed: %v", err))
		return
	}
	d.syncState()

	ipc.WriteOK(conn, map[string]interface{}{
		"name":      req.Name,
		"port":      req.Port,
		"url":       fmt.Sprintf("http://%s.localhost", req.Name),
		"url_web":   fmt.Sprintf("http://%s.web", req.Name),
		"url_local": fmt.Sprintf("http://%s.local", req.Name),
	})
}

func (d *Daemon) handleOpen(conn net.Conn, payload json.RawMessage) {
	var req ipc.OpenRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		ipc.WriteError(conn, fmt.Sprintf("invalid open request: %v", err))
		return
	}

	proj, err := d.store.Get(req.Name)
	if err != nil {
		ipc.WriteError(conn, err.Error())
		return
	}

	url := fmt.Sprintf("http://%s.localhost", proj.Name)
	ipc.WriteOK(conn, map[string]interface{}{
		"name":      proj.Name,
		"url":       url,
		"url_web":   fmt.Sprintf("http://%s.web", proj.Name),
		"url_local": fmt.Sprintf("http://%s.local", proj.Name),
	})
}

func (d *Daemon) handleRegister(conn net.Conn, payload json.RawMessage) {
	var req ipc.RegisterRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		ipc.WriteError(conn, fmt.Sprintf("invalid register request: %v", err))
		return
	}

	name := req.Name
	if name == "" {
		name = util.FromCwd(req.Cwd)
	}

	projects, _ := d.store.All()
	takenNames := make(map[string]bool)
	takenPorts := make(map[int]bool)
	for n, p := range projects {
		takenNames[n] = true
		takenPorts[p.Port] = true
	}

	if existing, ok := projects[name]; ok {
		if existing.PID != 0 {
			// If the port isn't reachable, treat the PID as stale (common after daemon restart).
			if !process.Probe(existing.Port, 500*time.Millisecond) {
				d.logger.Warn("clearing stale pid", "name", name, "pid", existing.PID, "port", existing.Port)
				existing.PID = 0
				_ = d.store.Set(existing)
			} else {
				ipc.WriteError(conn, fmt.Sprintf("%s is already running at http://%s.localhost (pid %d)", name, name, existing.PID))
				return
			}
		}
		if existing.Cwd != req.Cwd {
			name = util.Deconflict(name, takenNames)
			d.logger.Warn("name collision, deconflicted", "original", req.Name, "resolved", name)
		}
	}

	portRange := "4000-4999"
	if req.PortRange != "" {
		portRange = req.PortRange
	}
	alloc, err := allocator.ParseRange(portRange)
	if err != nil {
		ipc.WriteError(conn, fmt.Sprintf("invalid port range: %v", err))
		return
	}

	port, err := alloc.Pick(takenPorts)
	if err != nil {
		ipc.WriteError(conn, fmt.Sprintf("port allocation failed: %v", err))
		return
	}

	home, err := os.UserHomeDir()
	if err != nil {
		ipc.WriteError(conn, fmt.Sprintf("failed to get home directory: %v", err))
		return
	}
	proj := &state.Project{
		Name:      name,
		Port:      port,
		Cmd:       req.Cmd,
		Cwd:       req.Cwd,
		PID:       0, // No PID - foreground mode
		Restart:   "no",
		StartedAt: time.Now(),
		LogFile:   filepath.Join(home, ".port0", "logs", name+".log"),
		Domain:    req.Domain,
	}

	if err := d.store.Set(proj); err != nil {
		ipc.WriteError(conn, fmt.Sprintf("register failed: %v", err))
		return
	}
	d.syncState()

	// Build URLs - if domain is set, use subdomain format
	baseHost := name
	if req.Domain != "" {
		baseHost = name + "." + req.Domain
	}

	result := map[string]interface{}{
		"name":      name,
		"port":      port,
		"url":       fmt.Sprintf("http://%s.localhost", baseHost),
		"url_web":   fmt.Sprintf("http://%s.web", baseHost),
		"url_local": fmt.Sprintf("http://%s.local", baseHost),
	}
	if req.Domain != "" {
		result["domain"] = req.Domain
	}

	ipc.WriteOK(conn, result)
}

func (d *Daemon) handleUnregister(conn net.Conn, payload json.RawMessage) {
	var req map[string]string
	if err := json.Unmarshal(payload, &req); err != nil {
		ipc.WriteError(conn, fmt.Sprintf("invalid unregister request: %v", err))
		return
	}

	name := req["name"]
	if name == "" {
		ipc.WriteError(conn, "name is required")
		return
	}

	d.store.Delete(name)
	d.syncState()

	ipc.WriteOK(conn, map[string]interface{}{
		"name":    name,
		"removed": true,
	})
}

func (d *Daemon) handleStatus(conn net.Conn) {
	projects, _ := d.store.All()
	running := 0
	for _, p := range projects {
		if d.manager.IsRunning(p.Name) {
			running++
		}
	}
	ipc.WriteOK(conn, map[string]interface{}{
		"pid":      os.Getpid(),
		"projects": len(projects),
		"running":  running,
	})
}

func (d *Daemon) onProcessExit(name string, exitCode int, restart string) {
	d.logger.Info("process exited", "name", name, "exit_code", exitCode, "restart", restart)
	d.syncState()

	shouldRestart := false
	switch restart {
	case "always":
		shouldRestart = true
	case "on-failure":
		shouldRestart = exitCode != 0
	}

	if shouldRestart {
		d.logger.Info("restarting process", "name", name)
		time.Sleep(1 * time.Second)
		proj, err := d.store.Get(name)
		if err != nil {
			d.logger.Error("restart: get project", "name", name, "err", err)
			return
		}
		if err := d.manager.Spawn(proj); err != nil {
			d.logger.Error("restart: spawn", "name", name, "err", err)
		}
		d.syncState()
	}
}

func (d *Daemon) reconcileState() {
	st, err := d.store.Load()
	if err != nil {
		d.logger.Error("reconcile state", "err", err)
		return
	}

	changed := false
	for _, p := range st.Projects {
		if p.PID == 0 {
			continue
		}
		// If nothing is listening on the assigned port, the process is effectively dead.
		// Clear PID so users can re-run/register without "already running" lockouts.
		if !process.Probe(p.Port, 500*time.Millisecond) {
			d.logger.Warn("clearing stale pid", "name", p.Name, "pid", p.PID, "port", p.Port)
			p.PID = 0
			changed = true
		}
	}

	if changed {
		if err := d.store.Save(st); err != nil {
			d.logger.Error("reconcile state: save", "err", err)
		}
	}
}

func (d *Daemon) syncState() {
	st, err := d.store.Load()
	if err != nil {
		d.logger.Error("sync state", "err", err)
		return
	}
	d.proxy.UpdateState(st)
}

func (d *Daemon) shutdown() {
	d.logger.Info("shutting down")
	d.manager.KillAll()
	if d.listener != nil {
		d.listener.Close()
	}
	d.proxy.Stop()
	d.dns.Stop()
	os.Remove(ipc.SocketPath())
	os.Remove(ipc.PidPath())
	d.logger.Info("shutdown complete")
}

func tailLines(s string, n int) []string {

	lines := make([]string, 0)
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return lines
}
