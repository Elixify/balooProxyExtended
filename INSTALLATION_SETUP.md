# Installation & Setup Guide - balooProxyExtended

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Installation](#installation)
3. [Configuration](#configuration)
4. [Running the Proxy](#running-the-proxy)
5. [DNS Setup](#dns-setup)
6. [SSL/TLS Certificates](#ssltls-certificates)
7. [New Features Setup](#new-features-setup)
8. [Monitoring Setup](#monitoring-setup)
9. [Production Deployment](#production-deployment)
10. [Troubleshooting](#troubleshooting)

---

## Prerequisites

### System Requirements

**Minimum:**
- CPU: 2 cores
- RAM: 2GB
- Disk: 1GB free space
- OS: Linux, macOS, or Windows

**Recommended:**
- CPU: 4+ cores
- RAM: 4GB+
- Disk: 5GB+ free space
- OS: Linux (Ubuntu 20.04+, Debian 11+, CentOS 8+)

### Software Requirements

- **Go 1.19 or higher** - [Download](https://golang.org/dl/)
- **Git** - For cloning the repository
- **SSL Certificates** - For HTTPS support
- **Optional**: Docker, Prometheus, Grafana

---

## Installation

### Method 1: Build from Source (Recommended)

#### Step 1: Install Go

**Linux:**
```bash
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
```

**macOS:**
```bash
brew install go
```

**Windows:**
Download and install from [golang.org/dl](https://golang.org/dl/)

#### Step 2: Clone Repository

```bash
git clone https://github.com/YOUR_USERNAME/balooProxyExtended.git
cd balooProxyExtended
```

#### Step 3: Install Dependencies

```bash
# Install all dependencies including new features
go get github.com/prometheus/client_golang@v1.19.0
go mod tidy
```

#### Step 4: Build

```bash
# Standard build
go build -o main

# Optimized build (recommended for production)
go build -ldflags="-s -w" -o main

# Build with race detector (development only)
go build -race -o main
```

#### Step 5: Verify Installation

```bash
./main --help
```

You should see the proxy start and prompt for configuration.

### Method 2: Docker

#### Step 1: Build Docker Image

```bash
docker build -t balooproxy-extended .
```

#### Step 2: Run Container

```bash
docker run -d \
  -p 80:80 \
  -p 443:443 \
  -p 9090:9090 \
  -v $(pwd)/config.json:/app/config.json \
  -v $(pwd)/assets:/app/assets \
  --name balooproxy \
  balooproxy-extended
```

---

## Configuration

### Step 1: Generate Configuration

On first run, the proxy will guide you through configuration:

```bash
./main
```

Answer the prompts to generate `config.json`.

### Step 2: Edit Configuration

Copy the example configuration with new features:

```bash
cp examples/config-with-features.json config.json
```

### Step 3: Configure Secrets

**IMPORTANT**: Change all `CHANGE_ME` values!

```json
{
  "proxy": {
    "adminsecret": "YOUR_RANDOM_STRING_HERE",
    "apisecret": "YOUR_RANDOM_STRING_HERE",
    "secrets": {
      "captcha": "YOUR_RANDOM_STRING_1",
      "cookie": "YOUR_RANDOM_STRING_2",
      "javascript": "YOUR_RANDOM_STRING_3"
    }
  }
}
```

**Generate secure secrets:**
```bash
# Linux/macOS
openssl rand -hex 32

# Or use online tool
# https://www.random.org/strings/
```

### Step 4: Configure Domains

```json
{
  "domains": [
    {
      "name": "example.com",
      "backend": "127.0.0.1:8080",
      "scheme": "http",
      "certificate": "assets/server/server.crt",
      "key": "assets/server/server.key",
      "bypassStage1": 75,
      "bypassStage2": 250,
      "disableBypassStage3": 100,
      "disableRawStage3": 250,
      "disableBypassStage2": 50,
      "disableRawStage2": 75,
      "stage2Difficulty": 5
    }
  ]
}
```

**Configuration Parameters:**

- `name`: Your domain name
- `backend`: Backend server address (IP:port)
- `scheme`: `http` or `https` (use `http` for better performance)
- `certificate`: Path to SSL certificate
- `key`: Path to SSL private key
- `bypassStage1`: Requests/sec to trigger stage 2
- `bypassStage2`: Requests/sec to trigger stage 3
- `stage2Difficulty`: JS challenge difficulty (1-5)

### Step 5: Configure New Features

Add metrics and deduplication:

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

### Step 6: Configure Rate Limits

```json
{
  "proxy": {
    "ratelimits": {
      "requests": 500,
      "unknownFingerprint": 150,
      "challengeFailures": 40,
      "noRequestsSent": 10
    }
  }
}
```

### Step 7: Configure Firewall Rules (Optional)

```json
{
  "domains": [{
    "firewallRules": [
      {
        "expression": "(http.path eq \"/admin\")",
        "action": "3"
      },
      {
        "expression": "(ip.engine eq \"\")",
        "action": "+1"
      }
    ]
  }]
}
```

---

## SSL/TLS Certificates

### Option 1: Let's Encrypt (Recommended)

```bash
# Install certbot
sudo apt install certbot

# Generate certificate
sudo certbot certonly --standalone -d example.com

# Certificates will be in:
# /etc/letsencrypt/live/example.com/fullchain.pem
# /etc/letsencrypt/live/example.com/privkey.pem
```

Update config.json:
```json
{
  "certificate": "/etc/letsencrypt/live/example.com/fullchain.pem",
  "key": "/etc/letsencrypt/live/example.com/privkey.pem"
}
```

### Option 2: Self-Signed (Testing Only)

```bash
openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.crt -days 365 -nodes
```

### Option 3: Commercial Certificate

Upload your certificate files and update paths in config.json.

---

## Running the Proxy

### Development Mode

```bash
# Run with console output
./main

# Run in daemon mode (no console)
./main -d
```

### Production Mode (Systemd Service)

#### Step 1: Create Service File

```bash
sudo nano /etc/systemd/system/balooproxy.service
```

```ini
[Unit]
Description=balooProxyExtended DDoS Protection
After=network.target

[Service]
Type=simple
User=balooproxy
WorkingDirectory=/opt/balooproxy
ExecStart=/opt/balooproxy/main -d
Restart=always
RestartSec=10

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/balooproxy

[Install]
WantedBy=multi-user.target
```

#### Step 2: Create User and Set Permissions

```bash
# Create user
sudo useradd -r -s /bin/false balooproxy

# Create directory
sudo mkdir -p /opt/balooproxy
sudo cp main config.json /opt/balooproxy/
sudo cp -r assets /opt/balooproxy/
sudo chown -R balooproxy:balooproxy /opt/balooproxy
```

#### Step 3: Set Capabilities (for ports 80/443)

```bash
sudo setcap 'cap_net_bind_service=+ep' /opt/balooproxy/main
```

#### Step 4: Start Service

```bash
sudo systemctl daemon-reload
sudo systemctl enable balooproxy
sudo systemctl start balooproxy
```

#### Step 5: Check Status

```bash
sudo systemctl status balooproxy
sudo journalctl -u balooproxy -f
```

---

## DNS Setup

### Step 1: Get Server IP

```bash
curl ifconfig.me
```

### Step 2: Configure DNS Records

In your DNS provider (Cloudflare, Route53, etc.):

**Without Cloudflare:**
```
Type: A
Name: example.com
Value: YOUR_SERVER_IP
TTL: 300
Proxy: OFF (DNS only)
```

**With Cloudflare:**
```
Type: A
Name: example.com
Value: YOUR_SERVER_IP
TTL: Auto
Proxy: ON (Proxied)
```

Set `"cloudflare": true` in config.json when using Cloudflare.

### Step 3: Verify DNS

```bash
dig example.com
nslookup example.com
```

Wait 5-10 minutes for DNS propagation.

### Step 4: Test Proxy

```bash
curl -I https://example.com
```

Look for `baloo-proxy` header (if stealth mode is off).

---

## New Features Setup

### Metrics & Observability

#### Step 1: Verify Metrics Endpoint

```bash
curl http://localhost:9090/metrics
```

You should see Prometheus metrics.

#### Step 2: Install Prometheus (Optional)

```bash
# Download Prometheus
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
./prometheus --config.file=prometheus.yml
```

Access Prometheus UI at `http://localhost:9090`

#### Step 3: Install Grafana (Optional)

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
```

Access Grafana at `http://localhost:3000` (admin/admin)

### Request Deduplication

Deduplication is automatic once enabled in config. Monitor effectiveness:

```bash
# Check deduplication metrics
curl -s http://localhost:9090/metrics | grep deduplication

# Expected output:
# balooproxy_deduplicated_requests_total{domain="example.com"} 42
# balooproxy_deduplication_savings_total{domain="example.com"} 380
```

---

## IP Whitelist

### Step 1: Create Whitelist File

```bash
nano ipwhitelist.conf
```

### Step 2: Add IPs

```
# Comments start with #
# One IP per line

# Office IP
203.0.113.10

# Monitoring service
198.51.100.25

# API client
192.0.2.100
```

### Step 3: Reload Proxy

```bash
sudo systemctl reload balooproxy
```

Whitelisted IPs bypass all challenges and rate limits.

---

## Production Deployment

### Security Checklist

- [ ] Change all `CHANGE_ME` secrets
- [ ] Use strong, random secrets (32+ characters)
- [ ] Enable firewall (ufw, iptables)
- [ ] Use Let's Encrypt certificates
- [ ] Run as non-root user
- [ ] Set file permissions correctly
- [ ] Enable stealth mode
- [ ] Configure rate limits appropriately
- [ ] Set up monitoring and alerts
- [ ] Test failover procedures

### Performance Tuning

**System Limits:**
```bash
# Increase file descriptors
sudo nano /etc/security/limits.conf
```

Add:
```
* soft nofile 65536
* hard nofile 65536
```

**Kernel Parameters:**
```bash
sudo nano /etc/sysctl.conf
```

Add:
```
net.core.somaxconn = 65536
net.ipv4.tcp_max_syn_backlog = 8192
net.ipv4.ip_local_port_range = 1024 65535
```

Apply:
```bash
sudo sysctl -p
```

### Monitoring

**Key Metrics to Watch:**
- Request rate (total and allowed)
- Challenge success rates
- Backend latency
- Deduplication savings
- CPU and memory usage
- Attack detection events

**Set Up Alerts:**
- High error rate (>5%)
- Attack detected
- Backend down
- High latency (>1s P95)
- Memory usage >80%

---

## Troubleshooting

### Proxy Won't Start

**Check logs:**
```bash
tail -f crash.log
sudo journalctl -u balooproxy -f
```

**Common issues:**
- Port 80/443 already in use
- Invalid config.json syntax
- Missing certificates
- Secrets contain "CHANGE_ME"

### Metrics Not Showing

```bash
# Check if metrics endpoint is accessible
curl http://localhost:9090/metrics

# Check if port is open
sudo netstat -tulpn | grep 9090

# Verify config
cat config.json | grep -A 5 metrics
```

### Deduplication Not Working

```bash
# Verify it's enabled
cat config.json | grep -A 3 deduplication

# Check metrics
curl -s http://localhost:9090/metrics | grep dedup

# Test with concurrent requests
for i in {1..10}; do curl -s http://example.com/test & done
```

### Backend Connection Failed

```bash
# Test backend directly
curl http://127.0.0.1:8080

# Check backend is running
sudo netstat -tulpn | grep 8080

# Verify config
cat config.json | grep backend
```

### SSL Certificate Errors

```bash
# Verify certificate files exist
ls -la /path/to/certificate.crt
ls -la /path/to/private.key

# Check certificate validity
openssl x509 -in certificate.crt -text -noout

# Test SSL
openssl s_client -connect example.com:443
```

### High Memory Usage

```bash
# Check memory usage
ps aux | grep main

# Monitor in real-time
top -p $(pgrep main)

# Reduce deduplication TTL if needed
# Edit config.json: "ttl_seconds": 10
```

---

## Next Steps

1. **Read Documentation**
   - [NEW_FEATURES.md](NEW_FEATURES.md) - Learn about metrics and deduplication
   - [QUICK_START_NEW_FEATURES.md](QUICK_START_NEW_FEATURES.md) - 5-minute guide

2. **Set Up Monitoring**
   - Install Prometheus and Grafana
   - Import pre-built dashboards
   - Configure alerts

3. **Optimize Configuration**
   - Tune rate limits for your traffic
   - Configure firewall rules
   - Adjust challenge difficulty

4. **Test Thoroughly**
   - Load test with realistic traffic
   - Simulate DDoS attacks
   - Verify failover procedures

5. **Monitor in Production**
   - Watch metrics daily
   - Review attack logs
   - Optimize based on data

---

## Support

- **Documentation**: Check all `.md` files in the repository
- **Logs**: `crash.log` and `journalctl -u balooproxy`
- **Metrics**: `http://localhost:9090/metrics`
- **Community**: GitHub Issues

---

**Congratulations!** Your balooProxyExtended installation is complete. 🎉

For questions or issues, check the logs and documentation first.
