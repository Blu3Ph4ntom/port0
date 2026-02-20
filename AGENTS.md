# port0
### No ports. Just names.

## What this project is

port0 is a Go daemon that wraps any dev server command, auto-assigns a
free port via PORT env injection, and reverse-proxies it under a
human-readable hostname (project.localhost, project.web, project.local).
Zero config. No ports in URLs. Works with any language/framework.

***

## Repo layout

```text
port0/
├── main.go
├── cmd/
│   ├── root.go
│   ├── run.go
│   ├── ls.go
│   ├── kill.go
│   ├── open.go
│   ├── logs.go
│   └── link.go
├── internal/
│   ├── allocator/
│   │   └── allocator.go
│   ├── process/
│   │   └── manager.go
│   ├── proxy/
│   │   └── proxy.go
│   ├── dns/
│   │   └── server.go
│   ├── state/
│   │   └── state.go
│   ├── setup/
│   │   └── setup.go
│   └── util/
│       └── name.go
├── go.mod
├── go.sum
└── AGENTS.md
```

***

## Core data model

// state.go
type Project struct {
    Name      string    // "supaclaw" - derived from cwd basename
    Port      int       // auto-assigned from allocator, e.g. 4213
    Cmd       []string  // ["npm", "run", "dev"]
    Cwd       string    // absolute path of cwd at spawn time
    PID       int       // child process PID, 0 if stopped
    StartedAt time.Time
    LogFile   string    // path to ~/.port0/logs/<name>.log
}

type State struct {
    Projects map[string]*Project
    // key is Project.Name
}

State is persisted to ~/.port0/state.json on every write.
File is read fresh on every CLI invocation (no in-memory caching across
separate CLI calls - the daemon proxy owns the live state).

***

## Architecture: two runtime modes

port0 operates in two modes that coexist:

1. DAEMON MODE
   Started once (background) by the first `port0 run ...` call if not
   already running. Listens on a Unix socket at ~/.port0/daemon.sock.
   Owns:
   - reverse proxy (net/http/httputil.ReverseProxy) on :80
   - embedded DNS server (miekg/dns) on :53 for *.web
   - process supervision loop (restarts on crash if configured)
   - state.json writes

2. CLI MODE
   Every `port0 ls`, `port0 kill`, etc. connects to daemon.sock via a
   simple JSON-over-unix-socket IPC protocol and issues commands.
   The daemon responds with JSON.

If the daemon is not running, CLI commands that require it print:
  "port0 daemon not running. Start it with: port0 daemon start"

***

## Key design decisions (don't break these)

- PORT injection is the only coupling to the child process.
  port0 NEVER modifies the child process's listen address directly.
  It trusts the child to honor the PORT env var.
  If PORT is not honored, warn the user after a probe timeout (5s).

- Host header prefix routing.
  The proxy strips the TLD suffix (.localhost / .web / .local) and looks
  up only the name prefix in state. This means routing logic does NOT
  care about suffixes - DNS is the only per-suffix concern.

- No global config file.
  Project names come from cwd. Port range is a CLI flag (default 4000-4999).
  The only persisted file is ~/.port0/state.json.

- `port0 link <name>` is the escape hatch for ugly folder names.
  It writes an alias entry into state without spawning a process.
  It does NOT introduce a per-project config file.

- Port 80 / port 53 require elevated privileges.
  On first run, `port0 setup` must be run (or auto-triggered).
  On Linux: grant CAP_NET_BIND_SERVICE to the binary.
  On macOS: launchd plist runs daemon as root.
  DO NOT silently fall back to port 8080 without telling the user clearly.

***

## DNS handling per suffix

*.localhost
  Chrome and Firefox resolve *.localhost → 127.0.0.1 natively per
  RFC 6761 special-use rules. NO custom DNS needed for this suffix.
  This is the zero-config, always-works path. Prefer this in all docs.

*.web
  port0 runs an embedded DNS server (miekg/dns) on :53 (UDP+TCP).
  On setup, writes resolver config:
    macOS  → /etc/resolver/web  (nameserver 127.0.0.1)
    Linux  → systemd-resolved stub, writes drop-in under
             /etc/systemd/resolved.conf.d/port0-web.conf
  This covers *.web → 127.0.0.1.

*.local
  .local is reserved for mDNS (RFC 6762). port0 shows a warning on
  setup and marks .local support as "best-effort".
  DO NOT implement unicast DNS override for .local without a loud
  warning. The mDNS conflict will break Bonjour/Avahi on the machine.
  Recommended: tell users to prefer .localhost or .web.

***

## Process lifecycle rules

Spawn:
  1. Normalize cwd basename → project name (lowercase, strip non-alphanum
     to hyphens, deduplicate hyphens)
  2. If name already exists in state and PID is alive → error, show URL
  3. If name exists but PID is dead → reuse name, reassign port
  4. If two different cwds collide on the same name → auto-suffix: api-2,
     api-3, ... and warn the user

