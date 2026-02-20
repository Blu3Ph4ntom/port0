# port0 Examples

Test each example to verify port0 works with all frameworks.

## Node.js HTTP Server

```bash
cd examples/node-http
../../port0 node server.js

# Test
curl -H "Host: node-http.localhost" http://127.0.0.1/
```

## Python HTTP Server

```bash
cd examples/python-http
../../port0 python server.py

# Test
curl -H "Host: python-http.localhost" http://127.0.0.1/
```

## Go HTTP Server

```bash
cd examples/go-http
../../port0 go run main.go

# Test
curl -H "Host: go-http.localhost" http://127.0.0.1/
```

## PHP Built-in Server

```bash
cd examples/php-server
../../port0 php -S 0.0.0.0:\$PORT

# Test
curl -H "Host: php-server.localhost" http://127.0.0.1/
```

## Ruby WEBrick

```bash
cd examples/ruby-server
../../port0 ruby server.rb

# Test
curl -H "Host: ruby-server.localhost" http://127.0.0.1/
```

## Deno HTTP Server

```bash
cd examples/deno-http
../../port0 deno run --allow-net --allow-env server.ts

# Test
curl -H "Host: deno-http.localhost" http://127.0.0.1/
```

## All at once (monorepo simulation)

```bash
cd examples/node-http && ../../port0 -n node -d node server.js
cd examples/python-http && ../../port0 -n python -d python server.py
cd examples/go-http && ../../port0 -n go -d go run main.go

# List all
../../port0 ls

# Test all
curl -H "Host: node.localhost" http://127.0.0.1/
curl -H "Host: python.localhost" http://127.0.0.1/
curl -H "Host: go.localhost" http://127.0.0.1/

# Kill all
../../port0 kill node
../../port0 kill python
../../port0 kill go
```
