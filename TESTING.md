# port0 Testing Guide

## Prerequisites

Make sure you have these installed:
- Node.js
- Python 3
- Go
- PHP (optional)
- Ruby (optional)
- Deno (optional)

## Quick Test (Node.js only)

```bash
cd examples/node-http
../../port0 node server.js
```

In another terminal:
```bash
curl -H "Host: node-http.localhost" http://127.0.0.1/
```

Expected output: HTML page showing port, host, path

## Full Test Suite

### 1. Test Each Language

```bash
# Terminal 1 - Node
cd examples/node-http
../../port0 node server.js

# Terminal 2 - Python
cd examples/python-http
../../port0 python server.py

# Terminal 3 - Go
cd examples/go-http
../../port0 go run main.go
```

### 2. Test Custom Names (Monorepo)

```bash
cd examples/node-http
../../port0 -n api node server.js &

cd examples/python-http
../../port0 -n web python server.py &

cd examples/go-http
../../port0 -n admin go run main.go &

# List all
../../port0 ls
```

### 3. Test All URLs

```bash
curl -H "Host: node-http.localhost" http://127.0.0.1/
curl -H "Host: node-http.web" http://127.0.0.1/
curl -H "Host: node-http.local" http://127.0.0.1/
```

### 4. Test Background Mode

```bash
cd examples/node-http
../../port0 -d node server.js

# View logs
../../port0 logs node-http

# Kill
../../port0 kill node-http
```

### 5. Test Daemon Commands

```bash
../../port0 daemon status
../../port0 ls
../../port0 daemon stop
```

## Expected Results

✅ Each server starts and shows "Server listening on port XXXX"
✅ Port is auto-assigned from 4000-4999 range
✅ All three URLs (.localhost, .web, .local) respond
✅ Proxy correctly routes to backend
✅ Logs stream in real-time (foreground mode)
✅ Custom names work for monorepo scenarios
✅ Background mode works with -d flag

## Common Issues

**Issue**: "unknown command" error
**Fix**: Make sure you're using `port0 <cmd>` not `port0 run <cmd>`

**Issue**: DNS not resolving .web or .local
**Fix**: Run `sudo port0 setup` first

**Issue**: Port 80 permission denied
**Fix**: Run daemon with elevated privileges or use setup

## Cleanup

```bash
# Kill all running projects
../../port0 ls
../../port0 kill <name>

# Stop daemon
../../port0 daemon stop
```
