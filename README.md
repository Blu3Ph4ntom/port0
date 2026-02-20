# port0

No ports. Just names.

Auto-assigns ports, injects PORT env, proxies to `project.localhost`

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/blu3ph4ntom/port0/main/install.sh | bash
```

## Usage

```bash
port0 npm run dev
port0 python -m http.server
port0 go run main.go
```

Access at: `http://projectname.localhost` (or `.web`, `.local`)

## Commands

```bash
port0 <cmd>              # Start server (default)
port0 -n myapp npm start # Custom name
port0 -d npm run dev     # Background mode
port0 ls                 # List projects  
port0 logs <name>        # View logs
port0 kill <name>        # Stop project
```

## How it works

1. Wraps your dev command
2. Assigns free port (4000-4999)
3. Injects PORT env var
4. Reverse proxies on :80
5. Access via `project.localhost`

Zero config. Works everywhere.
