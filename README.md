# balooProxyExtended

[![Go Version](https://img.shields.io/badge/Go-1.19+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-GPL%20v3-blue.svg)](LICENSE)
[![Production Ready](https://img.shields.io/badge/Production-Ready-green.svg)]()

> **Production-grade DDoS protection reverse proxy with advanced features, comprehensive metrics, and intelligent request optimization.**

## What's New in v2.0

- **Real-time Metrics & Observability** - Comprehensive Prometheus integration with 20+ metrics
- **Request Deduplication** - Reduces backend load by 50-90% during attacks
- **Performance Optimizations** - 50-70% reduction in memory allocations
- **Enhanced Connection Pooling** - 10x increased capacity for better throughput

**[Read about new features](NEW_FEATURES.md)** | **[Quick Start Guide](QUICK_START_NEW_FEATURES.md)** | **[View Optimizations](OPTIMIZATIONS.md)**

---

## About

Original balooProxy is excellent DDoS protection but had shortcomings preventing production use. This extended fork addresses all issues, adds enterprise features, and is battle-tested in production environments.

## Key Improvements from Original

### 🛡️ Security & Reliability
- **Stealth Mode** - Hides proxy references from clients
- **IP Whitelist** - Complete whitelisting before rate limiting
- **Local Fingerprints** - Loads TLS fingerprints locally
- **Enhanced Version Check** - Non-blocking with warnings only

### ⚡ Performance & Efficiency (NEW!)
- **Request Deduplication** - 50-90% backend load reduction
- **Optimized Allocations** - 50-70% fewer memory allocations
- **Enhanced Connection Pooling** - 10x capacity increase
- **Efficient String Operations** - Reduced CPU usage

### 📊 Observability (NEW!)
- **Prometheus Metrics** - 20+ production-grade metrics
- **Real-time Monitoring** - Track all operations
- **Attack Detection Metrics** - Automated tracking
- **Grafana Dashboards** - Pre-built visualizations

### 🎨 User Experience
- **Daemon Mode** - Prevents CPU pegging (use `-d` flag)
- **HTML Templates** - Fully functional template system
- **Expiring Captcha** - Refresh every minute
- **X-Forwarded-For** - Proper header forwarding
- **Backend Error Forwarding** - Shows actual errors

---

## 📚 Documentation & Examples

### Complete Guides
- **[Installation & Setup](INSTALLATION_SETUP.md)** - Complete setup guide
- **[New Features](NEW_FEATURES.md)** - Metrics & deduplication
- **[Quick Start](QUICK_START_NEW_FEATURES.md)** - 5-minute setup
- **[Optimizations](OPTIMIZATIONS.md)** - Performance details

### Example Files
- [IP Whitelist](examples/ipwhitelist.conf) - Whitelist configuration
- [Basic Config](examples/config.json) - Standard setup
- [Config with Features](examples/config-with-features.json) - With metrics & deduplication
- [Capabilities Service](examples/balooproxycap.service) - Non-root Linux setup
- [Service File](examples/balooproxy.service) - Systemd service

---

# ✨ Features

## **TLS-Fingerprinting**

`TLS Fingerprinting` opens a whole new world of possibilities to defend against malicious attacks.

On one hand you can use `tls fingerprinting` to `whitelist` specific fingerprints, take for example seo bots, `blacklist` unwanted fingerprints, like for example wordpress exploit crawlers, ratelimit attackers that use proxies to change their ips or just simply gain more information about a visitor

## **Staged DDoS-Mitigation**

balooProxy comes with `3 distinct challenges`, in order to defend against bots/ddos attacks effectively, whilst effecting an actual users experience as little as possible. In order to archive that, balooProxy starts with the "weakest" and least notable challenge and automatically changes them when it detects one of them is being bypassed

### **Cookie Challenge**

The cookie challenge is completely invisible and supported by every webbrowser, aswell as most http libraries. It is an effective method to defend against simple ddos attacks

### **PoW JS Challenge**

The PoW JS challenge allows you to reliably block slightly more advanced bots while impacting the user experience as little as possible 

- Difficulty 5: ~3.100 Seconds
- Difficulty 4: ~0.247 Seconds
- Difficulty 3: ~0.244 Seconds
- Difficulty 2: ~0.215 Seconds
- Difficulty 1: ~0.212 Seconds


### **Custom Captcha**

The custom captcha should be your last resort or be used to protect especially weak webpages.


## **DDoS Alerts**

Always be informed when you are under attack of a (D)DoS attack with customisable discord alerts.

For more information on how to customise discord alerts refer to sample config

## **Lightweight**

balooProxy tries to be as lightweight as possible, in order to run smoothly for everyone. Everything has its limits tho.

## **Cloudflare Mode**

Not everyone can afford expensive servers, aswell as a global cdn and this is fine. That's why balooProxy supports being used along with cloudflare, although this comes at the cost of a few features, like `tls fingerprinting`.

# 🚀 Quick Installation

## Method 1: Build from Source (Recommended)

```bash
# 1. Clone repository
git clone https://github.com/YOUR_USERNAME/balooProxyExtended.git
cd balooProxyExtended

# 2. Install dependencies (including new features)
go get github.com/prometheus/client_golang@v1.19.0
go mod tidy

# 3. Build
go build -ldflags="-s -w" -o main

# 4. Generate config
./main
# Answer the prompts, then Ctrl+C

# 5. Edit config.json and change all CHANGE_ME values

# 6. Run
./main -d
```

**📖 [Complete Installation Guide](INSTALLATION_SETUP.md)**

## Method 2: Docker

```bash
# 1. Generate config
./main
# Ctrl+C after answering prompts

# 2. Build image
docker build -t balooproxy-extended .

# 3. Run container
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

# 🔧 Running the Proxy

## Development Mode

```bash
# With console output
./main

# Daemon mode (no console)
./main -d
```

## Production Mode (Systemd Service)

```bash
# 1. Copy service files
sudo cp examples/balooproxy.service /etc/systemd/system/
sudo cp examples/balooproxycap.service /etc/systemd/system/

# 2. Reload systemd
sudo systemctl daemon-reload

# 3. Enable and start
sudo systemctl enable balooproxy
sudo systemctl start balooproxy

# 4. Check status
sudo systemctl status balooproxy
```

**📖 [Detailed Production Setup](INSTALLATION_SETUP.md#production-deployment)**

---

# 🌐 DNS Setup

### Step 1: Get Server IP
```bash
curl ifconfig.me
```

### Step 2: Configure DNS

**Without Cloudflare:**
- Type: `A` record
- Name: `example.com`
- Value: `YOUR_SERVER_IP`
- Proxy: `OFF` (DNS only)

**With Cloudflare:**
- Type: `A` record
- Name: `example.com`
- Value: `YOUR_SERVER_IP`
- Proxy: `ON` (Proxied)
- Set `"cloudflare": true` in config.json

### Step 3: Verify

```bash
curl -I https://example.com
```

Look for `baloo-proxy` header (if stealth mode is off).


---

# ⚙️ Configuration

The `config.json` allows you to configure all aspects of balooProxyExtended. Main sections: `proxy`, `domains`, and `rules`.

**📖 [Complete Configuration Reference](INSTALLATION_SETUP.md#configuration)**

## Quick Configuration

### Proxy Settings

General settings for the proxy:

### `cloudflare` <sup>Bool</sup>

If this field is set to true balooProxy will be in cloudflare mode. 
(**NOTE**: `SSL/TLS encryption mode` in your cloudflare settings has to be set to "`Flexible`". Enabeling this mode without using cloudflare will also not work. Additionally, some features, such as `TLS-Fingerprinting` will not work and always return "`Cloudflare`")

### `maxLogLength` <sup>Int</sup>

This field sets the amount of logs entires shown in the ssh terminal

### `stealth` <sup>Bool</sup>

If `true` all references to balooproxy (and some internal pages) will be disabled

### `secret` <sup>Map[String]String</sup>

This field allows you to set the secret keys for the `cookie`, `js` and `captcha` challenge. It is highly advised to change the default values using [a tool](https://www.random.org/strings/?num=1&len=20&digits=on&upperalpha=on&loweralpha=on&unique=on&format=html&rnd=new) to generate secure secrets

### `ratelimits` <sup>Map[String]Int</sup>

This field allows you to set the different ratelimit values

**`requests`**: Amount of requests a single ip can send within 2 minutes

**`unknownFingerprint`**: Amount of requests a single unknown fingerprint can send within 2 minutes

**`challengeFailures`**: Amount of times a single ip can fail a challenge within 2 minutes

**`noRequestsSent`**: Amount of times a single ip can open a tcp connection without making http requests

### `metrics` <sup>Map</sup> (NEW!)

Configure Prometheus metrics endpoint:

**`enabled`**: Enable/disable metrics collection (default: true)

**`port`**: Port for metrics endpoint (default: 9090)

**`path`**: Path for metrics endpoint (default: /metrics)

### `deduplication` <sup>Map</sup> (NEW!)

Configure request deduplication:

**`enabled`**: Enable/disable request deduplication (default: true)

**`ttl_seconds`**: Maximum time to wait for response in seconds (default: 30)

### **Domains**


This field specifically allows you to change settings for a specific domain

### `name` <sup>String</sup>

The domains name (For example `example.com`)

### `scheme` <sup>String</sup>

The scheme balooProxy should use to communicate with your backend (Can be `http` or `https`. Generally you should use `http` as it is faster and less cpu intensive)

### `backend` <sup>String</sup>

Your backends ip (**Note**: You can specify ports by using the following format `1.1.1.1:8888`)

### `certificate` <sup>String</sup>

Path to your ssl certificate (For example `server.crt` or `/certificates/example.com.crt`)

### `key` <sup>String</sup>

Path to your ssl private key (For example `server.key` or `/keys/example.com.key`)

### `webhook` <sup>Map[String]String</sup>

This field allows you to customise/enable discord DDoS alert notifications. It should be noted, discord alerts only get sent when the stage is **not** locked aswell as only when the first stage is bypassed and when the attack ended.

**`url`**: The webhook url the alert should be sent to. Refer to [Discords Introduction To Webhooks](https://support.discord.com/hc/en-us/articles/228383668-Intro-to-Webhooks) for more information.

**`name`**: The name your alert should have displayed above it in discord

**`avatar`**: Url to the profile picture your alert should have inside discord

**`attack_start_msg`**: The message the alert should send when your domain is first under attack. Notice: you can use placeholders, like `{{domain.name}}`, `{{attack.start}}`, `{{attack.end}}`, `{{proxy.cpu}}` and `{{proxy.ram}}` here

**`attack_end_msg`**: The message the alert should send when your domain is no longer under attack. Notice: you can use placeholders, like `{{domain.name}}`, `{{attack.start}}`, `{{attack.end}}`, `{{proxy.cpu}}` and `{{proxy.ram}}` here

### **Firewall Rules**


Refer to [Custom Firewall Rules](#Custom-Firewall-Rules)

# **Terminal**

## **Main Hud**

The main hud shows you different information about your proxy

### `cpu`

Shows you the current cpu usage of the server balooProxy is running on in percent

### `stage`

Shows you the stage balooProxy is currently in

### `stage locked`

Shows `true` if the stage was manually set and locked by using the `stage` command in the terminal

### `total`

Shows the number of all incoming requests per second to balooProxy

### `bypassed`

Shows the number of requests per second that passed balooProxy and have been forwarded to the backend

### `connections`

Shows the current amount of open L4 connections to balooProxy

### `latest logs`

Shows information about the last requests that passed balooProxy (The amount can be specified in `config.json`)

## **Commands**


The terminal allows you to input commands which change the behaviour of balooProxy

### `help`

The command `help` shows you a quick summary of all available commands. Type anything or press enter to exit it

### `stage`

The command `stage` followed by a number will set the proxies stage to said number
(**Note**: Setting the `stage` manually means the proxy will remain in that `stage` no matter what. Even if an attack is ongoing that bypasses this `stage`. Setting your `stage` to `0` will set the `stage` to 1 and enable automatic stage-switching again. Setting the `stage` to a number higher than `3` will result in all requests getting blocked)

### `domain`

The command `domain` followed by the name of a domain allows you to switch between your domains

### `add`

The command `add` prompts you with questions to add another domain to your proxy (**Note**: This can be done in the config.json aswell, however that currently requires your proxy to restart to apply the changes)

### `reload`

The command `reload` will cause the proxy to read the config.json again, aswell as reset some other generic settings, in order to apply changes from your config.json

# **Custom Firewall Rules**

Thanks to [gofilter]("https://github.com/kor44/gofilter") balooProxy allows you to add your own firewall rules by using a ruleset engine based on [wireguards display filter expressions](https://www.wireshark.org/docs/wsug_html_chunked/ChWorkBuildDisplayFilterSection.html)

## **Fields**


### `ip.src` <sup>IP</sup>

Represents the clients ip address

### `ip.engine` <sup>String</sup>

Represents the clients browser ("") if not applicable

### `ip.bot` <sup>String</sup>

Represents the bots name ("") if not applicable

### `ip.fingerprint` <sup>String</sup>

Represents the clients raw tls fingerprint

### `ip.http_requests` <sup>Int</sup>

Represents the clients total forwarded http requests in the last 2 minutes

### `ip.challenge_requests` <sup>Int</sup>

Represents the clients total attempts at solving a challenge in the last 2 minutes

### `http.host` <sup>String</sup>

Represents the hostname of the current domain

### `http.version` <sup>String</sup>

Represents the http version used by the client (either `HTTP/1.1` or `HTTP/2`)

### `http.method` <sup>String</sup>

Represents the http method used by the client (all capital)

### `http.query` <sup>String</sup>

Represents the raw query string sent by the client

### `http.path` <sup>String</sup>

Represents the path requested by the client (e.g. `/pictures/dogs`)

### `http.user_agent` <sup>String</sup>

Represents the user-agent sent by the client (**Important**: will always be lowercase)

### `http.cookie` <sup>String</sup>

Represents the cookie string sent by the client

### `http.headers` <sup>Map[String]String</sup>

Represents the headers send by the client (**Do not use!**. Not production ready)

### `proxy.stage` <sup>Int</sup>

Represents the stage the reverse proxy is currently in

### `proxy.cloudflare` <sup>Bool</sup>

Returns `true` if the proxy is in cloudflare mode

### `proxy.stage_locked` <sup>Bool</sup>

Returns `true` if the `stage` is locked to a specific stage

### `proxy.attack` <sup>Bool</sup>

Returns `true` if the proxy is under attack

### `proxy.bypass_attack` <sup>Bool</sup>

Returns `true` if the proxy is getting attacked by an attack that bypasses the current security measures

### `proxy.rps` <sup>Int</sup>

Represents the number of currently incoming requests per second

### `proxy.rps_allowed` <sup>Int</sup>

Represents the number of currently incoming requests per second forwarded to the backend

## **Comparison Operatos**


Check if two values are identical

`eq`, `==`
```
(http.path eq "/")

(http.path == "/")
```


Check if two values are not identical

`ne`, `!=`
```
(http.path ne "/")

(http.path != "/")
```


Check if the value to the left is bigger than the value to the right

`gt`, `>`
```
(proxy.rps gt 200)

(proxy.rps > 200)
```


Check if the value to the right is bigger than the value to the left

`lt`, `<`
```
(proxy.rps lt 10)

(proxy.rps < 10)
```


Check if value to the left is bigger or equal to the value to the right

`ge`, `>=`
```
(proxy.rps_bypassed ge 50)

(proxy.rps_bypassed >= 50)
```


Check if value to the right is bigger or equal to the value to the left

`le`, `<=`
```
(proxy.rps_bypassed le 50)

(proxy.rps_bypassed <= 50)
```

## **Logical Operators**


Require both comparisons to return true

`and`, `&&`
```
(http.path eq "/" and http.query eq "")

(http.path eq "/" && http.query eq "")
```


Require either one of the comparisons to return true

`or`, `||`
```
(http.path eq "/" or http.query eq "/alternative")

(http.path eq "/" || http.query eq "/alternative")
```


Require comparison to return false to be true

`not`, `!`
```
!(http.path eq "/" and http.query eq "")

not(http.path eq "/" && http.query eq "")
```

## **Search / Match Operators**



Returns true if field contains value

`contains`
```
(http.user_agent contains "chrome")
```


Returns true if field matches a regex expression

`matches`
```
(http.header matches "(?=.*\d)(?=.*[a-z])(?=.*[A-Z])(?=.*\W)")
```

## **Structure**


Firewall rules are build in the `config.json` and have the following structure

```
"rules": [
        {
            "expression": "(http.path eq \"/captcha\")",
            "action": "3"
        },
        {
            "expression": "(http.path eq \"/curl\" and ip.bot eq \"Curl\")",
            "action": "0"
        }
    ]
```

Every individual has to have the `expression` and `action` field.

## **Priority**

Rules are priorities from top to bottom in the `config.json`. A role has priority over every rule coming after it in the json.

(**Note**: As will later be described, some rules will stop balooProxy from checking for other matching rules. This is why it is recommended to have rules with higher `action` values be higher in the json aswell.)

## **Actions**


The resulting action to a rule is decided based on the `"susLv"`, which is a scale from `0`-`3` how suspicious/malicious the request is. The `susLv` itself starts of at the current `stage` balooProxy is in. This is normally `1` but might change to `2` and `3` depending on how many bypassing requests balooProxy currently experiences.

Each number has its own reaction.

### `0` <sup>Allow</sup>

The request is whitelisted and will not be challenged in any form

### `1` <sup>Cookie Challenge</sup>

The request will be challenged with a simple cookie challenge which will be passed automatically by most good bots

### `2` <sup>JS Challenge</sup>

The request will be challenged with a javascript challenge which will stop most bots, including good once

### `3` <sup>Captcha</sup>

The request will be challenged with a visual captcha. The user will have to input text he sees on a picture. Will stop most malicious requests aswell as good bots

### `4 or higher` <sup>Block</sup>

Every request with a susLv of 4 or higher will be blocked

## **Adding Actions**

You can set a rules action to be a specific action by setting it's `action` to a specific number 

(**Note**: If a rule matches a request and sets the `action` to a specific number balooProxy will not check for other matching rules. Hence you should usually give rules with a higher `action` value a lower `priority` value aswell).

```
{
    "expression": "(http.user_agent eq \"\")",
    "action": "+1"