package metrics

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Request metrics
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "balooproxy_requests_total",
			Help: "Total number of requests received",
		},
		[]string{"domain", "method", "stage"},
	)

	RequestsAllowed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "balooproxy_requests_allowed_total",
			Help: "Total number of requests that passed challenges",
		},
		[]string{"domain", "method"},
	)

	RequestsBlocked = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "balooproxy_requests_blocked_total",
			Help: "Total number of requests blocked",
		},
		[]string{"domain", "reason"},
	)

	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "balooproxy_request_duration_seconds",
			Help:    "Request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"domain", "method"},
	)

	// Challenge metrics
	ChallengeAttempts = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "balooproxy_challenge_attempts_total",
			Help: "Total number of challenge attempts",
		},
		[]string{"domain", "challenge_type"},
	)

	ChallengeSuccess = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "balooproxy_challenge_success_total",
			Help: "Total number of successful challenges",
		},
		[]string{"domain", "challenge_type"},
	)

	ChallengeFailed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "balooproxy_challenge_failed_total",
			Help: "Total number of failed challenges",
		},
		[]string{"domain", "challenge_type"},
	)

	// Connection metrics
	ActiveConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "balooproxy_active_connections",
			Help: "Current number of active connections",
		},
		[]string{"domain"},
	)

	// Cache metrics
	CacheHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "balooproxy_cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"cache_type"},
	)

	CacheMisses = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "balooproxy_cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"cache_type"},
	)

	CacheSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "balooproxy_cache_size",
			Help: "Current cache size",
		},
		[]string{"cache_type"},
	)

	// TLS Fingerprint metrics
	TLSFingerprints = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "balooproxy_tls_fingerprints_total",
			Help: "Total TLS fingerprints seen",
		},
		[]string{"browser", "bot"},
	)

	// Rate limit metrics
	RateLimitHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "balooproxy_ratelimit_hits_total",
			Help: "Total number of rate limit hits",
		},
		[]string{"domain", "limit_type"},
	)

	// Backend metrics
	BackendRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "balooproxy_backend_requests_total",
			Help: "Total requests forwarded to backend",
		},
		[]string{"domain", "status_code"},
	)

	BackendRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "balooproxy_backend_request_duration_seconds",
			Help:    "Backend request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"domain"},
	)

	BackendErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "balooproxy_backend_errors_total",
			Help: "Total backend errors",
		},
		[]string{"domain", "error_type"},
	)

	// Attack detection metrics
	AttackDetected = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "balooproxy_attacks_detected_total",
			Help: "Total attacks detected",
		},
		[]string{"domain", "attack_type"},
	)

	AttackDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "balooproxy_attack_duration_seconds",
			Help:    "Attack duration in seconds",
			Buckets: []float64{10, 30, 60, 300, 600, 1800, 3600},
		},
		[]string{"domain"},
	)

	// Deduplication metrics
	DeduplicatedRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "balooproxy_deduplicated_requests_total",
			Help: "Total number of deduplicated requests",
		},
		[]string{"domain"},
	)

	DeduplicationSavings = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "balooproxy_deduplication_savings_total",
			Help: "Total backend requests saved by deduplication",
		},
		[]string{"domain"},
	)
)

// MetricsCollector holds real-time metrics that aren't suitable for Prometheus
type MetricsCollector struct {
	mu sync.RWMutex

	// Per-domain metrics
	DomainMetrics map[string]*DomainMetrics
}

type DomainMetrics struct {
	RequestsPerSecond         int
	RequestsBypassedPerSecond int
	CurrentStage              int
	UnderAttack               bool
	AttackStartTime           time.Time
}

var Collector = &MetricsCollector{
	DomainMetrics: make(map[string]*DomainMetrics),
}

// UpdateDomainMetrics updates real-time domain metrics
func (mc *MetricsCollector) UpdateDomainMetrics(domain string, rps, bypassed, stage int, underAttack bool) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if mc.DomainMetrics[domain] == nil {
		mc.DomainMetrics[domain] = &DomainMetrics{}
	}

	dm := mc.DomainMetrics[domain]
	dm.RequestsPerSecond = rps
	dm.RequestsBypassedPerSecond = bypassed
	dm.CurrentStage = stage

	// Track attack duration
	if underAttack && !dm.UnderAttack {
		dm.AttackStartTime = time.Now()
	} else if !underAttack && dm.UnderAttack {
		duration := time.Since(dm.AttackStartTime).Seconds()
		AttackDuration.WithLabelValues(domain).Observe(duration)
	}

	dm.UnderAttack = underAttack
}

// GetDomainMetrics returns metrics for a specific domain
func (mc *MetricsCollector) GetDomainMetrics(domain string) *DomainMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.DomainMetrics[domain]
}

// RecordRequest records a request with timing
func RecordRequest(domain, method, stage string, startTime time.Time) {
	RequestsTotal.WithLabelValues(domain, method, stage).Inc()
	duration := time.Since(startTime).Seconds()
	RequestDuration.WithLabelValues(domain, method).Observe(duration)
}

// RecordAllowedRequest records a request that passed challenges
func RecordAllowedRequest(domain, method string) {
	RequestsAllowed.WithLabelValues(domain, method).Inc()
}

// RecordBlockedRequest records a blocked request
func RecordBlockedRequest(domain, reason string) {
	RequestsBlocked.WithLabelValues(domain, reason).Inc()
}

// RecordChallenge records a challenge attempt
func RecordChallenge(domain, challengeType string, success bool) {
	ChallengeAttempts.WithLabelValues(domain, challengeType).Inc()
	if success {
		ChallengeSuccess.WithLabelValues(domain, challengeType).Inc()
	} else {
		ChallengeFailed.WithLabelValues(domain, challengeType).Inc()
	}
}

// RecordCacheAccess records cache hit or miss
func RecordCacheAccess(cacheType string, hit bool) {
	if hit {
		CacheHits.WithLabelValues(cacheType).Inc()
	} else {
		CacheMisses.WithLabelValues(cacheType).Inc()
	}
}

// RecordTLSFingerprint records a TLS fingerprint
func RecordTLSFingerprint(browser, bot string) {
	TLSFingerprints.WithLabelValues(browser, bot).Inc()
}

// RecordRateLimit records a rate limit hit
func RecordRateLimit(domain, limitType string) {
	RateLimitHits.WithLabelValues(domain, limitType).Inc()
}

// RecordBackendRequest records a backend request
func RecordBackendRequest(domain string, statusCode int, duration time.Duration) {
	BackendRequestsTotal.WithLabelValues(domain, strconv.Itoa(statusCode)).Inc()
	BackendRequestDuration.WithLabelValues(domain).Observe(duration.Seconds())
}

// RecordBackendError records a backend error
func RecordBackendError(domain, errorType string) {
	BackendErrors.WithLabelValues(domain, errorType).Inc()
}

// RecordAttack records an attack detection
func RecordAttack(domain, attackType string) {
	AttackDetected.WithLabelValues(domain, attackType).Inc()
}

// RecordDeduplication records request deduplication
func RecordDeduplication(domain string, savedRequests int) {
	DeduplicatedRequests.WithLabelValues(domain).Inc()
	DeduplicationSavings.WithLabelValues(domain).Add(float64(savedRequests))
}

// Handler returns the Prometheus metrics HTTP handler
func Handler() http.Handler {
	return promhttp.Handler()
}
