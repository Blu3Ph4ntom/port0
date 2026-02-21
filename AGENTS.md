# port0
### No ports. Just names.

## Project Overview
port0 is a Go-based local development daemon that automates the assignment of free ports via `PORT` environment variable injection and provides human-readable hostnames (e.g., `project.localhost`, `project.web`) via a reverse proxy.

## Repo Layout
```text
port0/
├── main.go                 // Main entry point (switches between CLI and Daemon)
├── daemon_entry.go         // Daemon bootstrapper
├── cmd/                    // Cobra CLI Commands
│   ├── root.go             // CLI setup and versioning
│   ├── daemon.go           // Control the daemon process (start/stop/status)
│   ├── run.go              // Spawn a background managed process
│   ├── ls.go               // List projects and their statuses
│   ├── kill.go             // Stop processes or remove project entries
│   ├── open.go             // Open project URL in the default browser
│   ├── logs.go             // Stream/tail stdout/stderr for a project
│   ├── link.go             // Map a static port to a name without spawning
│   └── setup.go            // OS-specific configuration (Privileged ports/DNS)
├── internal/
│   ├── allocator/          // Scans and picks free ports from a range
│   ├── cert/               // Internal self-signed TLS certificate generation
│   ├── daemon/             // Core daemon server and IPC request handling
│   ├── dns/                // Embedded DNS server (miekg/dns) on :53
│   ├── ipc/                // Unix socket protocol (newline-delimited JSON)
│   ├── process/            // Lifecycle management (spawn, probe, kill, restart)
│   ├── proxy/              // httputil.ReverseProxy + WebSocket hijacking
│   ├── state/              // persistence logic for ~/.port0/state.json
│   ├── setup/              // OS-specific implementation (Darwin, Linux, Windows)
│   └── util/               // Name normalization and deconfliction
└── go.mod
```

## Core Data Model
Defined in `internal/state/state.go`:

```go
type Project struct {
    Name      string    // Normalized name (e.g., "my-api")
    Port      int       // Assigned port (e.g., 4001)
    Cmd       []string  // Command executed (e.g., ["npm", "run", "dev"])
    Cwd       string    // Absolute path where the command was run
    PID       int       // System PID (0 if stopped)
    StartedAt time.Time // Timestamp of last start
    LogFile   string    // Path to ~/.port0/logs/<name>.log
    Restart   string    // "no", "always", "on-failure"
    Domain    string    // Optional parent domain for subdomain routing
}

type State struct {
    Projects map[string]*Project // key: Project.Name
}
```

## Architecture
- **Daemon Mode**: Background process owning Port 80 (HTTP), Port 443 (HTTPS), Port 53 (DNS), and the Unix Socket (`~/.port0/daemon.sock`). It manages the lifecycle of all child processes.
- **CLI Mode**: Stateless commands that talk to the daemon over the Unix socket.
- **Persistence**: `~/.port0/state.json` is the source of truth, updated by the daemon.

## Key Design Principles
1. **Zero Config**: No project-level config files. Configuration is derived from CWD or CLI flags.
2. **PORT Injection**: The app MUST use the `PORT` environment variable. port0 probes the port for 10s after spawning to verify binding.
3. **Subdomains**: If `Domain` is set (e.g., "myapp"), requests to `*.myapp.localhost` are routed to that project.
4. **Proxy**: Uses `httputil.ReverseProxy`. Handles WebSockets by hijacking the connection and performing a raw TCP copy.
5. **DNS**:
    - `.localhost`: Resolved by OS/Browser natively.
    - `.web`: Handled by internal DNS. Requires `port0 setup`.
    - `.local`: Best-effort (mDNS conflict risk).

## IPC Protocol
The CLI sends newline-delimited JSON to the daemon:
- `spawn`: Execute a command in the background.
- `register`: Add a project being run manually in a foreground terminal.
- `kill`: Terminate a process (optional `--remove` to delete from state).
- `ls`: Return all project metadata.
- `logs`: Stream logs (supports `--follow`).
- `status`: Return daemon PID and project counts.

## Environment & Logs
- **Daemon Logs**: `~/.port0/daemon.log` (JSON via `slog`).
- **Project Logs**: `~/.port0/logs/<name>.log` (Tee'd output).
- **Rotation**: Lumberjack handles 10MB rotations with 3 backups.