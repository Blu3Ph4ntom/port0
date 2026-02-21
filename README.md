# port0

No ports. Just names.

port0 auto-assigns a free port, injects `PORT` into your process, and reverse-proxies HTTP traffic on port 80 to a short hostname (for example `project.localhost`, `project.web`, `project.local`). Use `localhost` for zero-config.

---

## one-line installer (easy)

macOS / linux
```bash
curl -fsSL https://raw.githubusercontent.com/blu3ph4ntom/port0/main/install.sh | bash
```

windows (powershell)
```powershell
irm https://raw.githubusercontent.com/blu3ph4ntom/port0/main/install.bat | iex
```

The scripts detect OS/arch, download the proper release binary and place it in a common path (may prompt for sudo).

---

## manual (build from source)

```bash
git clone https://github.com/blu3ph4ntom/port0.git
cd port0
go build -o port0 .
# (optional) install system-wide:
sudo mv port0 /usr/local/bin/port0
```

windows manual:
```powershell
git clone https://github.com/blu3ph4ntom/port0.git
cd port0
go build -o port0.exe .
# move to a folder on PATH, e.g. %USERPROFILE%\bin
```

---

## quick start

Wrap your usual dev command with `port0`. It injects `PORT` and exposes the service at:
- `http://<name>.localhost` (no setup required)
- `http://<name>.web` (requires one-time setup)
- `http://<name>.local` (requires one-time setup)

```bash
cd ~/projects/myapp
port0 npm run dev
# or
port0 go run ./cmd/server
```

---

## one-time system setup (optional)

Only required to enable `.web` / `.local` or to allow binding privileged ports (80/53).

- macOS: `sudo port0 setup`
- linux (systemd): `sudo port0 setup`
- windows: run Administrator PowerShell, then `port0 setup`

Undo: `sudo port0 teardown` (or run port0 teardown in Administrator PowerShell on Windows).

---

## integration examples

port0 only injects the `PORT` environment variable. All of these examples wrap your existing dev command so contributor workflows stay the same.

### package.json (npm / yarn)
```json
{
  "scripts": {
    "dev": "port0 vite",
    "start": "port0 node server.js",
    "serve": "port0 -d npm run start"
  }
}
```

### pnpm workspaces (root package.json)
```json
{
  "scripts": {
    "dev:web": "cd packages/web && port0 pnpm dev",
    "dev:api": "cd packages/api && port0 pnpm --filter api dev"
  }
}
```

### bun
```bash
# run bun dev under port0
port0 bun run dev
```

### cargo (rust)
```bash
# run the `api` binary with PORT injected
port0 cargo run --bin api
```

### go
```bash
# run main (or any go command) with PORT injected
port0 go run ./cmd/server
# or if you build a binary:
port0 ./bin/server
```

### python (poetry / direct)
```bash
# poetry
port0 poetry run uvicorn myapp:app --host 0.0.0.0 --port $PORT

# plain python
port0 python -m myapp
```

Notes:
- Put `port0` into your existing `scripts` so `npm run dev` etc. continue to work for other contributors.
- For CI, prefer building from source or using pinned release artifacts rather than curl/iex one-liners.

---

## common commands (short)

- `port0 <cmd...>` — run command with `PORT` injected
- `port0 -n <name> <cmd...>` — set custom name
- `port0 -d <cmd...>` — run detached/background
- `port0 ls` — list projects
- `port0 logs <name>` — view logs (`-f` to follow)
- `port0 kill <name>` — stop project
- `port0 setup` / `port0 teardown` — system config
- `port0 update` — download & replace binary with latest release
- `port0 daemon start|stop|status` — manage daemon

---

## language stats

This repo contains Go sources and support files. A `.gitattributes` exists to keep language statistics focused on Go. If you want the language bar to show 100% Go, remove or relocate non-Go files (examples / installer scripts).

---

## license

MIT
