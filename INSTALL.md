# port0 Installation

## Quick Install

### macOS / Linux
```bash
curl -fsSL https://raw.githubusercontent.com/blu3ph4ntom/port0/main/install.sh | bash
```

### Windows
```powershell
irm https://raw.githubusercontent.com/blu3ph4ntom/port0/main/install.bat | iex
```

## Manual Install

Download binary from [releases](https://github.com/blu3ph4ntom/port0/releases) and add to PATH.

## Setup

For *.web and *.local DNS support:
```bash
sudo port0 setup
```

## Usage

```bash
port0 npm run dev
port0 vite
port0 python -m http.server
port0 go run main.go

# Custom name
port0 -n myapi npm start

# Background mode
port0 -d npm run dev
```

Access at: `http://projectname.localhost`

## Commands

```bash
port0 <cmd>          # Start server (default)
port0 ls             # List projects
port0 logs <name>    # View logs
port0 kill <name>    # Stop project
port0 daemon status  # Check daemon
```
