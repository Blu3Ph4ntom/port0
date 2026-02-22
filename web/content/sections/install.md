---
title: "install"
weight: 2
---

<h2>Install</h2>
<div class="grid">
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
