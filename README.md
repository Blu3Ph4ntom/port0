# port0

No ports. Just names.

port0 auto-assigns a free port, injects `PORT` into your process, and reverse-proxies HTTP(S) traffic on port 80 to a short hostname (example: `project.localhost`, `project.web`, `project.local`).

---

## One-line installer

macOS / Linux
```bash
curl -fsSL https://raw.githubusercontent.com/blu3ph4ntom/port0/main/install.sh | bash
```

Windows (PowerShell)
```powershell
irm https://raw.githubusercontent.com/blu3ph4ntom/port0/main/install.bat | iex
```

---

## Manual (build from source)

```bash
git clone https://github.com/blu3ph4ntom/port0.git
cd port0
go build -o port0 .
# optional system install:
sudo mv port0 /usr/local/bin/port0
```

Windows manual build:
```powershell
git clone https://github.com/blu3ph4ntom/port0.git
cd port0
go build -o port0.exe .
# move to a folder on PATH, e.g. %USERPROFILE%\bin
```

---

## Quick start

Run any dev command (port0 injects `PORT`):
```bash
cd ~/projects/myapp
port0 npm run dev
# or
port0 go run ./cmd/server
```

Reach your service at:
- `http://myapp.localhost`
- `http://myapp.web`
- `http://myapp.local`

---

## One-time system setup (optional)

Run `port0 setup` if you want `.web` / `.local` resolution or need privileged port binding. Elevated privileges required.

- macOS: `sudo port0 setup` (writes /etc/resolver/* and optionally installs a LaunchDaemon)
- Linux (systemd): `sudo port0 setup` (writes systemd-resolved drop-in, may set CAP_NET_BIND_SERVICE, installs user service)
- Windows: run Administrator PowerShell and run `port0 setup` (configures firewall and NRPT rules)

To undo: `sudo port0 teardown` (or run teardown in Administrator PowerShell on Windows).

---

## Common commands

- `port0 <cmd...>` — run command with PORT injected
- `port0 -n <name> <cmd...>` — set custom name
- `port0 -d <cmd...>` — run detached/background
- `port0 ls` — list projects
- `port0 logs <name>` — view logs (`-f` to follow)
- `port0 kill <name>` — stop project
- `port0 link <name> <port>` — link existing server
- `port0 setup` / `port0 teardown` — system configuration
- `port0 update` — download & replace binary with latest release
- `port0 daemon start|stop|status` — manage daemon

---

## Integration note

port0 only injects the `PORT` env var. Ensure your app reads `PORT`:

- Node: `process.env.PORT`
- Go: `os.Getenv("PORT")`
- Python: `os.environ.get("PORT")`

---

## Language stats

This repo contains Go source and support files (installers, examples). A `.gitattributes` is present to keep language statistics focused on Go. If you want the language bar to show 100% Go, remove or relocate non-Go files (examples / installers).

---

## Update & releases

- `port0 update` fetches the latest release binary and replaces the running executable (may require sudo if installed system-wide).
- Build & release from source: `git tag vX.Y.Z && git push origin vX.Y.Z` and use your CI to publish artifacts.

---

## License

MIT