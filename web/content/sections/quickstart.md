---
title: "quickstart"
weight: 3
---

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
