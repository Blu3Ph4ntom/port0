# port0

No ports. Just names.

port0 auto-assigns a free port, injects `PORT` into your process, and reverse-proxies HTTP(S) traffic on port 80 to a human-friendly hostname (for example: `project.localhost`).

---

## One-line installer

macOS / Linux (downloads release binary and installs to /usr/local/bin):
```bash
curl -fsSL https://github.com/blu3ph4ntom/port0/releases/latest/download/port0-linux-amd64 -o /tmp/port0 && \
chmod +x /tmp/port0 && sudo mv /tmp/port0 /usr/local/bin/port0
```

Windows (PowerShell):
```powershell
Invoke-WebRequest -Uri "https://github.com/blu3ph4ntom/port0/releases/latest/download/port0-windows-amd64.exe" -OutFile "$env:TEMP\port0.exe"
Move-Item "$env:TEMP\port0.exe" "$env:USERPROFILE\bin\port0.exe"
```

---

## Manual (build from source)

```bash
git clone https://github.com/blu3ph4ntom/port0.git
cd port0
go build -o port0 .
# optional: install system-wide
sudo mv port0 /usr/local/bin/port0
```

---

## How it works

```mermaid
flowchart TB
  Run["port0 <cmd>"] --> Name["Derive name from folder"]
  Name --> Assign["Assign free port (4000-4999)"]
  Assign --> Spawn["Spawn process with PORT=<port>"]
  Spawn --> Proxy["Reverse proxy :80 -> 127.0.0.1:<port>"]
  Browser["Browser request to name.localhost"] --> Proxy
```

---

## One-time setup (only if you need `.web` / `.local`)

`.localhost` works without setup.

- macOS
  - `sudo port0 setup` (creates `/etc/resolver/web` and `/etc/resolver/local` and installs a LaunchDaemon)
- Linux (systemd)
  - `sudo port0 setup` (writes systemd-resolved drop-in, sets CAP_NET_BIND_SERVICE, installs a user service)
- Windows
  - Run Administrator PowerShell and run `port0 setup` (adds firewall rules and NRPT rules)

Remove system configuration:
```bash
sudo port0 teardown
# Windows: run teardown in Administrator PowerShell
```

---

## Usage (common)

Run in foreground (shows logs):
```bash
port0 npm run dev
```

Run detached:
```bash
port0 -d npm run dev
```

Custom name:
```bash
port0 -n myapi go run ./cmd/server
```

List / logs / kill:
```bash
port0 ls
port0 logs myapi
port0 kill myapi
```

Primary URL exposed: `http://myapp.localhost`  
Alternative TLDs: `http://myapp.web`, `http://myapp.local`

---

## Commands (summary)

- `port0 <cmd...>` — run command with PORT injection
- `port0 -n <name> <cmd...>` — custom name
- `port0 -d <cmd...>` — detached/background
- `port0 ls` — list projects
- `port0 logs <name>` — view logs (`-f` to follow)
- `port0 kill <name>` — stop project
- `port0 link <name> <port>` — link existing server
- `port0 setup` / `port0 teardown` — system configuration
- `port0 update` — download & replace binary with latest release
- `port0 daemon start|stop|status` — manage daemon

---

## Integration note

port0 only injects the `PORT` environment variable. Ensure your app reads it:
- Node: `process.env.PORT`
- Go: `os.Getenv("PORT")`
- Python: `os.environ.get('PORT')`

For monorepos use `-n` to choose distinct names.

---

## Update & releases

- `port0 update` downloads the latest release binary and replaces the current executable (may require sudo if installed system-wide).
- To build and publish releases: tag and push (for example `git tag v0.1.0 && git push origin v0.1.0`) and run your release CI.

---

## Language stats

A `.gitattributes` file marks docs, installer scripts and `examples/` as vendored/documentation so GitHub Linguist focuses on Go. GitHub may take a short time to reindex the language graph.

---

## License

MIT