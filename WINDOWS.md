# Windows Setup for port0

## Issue: .web and .local domains don't resolve

On Windows, custom DNS resolution requires administrator configuration.

### Option 1: Use .localhost (NO SETUP NEEDED)

```bash
port0 npm run dev
# Access at: http://projectname.localhost
```

**This works immediately!** All browsers support .localhost natively.

### Option 2: Setup .web and .local (Requires Admin)

#### Step 1: Edit Windows hosts file

Open as Administrator:
```
C:\Windows\System32\drivers\etc\hosts
```

Add these lines:
```
127.0.0.1 node-http.web
127.0.0.1 node-http.local
127.0.0.1 python-http.web
127.0.0.1 python-http.local
```

**Problem**: You need to add EVERY project manually.

#### Step 2: Use Acrylic DNS Proxy (Recommended)

1. Download Acrylic DNS Proxy: https://mayakron.altervista.org/support/acrylic/Home.htm
2. Install and run as administrator
3. Edit `AcrylicHosts.txt`, add:
```
127.0.0.1 *.web
127.0.0.1 *.local
```
4. Set Windows DNS to `127.0.0.1`

### Option 3: Just use .localhost

Honestly, just use `.localhost` - it works everywhere with zero setup!

```bash
port0 npm run dev
# Access at: http://projectname.localhost
```