Probe:
  After spawn, poll :PORT with a TCP dial every 500ms for up to 10s.
  If no bind detected in 10s → warn "process may not honor PORT env var"
  but keep the process running. Do NOT kill it.

Death:
  If the child exits (any exit code), log it to ~/.port0/logs/<name>.log
  and mark PID=0 in state. Do not auto-restart by default.
  Auto-restart is an opt-in flag: port0 run --restart always <cmd>

***

## Proxy implementation notes

Use net/http/httputil.ReverseProxy.
Director func:
  1. Parse Host header → extract name prefix (split on first ".")
  2. Look up name in state → get port
  3. Rewrite req.URL to http://127.0.0.1:<port>
  4. Preserve original Host header for apps that inspect it

WebSocket:
  Handle Upgrade: websocket headers - use a raw TCP tunnel
  (hijack the connection) for WS proxying. httputil.ReverseProxy
  does NOT handle WS out of the box.

HTTPS (optional flag: --tls):
  Use a self-signed cert per project generated via crypto/x509.
  Store certs in ~/.port0/certs/<name>.pem + <name>-key.pem.
  Proxy listens on :443 for TLS, :80 for plain.
  DO NOT use mkcert as a hard dependency - generate internally or
  optionally delegate to mkcert if present on PATH.

***

## IPC protocol (daemon.sock)

Simple newline-delimited JSON.

Request:
  { "op": "spawn", "payload": { ... } }
  { "op": "kill",  "payload": { "name": "supaclaw" } }
  { "op": "ls" }
  { "op": "logs",  "payload": { "name": "supaclaw", "follow": true } }

Response:
  { "ok": true,  "data": { ... } }
  { "ok": false, "error": "project not found: supaclaw" }

For `logs --follow`, the daemon streams lines as:
  { "line": "...", "ts": "..." }
until the client disconnects.

***

## Logging

Child process stdout+stderr are tee'd:
  - to the terminal if port0 was run in foreground (default)
  - always to ~/.port0/logs/<name>.log

Log rotation: rotate at 10MB, keep last 3 files. Use lumberjack or
implement manually. DO NOT let logs grow unbounded.

port0 daemon internal logs → ~/.port0/daemon.log (not shown to users
by default, only on --verbose flag).

***

## Testing approach

Unit:
  - allocator: test port range scanning, collision avoidance
  - state: marshal/unmarshal, concurrent write safety (use flock or mutex)
  - name: normalization edge cases (spaces, dots, unicode, empty)
  - proxy director: Host header parsing and rewrite logic

Integration:
  - spawn a real `nc -l $PORT` as child, verify probe detects bind
  - send HTTP request with Host: supaclaw.localhost, verify 200 proxied
  - DNS query for supaclaw.web against embedded server, verify A record

No mocks for the proxy or DNS - use real listeners on random ports in tests.

***

## Things NOT to do

- Don't add a per-project config file (port0.toml, .port0rc, etc.).
  The whole point is zero config. link is the only escape hatch.

- Don't use dnsmasq as a dependency. Everything DNS must be embedded
  (miekg/dns). External deps break the "single binary" promise.

- Don't silently fall back to a non-80 port for the proxy. If :80 is
  unavailable, error loudly and tell the user to run `port0 setup`.

- Don't use a global mutex around state reads in the proxy hot path.
  Copy state into a sync.Map or swap an atomic pointer on writes.

- Don't log to stdout in the daemon - it's a background process.
  All daemon output goes to ~/.port0/daemon.log.

- Don't auto-open a browser on `port0 run`. That's what `port0 open` is for.

***

## CLI UX conventions

- All commands accept --json flag for machine-readable output.
- Error messages go to stderr, formatted as: "error: <message>"
- Success output is minimal - show only what changed or was looked up.
- `port0 ls` output columns: NAME | PORT | URL | PID | STATUS | STARTED
- Use lipgloss or plain fmt - no heavy TUI frameworks (no bubbletea for
  basic ls output).

***

## Dependencies (intentional, keep minimal)

github.com/spf13/cobra        - CLI command tree
github.com/miekg/dns          - embedded DNS server
github.com/natefinsh/color    - terminal color output (optional)
gopkg.in/natefinch/npipe.v2   - named pipe for Windows (future)

Everything else: stdlib only.
net/http/httputil              - reverse proxy
os/exec                        - child process management
encoding/json                  - state + IPC
crypto/tls + crypto/x509       - self-signed certs for --tls

***

## First-class OS support

Priority order:
  1. macOS (primary, tested first)
  2. Linux (systemd-resolved assumed)
  3. Windows (future - named pipes, no port 80 by default)

For any OS-specific code, gate with build tags:
  setup_darwin.go
  setup_linux.go
  setup_windows.go

***

## Versioning and release

Single static binary. Built with:
  CGO_ENABLED=0 GOOS=... GOARCH=... go build -ldflags="-s -w" -o port0 .

Version embedded via ldflags:
  -ldflags "-X main.Version=$(git describe --tags --always)"

Releases via GitHub Actions → goreleaser.
