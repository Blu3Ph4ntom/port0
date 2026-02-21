---
title: "port0"
---

<section id="overview" class="panel-section">
<h1>Local development without port friction.</h1>
<p class="lede">port0 auto-assigns a free port, injects <code>PORT</code>, and reverse-proxies HTTP traffic on port 80 to clean hostnames like <code>project.localhost</code>, <code>project.web</code>, and <code>project.local</code>.</p>
<div class="cta">
<a class="btn primary" href="#install">Install</a>
<a class="btn external" href="https://github.com/blu3ph4ntom/port0" rel="noopener" target="_blank">GitHub</a>
</div>
<div class="quick">
<div class="cmd">
<span>macOS / Linux</span>
<code>curl -fsSL https://port0.bluephantom.dev/install.sh | bash</code>
</div>
<div class="cmd">
<span>Windows (PowerShell)</span>
<code>irm https://port0.bluephantom.dev/install.bat | iex</code>
</div>
</div>
<div class="section-footer">
<a class="edit-link external" href="https://github.com/blu3ph4ntom/port0/edit/main/web/content/_index.md" rel="noopener" target="_blank">Edit on GitHub</a>
</div>
</section>

<section id="install" class="panel-section">
<h2>Install</h2>
<div class="grid two">
<div>
<h3>One-line installer</h3>
<p>Detects OS/arch, downloads the proper release binary, and installs to a common path (may prompt for sudo).</p>
<pre><code>curl -fsSL https://port0.bluephantom.dev/install.sh | bash</code></pre>
<pre><code>irm https://port0.bluephantom.dev/install.bat | iex</code></pre>
</div>
<div>
<h3>Build from source</h3>
<pre><code>git clone https://github.com/blu3ph4ntom/port0.git
cd port0
go build -o port0 .
sudo mv port0 /usr/local/bin/port0</code></pre>
<pre><code>git clone https://github.com/blu3ph4ntom/port0.git
cd port0
go build -o port0.exe .
# move to a folder on PATH, e.g. %USERPROFILE%\bin</code></pre>
</div>
</div>
<div class="section-footer">
<a class="edit-link external" href="https://github.com/blu3ph4ntom/port0/edit/main/web/content/_index.md" rel="noopener" target="_blank">Edit on GitHub</a>
</div>
</section>

<section id="quickstart" class="panel-section">
<h2>Quick start</h2>
<p>Wrap your usual dev command with <code>port0</code>. It injects <code>PORT</code> and exposes the service at:</p>
<ul>
<li><code>http://&lt;name&gt;.localhost</code> (no setup required)</li>
<li><code>http://&lt;name&gt;.web</code> (requires one-time setup)</li>
<li><code>http://&lt;name&gt;.local</code> (requires one-time setup)</li>
</ul>
<pre><code>cd ~/projects/myapp
port0 npm run dev
# or
port0 go run ./cmd/server</code></pre>
<div class="section-footer">
<a class="edit-link external" href="https://github.com/blu3ph4ntom/port0/edit/main/web/content/_index.md" rel="noopener" target="_blank">Edit on GitHub</a>
</div>
</section>

<section id="subdomains" class="panel-section">
<h2>Subdomain support</h2>
<p>Group related projects under a shared parent domain for clean URLs in monorepos.</p>
<div class="grid two">
<div>
<h3>Quick syntax</h3>
<pre><code>port0 -n api.myapp npm run dev    # api.myapp.localhost
port0 -n web.myapp npm run dev    # web.myapp.localhost</code></pre>
</div>
<div>
<h3>Explicit syntax</h3>
<pre><code>port0 -n api --domain myapp npm run dev    # api.myapp.localhost
port0 -n web --domain myapp npm run dev    # web.myapp.localhost</code></pre>
</div>
</div>
<pre><code>port0 -n myapp npm run dev        # myapp.localhost
port0 npm run dev                 # uses folder name</code></pre>
<p>Use this for monorepos, micro-frontends, multi-repo domains, or environment separation.</p>
<div class="section-footer">
<a class="edit-link external" href="https://github.com/blu3ph4ntom/port0/edit/main/web/content/_index.md" rel="noopener" target="_blank">Edit on GitHub</a>
</div>
</section>

<section id="setup" class="panel-section">
<h2>One-time system setup (optional)</h2>
<p>Only required to enable <code>.web</code> / <code>.local</code> or to allow binding privileged ports (80/53).</p>
<ul>
<li>macOS: <code>sudo port0 setup</code></li>
<li>linux (systemd): <code>sudo port0 setup</code></li>
<li>windows: run Administrator PowerShell, then <code>port0 setup</code></li>
</ul>
<p>Undo: <code>sudo port0 teardown</code> (or run <code>port0 teardown</code> in Administrator PowerShell on Windows).</p>
<div class="section-footer">
<a class="edit-link external" href="https://github.com/blu3ph4ntom/port0/edit/main/web/content/_index.md" rel="noopener" target="_blank">Edit on GitHub</a>
</div>
</section>

<section id="commands" class="panel-section">
<h2>Common commands</h2>
<div class="grid two">
<div>
<ul>
<li><code>port0 &lt;cmd...&gt;</code> — run command with <code>PORT</code> injected</li>
<li><code>port0 -n &lt;name&gt; &lt;cmd...&gt;</code> — set custom name</li>
<li><code>port0 -d &lt;cmd...&gt;</code> — run detached/background</li>
<li><code>port0 ls</code> — list projects</li>
</ul>
</div>
<div>
<ul>
<li><code>port0 logs &lt;name&gt;</code> — view logs (<code>-f</code> to follow)</li>
<li><code>port0 kill &lt;name&gt;</code> — stop project</li>
<li><code>port0 setup</code> / <code>port0 teardown</code> — system config</li>
<li><code>port0 update</code> — download & replace binary with latest release</li>
</ul>
</div>
</div>
<p><code>port0 daemon start|stop|status</code> — manage daemon</p>
<div class="section-footer">
<a class="edit-link external" href="https://github.com/blu3ph4ntom/port0/edit/main/web/content/_index.md" rel="noopener" target="_blank">Edit on GitHub</a>
</div>
</section>

<section id="examples" class="panel-section">
<h2>Integration examples</h2>
<div class="grid two">
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
<div class="section-footer">
<a class="edit-link external" href="https://github.com/blu3ph4ntom/port0/edit/main/web/content/_index.md" rel="noopener" target="_blank">Edit on GitHub</a>
</div>
</section>