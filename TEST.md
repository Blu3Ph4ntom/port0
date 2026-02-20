# Quick Test

## Test the new features:

```bash
# Rebuild
go build -o port0.exe .

# Test with Node
cd examples
../port0.exe run node node-server.js

# You should see:
#   ✓ Server started
#   Name:  examples
#   Port:  XXXX
#   Access at:
#     • http://examples.localhost
#     • http://examples.web
#     • http://examples.local
#   
#   (live logs appear here)

# Test proxy (in another terminal):
curl -H "Host: examples.localhost" http://127.0.0.1:80/
curl -H "Host: examples.web" http://127.0.0.1:80/
curl -H "Host: examples.local" http://127.0.0.1:80/

# All three should work!

# Test background mode:
../port0.exe run -d node node-server.js

# View logs:
../port0.exe logs examples

# Clean up:
../port0.exe kill examples
../port0.exe daemon stop
```

## What's new:

1. **.local DNS works** - added to DNS handler
2. **Beautiful output** - colored, formatted with all URLs
3. **Foreground by default** - see logs live
4. **Background mode** - use `-d` flag
5. **Easy install** - scripts for all platforms

## To use on any project:

```bash
port0 run npm run dev
port0 run vite
port0 run python -m http.server
port0 run go run main.go
```

Works with ANY command that honors PORT env var!
