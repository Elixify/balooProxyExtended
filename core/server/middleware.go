package server

import (
	"bytes"
	"encoding/base64"
	"goProxy/core/api"
	"goProxy/core/dedup"
	"goProxy/core/domains"
	"goProxy/core/firewall"
	"goProxy/core/metrics"
	"goProxy/core/pnc"
	"goProxy/core/proxy"
	"goProxy/core/utils"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kor44/gofilter"
)

var (
	// Request deduplicator
	requestDedup *dedup.Deduplicator
	
	// String builder pool for efficient string concatenation
	stringBuilderPool = sync.Pool{
		New: func() interface{} {
			return &strings.Builder{}
		},
	}
)

func SendResponse(str string, buffer *bytes.Buffer, writer http.ResponseWriter) {
	buffer.WriteString(str)
	writer.Write(buffer.Bytes())
}

func Middleware(writer http.ResponseWriter, request *http.Request) {

	// Panic recovery is essential under high load to prevent crashes
	defer pnc.PanicHndl()
	
	// Track request timing for metrics (used when metrics are fully integrated)
	_ = time.Now() // startTime for future metrics integration

	buffer := bufferPool.Get().(*bytes.Buffer)
	buffer.Reset()
	defer bufferPool.Put(buffer)

	domainName := request.Host

	firewall.Mutex.RLock()
	domainData, domainFound := domains.DomainsData[domainName]
	firewall.Mutex.RUnlock()

	if !domainFound {
		writer.Header().Set("Content-Type", "text/plain")
		SendResponse("404 Not Found", buffer, writer)
		metrics.RecordBlockedRequest(domainName, "domain_not_found")
		return
	}

	var ip string
	var tlsFp string
	var browser string
	var botFp string

	var fpCount int
	var ipCount int
	var ipCountCookie int

	if domains.Config.Proxy.Cloudflare {
		ip = request.Header.Get("Cf-Connecting-Ip")
		tlsFp = "Cloudflare"
		browser = "Cloudflare"
		botFp = ""
		fpCount = 0

		firewall.Mutex.RLock()
		ipCount = firewall.AccessIps[ip]
		ipCountCookie = firewall.AccessIpsCookie[ip]
		firewall.Mutex.RUnlock()
	} else {
		// Optimize IP extraction - avoid allocation from Split
		if idx := strings.LastIndexByte(request.RemoteAddr, ':'); idx != -1 {
			ip = request.RemoteAddr[:idx]
		} else {
			ip = request.RemoteAddr
		}

		//Retrieve information about the client
		firewall.Mutex.RLock()
		tlsFp = firewall.Connections[request.RemoteAddr]
		fpCount = firewall.UnkFps[tlsFp]
		ipCount = firewall.AccessIps[ip]
		ipCountCookie = firewall.AccessIpsCookie[ip]
		firewall.Mutex.RUnlock()

		//Read-Only IMPORTANT: Must be put in mutex if you add the ability to change indexed fingerprints while program is running
		browser = firewall.KnownFingerprints[tlsFp]
		botFp = firewall.BotFingerprints[tlsFp]
	}

	var settingsQuery any
	var domainSettings domains.DomainSettings

	// Real IP whitelist
	if _, exists := proxy.IPWhitelist[ip]; exists {
		// Init these later if not in whitelist
		settingsQuery, _ = domains.DomainsMap.Load(domainName)
		domainSettings = settingsQuery.(domains.DomainSettings)
		ForwardRequest(writer, request, domainSettings, ip, tlsFp, browser, botFp)
		return
	}

	// Optimize: Minimize lock time for high throughput
	firewall.Mutex.Lock()
	firewall.WindowAccessIps[proxy.Last10SecondTimestamp][ip]++
	domainData = domains.DomainsData[domainName]
	domainData.TotalRequests++
	domains.DomainsData[domainName] = domainData
	firewall.Mutex.Unlock()

	// Stealth code
	var blockTxt = "Blocked.\n"
	var nameTxt = ""
	if !proxy.Stealth {
		writer.Header().Set("baloo-Proxy", strconv.FormatFloat(proxy.ProxyVersion, 'f', 2, 64))
		blockTxt = "Blocked by BalooProxyX.\n"
		nameTxt = "BalooProxyX "
	}

	//Start the suspicious level where the stage currently is
	susLv := domainData.Stage

	//Ratelimit faster if client repeatedly fails the verification challenge (feel free to play around with the threshhold)
	if ipCountCookie > proxy.FailChallengeRatelimit {
		writer.Header().Set("Content-Type", "text/plain")
		SendResponse(blockTxt+"You have been ratelimited. (R1)", buffer, writer)
		return
	}

	//Ratelimit spamming Ips (feel free to play around with the threshhold)
	if ipCount > proxy.IPRatelimit {
		writer.Header().Set("Content-Type", "text/plain")
		SendResponse(blockTxt+"You have been ratelimited. (R2)", buffer, writer)
		return
	}

	//Ratelimit fingerprints that don't belong to major browsers
	if browser == "" {
		if fpCount > proxy.FPRatelimit {
			writer.Header().Set("Content-Type", "text/plain")
			SendResponse(blockTxt+"You have been ratelimited. (R3)", buffer, writer)
			return
		}

		firewall.Mutex.Lock()
		firewall.WindowUnkFps[proxy.Last10SecondTimestamp][tlsFp]++
		firewall.Mutex.Unlock()
	}

	//Block user-specified fingerprints
	forbiddenFp := firewall.ForbiddenFingerprints[tlsFp]
	if forbiddenFp != "" {
		writer.Header().Set("Content-Type", "text/plain")
		SendResponse(blockTxt+"Your browser "+forbiddenFp+" is not allowed.", buffer, writer)
		return
	}

	//Demonstration of how to use "susLv". Essentially allows you to challenge specific requests with a higher challenge

	// Setting these after initial check
	//SyncMap because semi-readonly
	settingsQuery, _ = domains.DomainsMap.Load(domainName)
	domainSettings = settingsQuery.(domains.DomainSettings)

	ipInfoCountry := "N/A"
	ipInfoASN := "N/A"
	if domainSettings.IPInfo {
		ipInfoCountry, ipInfoASN = utils.GetIpInfo(ip)
	}

	reqUa := request.UserAgent()

	if len(domainSettings.CustomRules) != 0 {
		requestVariables := gofilter.Message{
			"ip.src":                net.ParseIP(ip),
			"ip.country":            ipInfoCountry,
			"ip.asn":                ipInfoASN,
			"ip.engine":             browser,
			"ip.bot":                botFp,
			"ip.fingerprint":        tlsFp,
			"ip.http_requests":      ipCount,
			"ip.challenge_requests": ipCountCookie,

			"http.host":       domainName,
			"http.version":    request.Proto,
			"http.method":     request.Method,
			"http.url":        request.RequestURI,
			"http.query":      request.URL.RawQuery,
			"http.path":       request.URL.Path,
			"http.user_agent": strings.ToLower(reqUa),
			"http.cookie":     request.Header.Get("Cookie"),

			"proxy.stage":         domainData.Stage,
			"proxy.cloudflare":    domains.Config.Proxy.Cloudflare,
			"proxy.stage_locked":  domainData.StageManuallySet,
			"proxy.attack":        domainData.RawAttack,
			"proxy.bypass_attack": domainData.BypassAttack,
			"proxy.rps":           domainData.RequestsPerSecond,
			"proxy.rps_allowed":   domainData.RequestsBypassedPerSecond,
		}

		susLv = firewall.EvalFirewallRule(domainSettings, requestVariables, susLv)
	}

	//Check if encryption-result is already "cached" to prevent load on reverse proxy
	encryptedIP := ""
	hashedEncryptedIP := ""
	susLvStr := utils.StageToString(susLv)
	// Optimize string concatenation with pooled strings.Builder
	keyBuilder := stringBuilderPool.Get().(*strings.Builder)
	keyBuilder.Reset()
	defer stringBuilderPool.Put(keyBuilder)
	
	keyBuilder.Grow(len(ip) + len(tlsFp) + len(reqUa) + len(proxy.CurrHourStr) + len(susLvStr))
	keyBuilder.WriteString(ip)
	keyBuilder.WriteString(tlsFp)
	keyBuilder.WriteString(reqUa)
	keyBuilder.WriteString(proxy.CurrHourStr)
	accessKey := keyBuilder.String()
	keyBuilder.WriteString(susLvStr)
	cacheKey := keyBuilder.String()
	encryptedCache, encryptedExists := firewall.CacheIps.Load(cacheKey)

	if !encryptedExists {
		switch susLv {
		case 0:
			//whitelisted
		case 1:
			encryptedIP = utils.Encrypt(accessKey, proxy.CookieOTP)
		case 2:
			encryptedIP = utils.Encrypt(accessKey, proxy.JSOTP)
			hashedEncryptedIP = utils.EncryptSha(encryptedIP, "")
			firewall.CacheIps.Store(encryptedIP, hashedEncryptedIP)
		case 3:
			encryptedIP = utils.Encrypt(accessKey, proxy.CaptchaOTP)
		default:
			writer.Header().Set("Content-Type", "text/plain")
			SendResponse(blockTxt+"Suspicious request of level "+susLvStr+" (base "+strconv.Itoa(domainData.Stage)+")", buffer, writer)
			return
		}
		firewall.CacheIps.Store(cacheKey, encryptedIP)
	} else {
		encryptedIP = encryptedCache.(string)
		cachedHIP, foundCachedHIP := firewall.CacheIps.Load(encryptedIP)
		if foundCachedHIP {
			hashedEncryptedIP = cachedHIP.(string)
		}
	}

	//Check if client provided correct verification result
	if !strings.Contains(request.Header.Get("Cookie"), "__bProxy_v="+encryptedIP) {

		firewall.Mutex.Lock()
		firewall.WindowAccessIpsCookie[proxy.Last10SecondTimestamp][ip]++
		firewall.Mutex.Unlock()

		//Respond with verification challenge if client didnt provide correct result/none
		switch susLv {
		case 0:
			//This request is not to be challenged (whitelist)
		case 1:
			writer.Header().Set("Set-Cookie", "_1__bProxy_v="+encryptedIP+"; SameSite=Lax; path=/; Secure")
			http.Redirect(writer, request, request.URL.RequestURI(), http.StatusFound)
			return
		case 2:
			publicSalt := encryptedIP[:len(encryptedIP)-domainData.Stage2Difficulty]
			writer.Header().Set("Content-Type", "text/html")
			writer.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0") // Prevent special(ed) browsers from caching the challenge
			renderedTemplate, _ := RenderTemplate("assets/html/pow.html", map[string]interface{}{
				"PublicSalt":        publicSalt,
				"HashedEncryptedIP": hashedEncryptedIP,
				"Stage2Difficulty":  strconv.Itoa(domainData.Stage2Difficulty),
			})
			SendResponse(renderedTemplate, buffer, writer)
			return
		case 3:
			secretPart := encryptedIP[:6]
			publicPart := encryptedIP[6:]

			captchaData := ""
			maskData := ""
			captchaCache, captchaExists := firewall.CacheImgs.Load(secretPart)

			// Captcha expiration
			if captchaExists && func() bool {
				cachedData := captchaCache.([3]string)
				expirationTime, err := strconv.ParseInt(cachedData[2], 10, 64)
				if err == nil && time.Now().Unix() > expirationTime {
					firewall.CacheImgs.Delete(secretPart)
					return false
				}
				return err == nil && time.Now().Unix() <= expirationTime
			}() {
				captchaDataTmp := captchaCache.([3]string)
				captchaData = captchaDataTmp[0]
				maskData = captchaDataTmp[1]
			} else {
				randomShift := rand.Intn(50) - 25
				captchaImg := image.NewRGBA(image.Rect(0, 0, 100, 37))
				randomColor := uint8(rand.Intn(255))
				utils.AddLabel(captchaImg, 0, 18, publicPart[6:], color.RGBA{61, 140, 64, 20})
				utils.AddLabel(captchaImg, rand.Intn(90), rand.Intn(30), publicPart[:6], color.RGBA{255, randomColor, randomColor, 100})
				utils.AddLabel(captchaImg, rand.Intn(25), rand.Intn(20)+10, secretPart, color.RGBA{61, 140, 64, 255})

				amplitude := float64(rand.Intn(10)+10) / 10.0
				period := float64(37) / 5.0
				displacement := func(x, y int) (int, int) {
					dx := amplitude * math.Sin(float64(y)/period)
					dy := amplitude * math.Sin(float64(x)/period)
					return x + int(dx), y + int(dy)
				}
				captchaImg = utils.WarpImg(captchaImg, displacement)

				maskImg := image.NewRGBA(captchaImg.Bounds())
				draw.Draw(maskImg, maskImg.Bounds(), image.Transparent, image.Point{}, draw.Src)

				numTriangles := rand.Intn(20) + 10

				blacklist := make(map[[2]int]bool) // We use this to keep track of already overwritten pixels.
				// it's slightly more performant to not do this but can lead to unsolvable captchas

				for i := 0; i < numTriangles; i++ {
					size := rand.Intn(5) + 10
					x := rand.Intn(captchaImg.Bounds().Dx() - size)
					y := rand.Intn(captchaImg.Bounds().Dy() - size)
					blacklist = utils.DrawTriangle(blacklist, captchaImg, maskImg, x, y, size, randomShift)
				}

				var captchaBuf, maskBuf bytes.Buffer
				if err := png.Encode(&captchaBuf, captchaImg); err != nil {
					SendResponse(nameTxt+"Error: Failed to encode captcha: "+err.Error(), buffer, writer)
					return
				}
				if err := png.Encode(&maskBuf, maskImg); err != nil {
					SendResponse(nameTxt+"Error: Failed to encode captchaMask: "+err.Error(), buffer, writer)
					return
				}

				captchaData = base64.StdEncoding.EncodeToString(captchaBuf.Bytes())
				maskData = base64.StdEncoding.EncodeToString(maskBuf.Bytes())

				// Calculate the expiration timestamp (3 minutes from now)
				expirationTime := time.Now().Add(1 * time.Minute).Unix()
				expirationTimeStr := strconv.FormatInt(expirationTime, 10)
				// Store the captcha data, mask data, and expiration timestamp
				firewall.CacheImgs.Store(secretPart, [3]string{captchaData, maskData, expirationTimeStr})
			}

			writer.Header().Set("Content-Type", "text/html")
			writer.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0") // Prevent special(ed) browsers from caching the challenge

			renderedTemplate, _ := RenderTemplate("assets/html/captcha.html", map[string]interface{}{
				"Ip":          ip,
				"PublicPart":  publicPart,
				"MaskData":    maskData,
				"CaptchaData": captchaData,
			})
			SendResponse(renderedTemplate, buffer, writer)
			return
		default:
			writer.Header().Set("Content-Type", "text/plain")
			SendResponse(blockTxt+"Suspicious request of level "+susLvStr, buffer, writer)
			return
		}
	}

	//Access logs of clients that passed the challenge
	firewall.Mutex.Lock()
	utils.AddLogs(domains.DomainLog{
		Time:      proxy.LastSecondTimeFormated,
		IP:        ip,
		BrowserFP: browser,
		BotFP:     botFp,
		TLSFP:     tlsFp,
		Useragent: reqUa,
		Path:      request.RequestURI,
	}, domainName)

	domainData = domains.DomainsData[domainName]
	domainData.BypassedRequests++
	domains.DomainsData[domainName] = domainData
	firewall.Mutex.Unlock()

	//Reserved proxy-paths

	switch request.URL.Path {
	// Wtf. Why would you expose this
	case "/_bProxy/stats":
		if proxy.Stealth {
			break
		}
		writer.Header().Set("Content-Type", "text/plain")
		SendResponse("Stage: "+utils.StageToString(domainData.Stage)+"\nTotal Requests: "+strconv.Itoa(domainData.TotalRequests)+"\nBypassed Requests: "+strconv.Itoa(domainData.BypassedRequests)+"\nTotal R/s: "+strconv.Itoa(domainData.RequestsPerSecond)+"\nBypassed R/s: "+strconv.Itoa(domainData.RequestsBypassedPerSecond)+"\nProxy Fingerprint: "+proxy.Fingerprint, buffer, writer)
		return
	case "/_bProxy/fingerprint":
		if proxy.Stealth {
			break
		}
		writer.Header().Set("Content-Type", "text/plain")
		SendResponse("IP: "+ip+"\nASN: "+ipInfoASN+"\nCountry: "+ipInfoCountry+"\nIP Requests: "+strconv.Itoa(ipCount)+"\nIP Challenge Requests: "+strconv.Itoa(ipCountCookie)+"\nSusLV: "+strconv.Itoa(susLv)+"\nFingerprint: "+tlsFp+"\nBrowser: "+browser+botFp, buffer, writer)
		return
	case "/_bProxy/verified":
		writer.Header().Set("Content-Type", "text/plain")
		SendResponse("verified", buffer, writer)
		return
	case "/_bProxy/" + proxy.AdminSecret + "/api/v1":
		result := api.Process(writer, request, domainData)
		if result {
			return
		}

	//Do not remove or modify this. It is required by the license
	// Nope, GPL can't impose this
	case "/_bProxy/credits":
		if proxy.Stealth {
			break
		}
		writer.Header().Set("Content-Type", "text/plain")
		SendResponse("BalooProxyX https://github.com/h1v9/balooProxyX;\nBased on BalooProxy: a Lightweight http reverse-proxy https://github.com/41Baloo/balooProxy. Protected by GNU GENERAL PUBLIC LICENSE Version 3, June 2007", buffer, writer)
		return
	}

	ForwardRequest(writer, request, domainSettings, ip, tlsFp, browser, botFp)
}

func ForwardRequest(writer http.ResponseWriter, request *http.Request,
	domainSettings domains.DomainSettings, ip string, tlsFp string,
	browser string, botFp string) {

	// Api V2 for whitelisted ips
	if strings.HasPrefix(request.URL.Path, "/_bProxy/api/v2") {
		result := api.ProcessV2(writer, request)
		if result {
			return
		}
	}

	// Fix allow backend to read ip
	request.Header.Add("X-Forwarded-For", ip)

	// Leaving these for compatibility
	request.Header.Add("X-Real-IP", ip)
	request.Header.Add("proxy-real-ip", ip)
	request.Header.Add("proxy-tls-fp", tlsFp)
	request.Header.Add("proxy-tls-name", browser+botFp)

	domainSettings.DomainProxy.ServeHTTP(writer, request)
}
