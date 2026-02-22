---
title: "overview"
weight: 1
---

<h1>Local development without port friction.</h1>
<p class="lede">port0 is a lightweight daemon that auto-assigns free ports, injects the <code>PORT</code> environment variable, and reverse-proxies traffic to clean hostnames like <code>project.localhost</code>.</p>

<div class="cta">
<a class="btn primary" href="#install">Get Started</a>
<a class="btn external" href="https://github.com/blu3ph4ntom/port0" rel="noopener" target="_blank">GitHub</a>
</div>

<div class="features">
<div class="feature-item">
<h4>Zero Configuration</h4>
<p>No project-level config files required. Port assignment and routing are derived from your working directory automatically.</p>
</div>
<div class="feature-item">
<h4>Port Injection</h4>
<p>Automatically finds an open port and injects it into your process via environment variables. Say goodbye to port 3000 conflicts.</p>
</div>
<div class="feature-item">
<h4>Clean Hostnames</h4>
<p>Access your local apps via <code>app.localhost</code> or <code>app.web</code> instead of memorizing messy IP addresses and port numbers.</p>
</div>
<div class="feature-item">
<h4>Native Proxy & DNS</h4>
<p>Built-in reverse proxy with WebSocket support and an embedded DNS server. Works seamlessly across macOS, Linux, and Windows.</p>
</div>
</div>

<div class="quick">
<div class="cmd">
<span>macOS / Linux</span>
<code>curl -fsSL https://port0.bluephantom.dev/install.sh | bash</code>
</div>
<div class="cmd">
<span>Windows (PowerShell)</span>
<code>irm https://port0.bluephantom.dev/install.ps1 | iex</code>
</div>
</div>
