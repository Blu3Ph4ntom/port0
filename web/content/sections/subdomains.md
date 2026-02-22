---
title: "subdomains"
weight: 4
---

<h2>Subdomain support</h2>
<p>Group related projects under a shared parent domain for clean URLs in monorepos.</p>
<div class="grid">
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
