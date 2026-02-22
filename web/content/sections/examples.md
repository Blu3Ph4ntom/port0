---
title: "examples"
weight: 7
---

<h2>Integration examples</h2>
<div class="grid">
<div>
<h3>package.json</h3>
<pre><code>{
  "scripts": {
    "dev": "port0 vite",
    "start": "port0 node server.js",
    "serve": "port0 -d npm run start"
  }
}</code></pre>
</div>
<div>
<h3>pnpm workspaces</h3>
<pre><code>{
  "scripts": {
    "dev:web": "cd packages/web && port0 pnpm dev",
    "dev:api": "cd packages/api && port0 pnpm --filter api dev"
  }
}</code></pre>
</div>
<div>
<h3>bun</h3>
<pre><code>port0 bun run dev</code></pre>
</div>
<div>
<h3>cargo</h3>
<pre><code>port0 cargo run --bin api</code></pre>
</div>
<div>
<h3>go</h3>
<pre><code>port0 go run ./cmd/server
port0 ./bin/server</code></pre>
</div>
<div>
<h3>python</h3>
<pre><code>port0 poetry run uvicorn myapp:app --host 0.0.0.0 --port $PORT
port0 python -m myapp</code></pre>
</div>
</div>
