# Complete Setup Guide - balooProxyExtended

**Follow this guide step-by-step for a successful installation.**

---

## 📋 Prerequisites Checklist

Before starting, ensure you have:

- [ ] Server with Linux/macOS/Windows
- [ ] Go 1.19 or higher installed
- [ ] Git installed
- [ ] Domain name configured
- [ ] SSL certificate (or ready to generate)
- [ ] Backend application running
- [ ] Root/sudo access (for production setup)

---

## 🚀 Step-by-Step Installation

### Step 1: Install Go (if not installed)

**Check if Go is installed:**
```bash
go version
```

**If not installed:**

**Linux:**
```bash
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
go version
```

**macOS:**
```bash
brew install go
go version
```

**Windows:**
- Download from [golang.org/dl](https://golang.org/dl/)
- Run installer
- Open new terminal and run `go version`

---

### Step 2: Clone or Download Repository

```bash
# Option 1: Clone with git
git clone https://github.com/YOUR_USERNAME/balooProxyExtended.git
cd balooProxyExtended

# Option 2: Download ZIP
# Extract and cd into directory
```

---

### Step 3: Install Dependencies

**This is crucial for new features to work!**

```bash
# Install Prometheus client library
go get github.com/prometheus/client_golang@v1.19.0

# Install all dependencies
go mod tidy

# Verify no errors
echo "Dependencies installed successfully!"
```

**Expected output:**
```
go: downloading github.com/prometheus/client_golang v1.19.0
go: downloading github.com/prometheus/client_model v0.6.0
...
```

---

### Step 4: Build the Proxy

```bash
# Standard build
go build -o main

# OR optimized build (recommended for production)
go build -ldflags="-s -w" -o main

# Verify build
ls -lh main
./main --help
```

**Expected output:**
```
-rwxr-xr-x 1 user user 15M Oct 6 16:00 main
```

---

### Step 5: Generate Initial Configuration

```bash
# Run proxy to generate config
./main
```

**Answer the prompts:**
1. Domain name: `example.com`
2. Backend address: `127.0.0.1:8080`
3. Use Cloudflare: `n` (or `y` if using Cloudflare)
4. Certificate path: `assets/server/server.crt`
5. Key path: `assets/server/server.key`

**Press Ctrl+C after answering all questions.**

A `config.json` file will be created.

---

### Step 6: Configure Secrets

**CRITICAL: Change all CHANGE_ME values!**

```bash
# Open config.json
nano config.json

# OR use the example with features
cp examples/config-with-features.json config.json
nano config.json
```

**Generate secure secrets:**
```bash
# Linux/macOS - generate 5 random secrets
for i in {1..5}; do openssl rand -hex 32; done
```

**Update these fields:**
```json
{
  "proxy": {
    "adminsecret": "YOUR_RANDOM_STRING_1",
    "apisecret": "YOUR_RANDOM_STRING_2",
    "secrets": {
      "captcha": "YOUR_RANDOM_STRING_3",
      "cookie": "YOUR_RANDOM_STRING_4",
      "javascript": "YOUR_RANDOM_STRING_5"
    }
  }
}
```

---

### Step 7: Configure Domain and Backend

```json
{
  "domains": [
    {
      "name": "example.com",
      "backend": "127.0.0.1:8080",
      "scheme": "http",
      "certificate": "assets/server/server.crt",
      "key": "assets/server/server.key"
    }
  ]
}
```

**Replace:**
- `example.com` with your actual domain
- `127.0.0.1:8080` with your backend address
- Certificate paths with your actual paths

---

### Step 8: Enable New Features

Add to your `config.json`:

```json
{
  "proxy": {
    "metrics": {
      "enabled": true,
      "port": 9090,
      "path": "/metrics"
    },
    "deduplication": {
      "enabled": true,
      "ttl_seconds": 30
    }
  }
}
```

---

### Step 9: Set Up SSL Certificates

**Option A: Let's Encrypt (Recommended)**

```bash
# Install certbot
sudo apt install certbot

# Generate certificate
sudo certbot certonly --standalone -d example.com

# Update config.json
{
  "certificate": "/etc/letsencrypt/live/example.com/fullchain.pem",
  "key": "/etc/letsencrypt/live/example.com/privkey.pem"
}
```

**Option B: Self-Signed (Testing Only)**

```bash
# Generate self-signed certificate
openssl req -x509 -newkey rsa:4096 \
  -keyout server.key -out server.crt \
  -days 365 -nodes \
  -subj "/CN=example.com"

# Move to assets directory
mkdir -p assets/server
mv server.crt server.key assets/server/
```

---

### Step 10: Configure DNS

**Get your server IP:**
```bash
curl ifconfig.me
```

**In your DNS provider:**

**Without Cloudflare:**
- Type: `A`
- Name: `example.com` (or `@` for root)
- Value: `YOUR_SERVER_IP`
- TTL: `300`

**With Cloudflare:**
- Type: `A`
- Name: `example.com`
- Value: `YOUR_SERVER_IP`
- Proxy Status: `Proxied` (orange cloud)
- Set `"cloudflare": true` in config.json

**Wait 5-10 minutes for DNS propagation.**

---

### Step 11: Test the Proxy

```bash
# Run in foreground (for testing)
./main

# You should see:
# Starting Proxy ...
# Loaded Config ...
# Initialising ...
# [+] [ Cpu Usage ] > [ 2.50 ]
# ...
```

**In another terminal:**

```bash
# Test HTTP (should redirect to HTTPS)
curl -I http://example.com

# Test HTTPS
curl -I https://example.com

# Test metrics
curl http://localhost:9090/metrics

# Test deduplication
for i in {1..10}; do curl -s http://example.com/test & done
curl http://localhost:9090/metrics | grep deduplication
```

**Expected results:**
- HTTP redirects to HTTPS (301)
- HTTPS returns 200 OK
- Metrics endpoint returns Prometheus data
- Deduplication shows savings > 0

---

### Step 12: Set Up as Service (Production)

```bash
# Create user
sudo useradd -r -s /bin/false balooproxy

# Create directory
sudo mkdir -p /opt/balooproxy
sudo cp main config.json /opt/balooproxy/
sudo cp -r assets /opt/balooproxy/

# Set permissions
sudo chown -R balooproxy:balooproxy /opt/balooproxy

# Set capabilities (for ports 80/443)
sudo setcap 'cap_net_bind_service=+ep' /opt/balooproxy/main

# Copy service file
sudo cp examples/balooproxy.service /etc/systemd/system/

# Edit service file if needed
sudo nano /etc/systemd/system/balooproxy.service

# Enable and start
sudo systemctl daemon-reload
sudo systemctl enable balooproxy
sudo systemctl start balooproxy

# Check status
sudo systemctl status balooproxy
```

---

### Step 13: Set Up Monitoring (Optional but Recommended)

**Install Prometheus:**

```bash
# Download
wget https://github.com/prometheus/prometheus/releases/download/v2.45.0/prometheus-2.45.0.linux-amd64.tar.gz
tar xvfz prometheus-*.tar.gz
cd prometheus-*

# Create config
cat > prometheus.yml <<EOF
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'balooproxy'
    static_configs:
      - targets: ['localhost:9090']
EOF

# Run Prometheus
./prometheus --config.file=prometheus.yml &

# Access UI
echo "Prometheus UI: http://localhost:9090"
```

**Install Grafana (Optional):**

```bash
# Ubuntu/Debian
sudo apt-get install -y software-properties-common
sudo add-apt-repository "deb https://packages.grafana.com/oss/deb stable main"
wget -q -O - https://packages.grafana.com/gpg.key | sudo apt-key add -
sudo apt-get update
sudo apt-get install grafana

# Start Grafana
sudo systemctl start grafana-server
sudo systemctl enable grafana-server

# Access UI (default: admin/admin)
echo "Grafana UI: http://localhost:3000"
```

---

### Step 14: Configure IP Whitelist (Optional)

```bash
# Create whitelist file
nano ipwhitelist.conf
```

**Add IPs (one per line):**
```
# Office IP
203.0.113.10

# Monitoring service
198.51.100.25

# API client
192.0.2.100
```

**Reload proxy:**
```bash
sudo systemctl reload balooproxy
```

---

### Step 15: Verify Everything Works

**Checklist:**

```bash
# 1. Proxy is running
sudo systemctl status balooproxy
# Should show: active (running)

# 2. Ports are listening
sudo netstat -tulpn | grep -E ':(80|443|9090)'
# Should show ports 80, 443, and 9090

# 3. DNS resolves
dig example.com
# Should return your server IP

# 4. HTTPS works
curl -I https://example.com
# Should return 200 OK

# 5. Metrics work
curl http://localhost:9090/metrics | head -20
# Should show Prometheus metrics

# 6. Deduplication works
for i in {1..20}; do curl -s https://example.com/test & done
curl http://localhost:9090/metrics | grep deduplication_savings
# Should show savings > 0

# 7. No errors in logs
sudo journalctl -u balooproxy -n 50
# Should show no errors
```

---

## 🎯 Post-Installation

### Configure Firewall Rules

Add custom rules to `config.json`:

```json
{
  "domains": [{
    "firewallRules": [
      {
        "expression": "(http.path eq \"/admin\")",
        "action": "3"
      },
      {
        "expression": "(ip.country eq \"CN\" or ip.country eq \"RU\")",
        "action": "+2"
      },
      {
        "expression": "(ip.engine eq \"\")",
        "action": "+1"
      }
    ]
  }]
}
```

### Set Up Alerts

**Prometheus Alertmanager:**

```yaml
# alertmanager.yml
route:
  receiver: 'discord'

receivers:
  - name: 'discord'
    webhook_configs:
      - url: 'YOUR_DISCORD_WEBHOOK_URL'
```

**Alert Rules:**

```yaml
# alerts.yml
groups:
  - name: balooproxy
    rules:
      - alert: HighErrorRate
        expr: rate(balooproxy_requests_blocked_total[5m]) > 100
        annotations:
          summary: "High error rate detected"
      
      - alert: AttackDetected
        expr: balooproxy_attacks_detected_total > 0
        annotations:
          summary: "DDoS attack detected"
```

### Monitor Performance

**Key metrics to watch:**

```bash
# Request rate
curl -s http://localhost:9090/metrics | grep requests_total

# Challenge success rate
curl -s http://localhost:9090/metrics | grep challenge_success

# Backend latency
curl -s http://localhost:9090/metrics | grep backend_request_duration

# Deduplication savings
curl -s http://localhost:9090/metrics | grep deduplication_savings
```

---

## 🐛 Troubleshooting

### Issue: Proxy won't start

```bash
# Check logs
tail -f crash.log
sudo journalctl -u balooproxy -f

# Common causes:
# - Port 80/443 already in use
# - Invalid config.json
# - Missing certificates
# - Secrets still contain CHANGE_ME
```

### Issue: Metrics not showing

```bash
# Test metrics endpoint
curl http://localhost:9090/metrics

# If connection refused:
# - Check if proxy is running
# - Verify port 9090 is not blocked
# - Check config.json metrics.enabled = true
```

### Issue: Deduplication not working

```bash
# Verify enabled
cat config.json | grep -A 3 deduplication

# Test with identical requests
for i in {1..20}; do curl -s http://example.com/api/test & done

# Check metrics
curl http://localhost:9090/metrics | grep deduplication

# If savings = 0:
# - Requests might have cookies/auth
# - Requests might not be identical
# - Requests might not be concurrent
```

### Issue: SSL certificate errors

```bash
# Verify certificate files exist
ls -la /path/to/certificate.crt
ls -la /path/to/private.key

# Test certificate
openssl x509 -in certificate.crt -text -noout

# Check certificate matches domain
openssl x509 -in certificate.crt -noout -subject
```

---

## ✅ Final Checklist

**Before going to production:**

- [ ] All dependencies installed (`go mod tidy` successful)
- [ ] Proxy builds without errors
- [ ] Config.json has no CHANGE_ME values
- [ ] SSL certificates configured and valid
- [ ] DNS points to server
- [ ] Backend is accessible from proxy
- [ ] Firewall allows ports 80, 443, 9090
- [ ] Service file configured
- [ ] Proxy starts successfully
- [ ] HTTPS works from external network
- [ ] Metrics endpoint accessible
- [ ] Deduplication showing savings
- [ ] No errors in logs
- [ ] Monitoring set up (Prometheus/Grafana)
- [ ] Alerts configured
- [ ] Backup plan ready
- [ ] Rollback procedure documented

---

## 📊 Verify Installation Success

Run this comprehensive test:

```bash
#!/bin/bash

echo "=== balooProxyExtended Installation Verification ==="
echo ""

# Test 1: Proxy running
echo "1. Checking if proxy is running..."
if systemctl is-active --quiet balooproxy; then
    echo "   ✅ Proxy is running"
else
    echo "   ❌ Proxy is not running"
fi

# Test 2: Ports listening
echo "2. Checking ports..."
if netstat -tuln | grep -q ':80 '; then
    echo "   ✅ Port 80 listening"
else
    echo "   ❌ Port 80 not listening"
fi

if netstat -tuln | grep -q ':443 '; then
    echo "   ✅ Port 443 listening"
else
    echo "   ❌ Port 443 not listening"
fi

if netstat -tuln | grep -q ':9090 '; then
    echo "   ✅ Port 9090 listening (metrics)"
else
    echo "   ❌ Port 9090 not listening"
fi

# Test 3: Metrics endpoint
echo "3. Checking metrics endpoint..."
if curl -s http://localhost:9090/metrics | grep -q balooproxy; then
    echo "   ✅ Metrics endpoint working"
else
    echo "   ❌ Metrics endpoint not working"
fi

# Test 4: Configuration
echo "4. Checking configuration..."
if grep -q "CHANGE_ME" config.json; then
    echo "   ❌ Config still contains CHANGE_ME values"
else
    echo "   ✅ Config properly configured"
fi

# Test 5: Deduplication
echo "5. Testing deduplication..."
for i in {1..10}; do curl -s http://localhost/test &>/dev/null & done
wait
sleep 2
if curl -s http://localhost:9090/metrics | grep -q deduplication_savings; then
    SAVINGS=$(curl -s http://localhost:9090/metrics | grep deduplication_savings_total | awk '{print $2}' | head -1)
    echo "   ✅ Deduplication working (savings: $SAVINGS)"
else
    echo "   ⚠️  Deduplication metrics not found"
fi

echo ""
echo "=== Verification Complete ==="
```

Save as `verify.sh`, make executable, and run:

```bash
chmod +x verify.sh
./verify.sh
```

---

## 🎉 Success!

If all checks pass, your balooProxyExtended installation is complete and ready for production!

### Next Steps:

1. **Monitor metrics** - Watch `http://localhost:9090/metrics`
2. **Set up Grafana** - Create dashboards
3. **Configure alerts** - Get notified of attacks
4. **Load test** - Verify performance
5. **Document** - Save your configuration

### Useful Commands:

```bash
# View logs
sudo journalctl -u balooproxy -f

# Restart proxy
sudo systemctl restart balooproxy

# Reload config
# (use 'reload' command in terminal or restart service)

# Check metrics
curl http://localhost:9090/metrics | grep balooproxy

# Monitor in real-time
watch -n 1 'curl -s http://localhost:9090/metrics | grep requests_total'
```

---

## 📚 Further Reading

- **[NEW_FEATURES.md](NEW_FEATURES.md)** - Complete feature documentation
- **[OPTIMIZATIONS.md](OPTIMIZATIONS.md)** - Performance details
- **[README.md](README.md)** - Main documentation

---

## 🆘 Need Help?

1. **Check logs**: `tail -f crash.log` or `sudo journalctl -u balooproxy -f`
2. **Verify config**: `cat config.json | jq .`
3. **Test components**: Test backend, DNS, certificates separately
4. **Read docs**: Check all `.md` files
5. **Check metrics**: Look for anomalies in metrics

---

**Congratulations on your successful installation!** 🎉

Your balooProxyExtended is now protecting your infrastructure with:
- ✅ Advanced DDoS protection
- ✅ Real-time metrics & monitoring
- ✅ Intelligent request deduplication
- ✅ Optimized performance

**Stay safe and monitor your metrics!** 📊
