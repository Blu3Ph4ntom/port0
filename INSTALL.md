# port0 Installation

## Quick Install

### macOS / Linux
```bash
curl -fsSL https://raw.githubusercontent.com/bluephantom/port0/main/install.sh | bash
```

### Windows
```powershell
irm https://raw.githubusercontent.com/bluephantom/port0/main/install.bat | iex
```

## Manual Install

Download binary from [releases](https://github.com/bluephantom/port0/releases) and add to PATH.

## Setup

For *.web and *.local DNS support:
```bash
sudo port0 setup
```

## Usage

```bash
port0 run npm run dev
port0 run vite
port0 run python -m http.server
port0 run go run main.go
```

Access at: `http://projectname.localhost`

## Commands

```bash
port0 run <cmd>      # Start server
port0 ls             # List projects
port0 logs <name>    # View logs
port0 kill <name>    # Stop project
port0 daemon status  # Check daemon
```
