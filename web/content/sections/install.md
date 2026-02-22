---
title: "install"
weight: 2
---

<h2>Install</h2>
<div class="grid">
<div>
<h3>One-line installer</h3>
<p>Detects OS/arch, downloads the proper release binary, and installs to a common path (may prompt for sudo).</p>
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
</div>
<div>
<h3>Build from source</h3>
<div class="quick">
    <div class="cmd">
        <span>macOS / Linux</span>
        <code>git clone https://github.com/blu3ph4ntom/port0.git </br>
        cd port0 </br>
        go build -o port0 . </br>
        sudo mv port0 /usr/local/bin/port0</code>
    </div>
    <div class="cmd">
        <span>Windows (PowerShell)</span>
        <code>git clone https://github.com/blu3ph4ntom/port0.git </br>
        cd port0 </br>
        go build -o port0.exe . </br>
        # move to a folder on PATH, e.g. %USERPROFILE%\bin</code>
    </div>
</div>
</div>
</div>
