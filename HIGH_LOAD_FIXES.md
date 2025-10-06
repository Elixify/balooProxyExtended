# High Load Crash Fixes

## Critical Issues Fixed

### 1. **Shared Transport Bug** (CRITICAL)
**Location**: `core/server/serve.go` line 247

**Problem**: All domains were sharing the same `defaultTransport` instance, causing:
- Connection pool conflicts between different backends
- Wrong backend routing (connections reused for wrong domains)
- Resource exhaustion under load
- Crashes when handling multiple domains

**Fix**: Changed `getTripperForDomain()` to create a **unique transport per domain** instead of sharing one global instance.

```go
// BEFORE (BROKEN):
transport, _ = transportMap.LoadOrStore(domain, defaultTransport)  // ❌ All domains share same transport!

// AFTER (FIXED):
newTransport := createTransport()  // ✅ Each domain gets its own transport
transport, _ = transportMap.LoadOrStore(domain, newTransport)
```

### 2. **No Panic Recovery in Middleware** (CRITICAL)
**Location**: `core/server/middleware.go` line 40

**Problem**: Panic recovery was disabled "for performance", causing:
- Any panic in request handling crashes the entire proxy
- Under high load, edge cases trigger panics
- No graceful error handling

**Fix**: Re-enabled `defer pnc.PanicHndl()` in the Middleware function. The overhead is negligible compared to preventing crashes.

### 3. **Template Path Issues** (FIXED EARLIER)
**Locations**: 
- `core/server/serve.go` line 163
- `core/server/middleware.go` lines 284, 367

**Problem**: Code looked for templates in `html/` but they're in `assets/html/`

**Fix**: Updated all template paths to use `assets/html/` prefix.

### 4. **Fingerprint Path Issues** (FIXED EARLIER)
**Location**: `core/config/init.go` lines 108, 111, 114

**Problem**: Code looked for fingerprints in `fingerprints/` but they're in `global/fingerprints/`

**Fix**: Updated all fingerprint paths to use `global/fingerprints/` prefix.

## Additional Recommendations

### Memory Management
The current cache clearing strategy (line 557 in `monitor.go`) only clears when:
- CPU < 15% AND Memory > 25%, OR
- Memory > 95%

This is reasonable but consider:
- Monitoring memory growth patterns
- Adjusting thresholds based on your server specs
- Adding metrics to track cache sizes

### Connection Limits
Current transport settings:
- `MaxIdleConns: 100`
- `MaxIdleConnsPerHost: 20`
- `MaxConnsPerHost: 0` (unlimited)

Under extreme load, consider:
- Setting `MaxConnsPerHost` to a reasonable limit (e.g., 200-500)
- Monitoring connection pool exhaustion
- Adjusting based on backend capacity

### Goroutine Monitoring
Multiple long-running goroutines are spawned:
- `clearProxyCache()` - every 2 minutes
- `generateOTPSecrets()` - periodic
- `evaluateRatelimit()` - every 5 seconds
- `monitorUpdate()` - every 1 second

All have panic recovery, which is good. Consider adding:
- Goroutine leak detection
- Health check endpoints
- Graceful shutdown handling

## Testing Recommendations

1. **Load Testing**: Use tools like `wrk`, `ab`, or `hey` to simulate high load
2. **Memory Profiling**: Use `pprof` endpoints (already in code) to monitor:
   - Heap allocations
   - Goroutine counts
   - Connection pools
3. **Monitor Logs**: Check `crash.log` for any recovered panics
4. **OS Limits**: Ensure your system has adequate:
   - File descriptor limits (`ulimit -n`)
   - Memory limits
   - Connection tracking limits (conntrack)

## Rebuild Required

After these fixes, you must rebuild the application:

```bash
go build -o main
```

Or use your existing build script.

## Expected Improvements

- ✅ No more crashes under high load
- ✅ Proper connection pooling per domain
- ✅ Graceful panic recovery
- ✅ Correct file paths for templates and fingerprints
- ✅ Better resource isolation between domains
