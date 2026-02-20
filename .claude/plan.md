# port0 Build Plan

## Overview
port0 is a Go daemon that wraps any dev server, auto-assigns a free port via PORT env injection, and reverse-proxies under human-readable hostnames (proj.localhost, proj.web, proj.local).

## Architecture
- **Daemon mode**: Background process owning reverse proxy (:80), DNS (:53), process supervision, state writes
- **CLI mode**: Connects to daemon via Unix socket IPC, sends JSON ops, prints results

## Status: ALL PHASES COMPLETE

### Repo Structure (Final)
```
port0/
├── main.go                          # Entry point with daemon/CLI switch
├── daemon_entry.go                  # Daemon mode entry
├── go.mod / go.sum                  # Dependencies
├── .goreleaser.yml                  # Cross-platform release
├── .github/workflows/
│   ├── ci.yml                       # CI: build+test+vet on 3 OS
│   └── release.yml                  # Release on tag push
├── cmd/
│   ├── root.go                      # Persistent flags, dir init
│   ├── run.go                       # Spawn with auto-daemon
│   ├── ls.go                        # Tabular listing
│   ├── kill.go                      # Stop/remove
│   ├── open.go                      # Browser launch
│   ├── logs.go                      # Tail/follow
│   ├── link.go                      # Alias entry
│   ├── daemon.go                    # start/stop/status
│   ├── setup.go                     # OS config
│   ├── daemon_proc_unix.go          # Unix fork attrs
│   └── daemon_proc_windows.go       # Windows fork attrs
├── internal/
│   ├── allocator/allocator.go       # Port scanning + tests
│   ├── cert/cert.go                 # Self-signed TLS
│   ├── daemon/daemon.go             # Full daemon runner
│   ├── dns/server.go                # Embedded DNS + tests
│   ├── ipc/ipc.go                   # JSON protocol + tests
│   ├── process/manager.go           # Spawn/kill/probe
│   ├── proxy/proxy.go               # Reverse proxy + tests
│   ├── setup/                       # OS-specific (darwin/linux/windows)
│   ├── state/state.go               # State persistence + tests
│   └── util/name.go                 # Name normalization + tests
└── .claude/
    ├── plan.md                      # This file
    └── progress.md                  # Build progress tracker
```

## Tech Stack
- Go 1.22+
- cobra (CLI), miekg/dns (DNS), fatih/color (terminal), lumberjack (log rotation)
- Everything else: stdlib (net/http, crypto/tls, os/exec, encoding/json)

## Key Design Decisions
1. PORT injection is the only coupling to child processes
2. Host header prefix routing (suffix-agnostic)
3. No global config file - names from cwd, state.json only
4. Atomic pointer for proxy state reads (no locks on hot path)
5. Single static binary via CGO_ENABLED=0
