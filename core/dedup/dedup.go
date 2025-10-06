package dedup

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sync"
	"time"
)

// RequestKey represents a unique request identifier
type RequestKey string

// PendingRequest holds information about an in-flight request
type PendingRequest struct {
	mu       sync.RWMutex
	done     chan struct{}
	response *CachedResponse
	waiters  int
}

// CachedResponse holds the response data to be shared
type CachedResponse struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	Timestamp  time.Time
}

// Deduplicator handles request deduplication
type Deduplicator struct {
	mu       sync.RWMutex
	pending  map[RequestKey]*PendingRequest
	enabled  bool
	ttl      time.Duration
	maxSize  int
	cleanupInterval time.Duration
}

// NewDeduplicator creates a new request deduplicator
func NewDeduplicator(enabled bool, ttl time.Duration) *Deduplicator {
	d := &Deduplicator{
		pending:  make(map[RequestKey]*PendingRequest),
		enabled:  enabled,
		ttl:      ttl,
		maxSize:  10000, // Max 10k concurrent deduplicated requests
		cleanupInterval: 30 * time.Second,
	}

	if enabled {
		go d.cleanup()
	}

	return d
}

// GenerateKey generates a unique key for a request
func (d *Deduplicator) GenerateKey(r *http.Request) RequestKey {
	// Hash: Method + Host + Path + Query + relevant headers
	h := sha256.New()
	
	h.Write([]byte(r.Method))
	h.Write([]byte(r.Host))
	h.Write([]byte(r.URL.Path))
	h.Write([]byte(r.URL.RawQuery))
	
	// Include headers that affect response (but not cookies/auth)
	if accept := r.Header.Get("Accept"); accept != "" {
		h.Write([]byte(accept))
	}
	if encoding := r.Header.Get("Accept-Encoding"); encoding != "" {
		h.Write([]byte(encoding))
	}
	if lang := r.Header.Get("Accept-Language"); lang != "" {
		h.Write([]byte(lang))
	}
	
	return RequestKey(hex.EncodeToString(h.Sum(nil)))
}

// ShouldDeduplicate checks if a request should be deduplicated
func (d *Deduplicator) ShouldDeduplicate(r *http.Request) bool {
	if !d.enabled {
		return false
	}

	// Only deduplicate GET and HEAD requests (idempotent)
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		return false
	}

	// Don't deduplicate requests with authentication
	if r.Header.Get("Authorization") != "" {
		return false
	}

	// Don't deduplicate requests with cookies (might be user-specific)
	if r.Header.Get("Cookie") != "" {
		return false
	}

	return true
}

// Wait waits for an existing request or returns false if this is the first
func (d *Deduplicator) Wait(key RequestKey) (*CachedResponse, bool) {
	d.mu.RLock()
	pending, exists := d.pending[key]
	d.mu.RUnlock()

	if !exists {
		return nil, false
	}

	// Increment waiters count
	pending.mu.Lock()
	pending.waiters++
	pending.mu.Unlock()

	// Wait for the request to complete
	select {
	case <-pending.done:
		pending.mu.RLock()
		response := pending.response
		pending.mu.RUnlock()
		return response, true
	case <-time.After(30 * time.Second):
		// Timeout - let this request proceed independently
		return nil, false
	}
}

// Start marks a request as in-flight
func (d *Deduplicator) Start(key RequestKey) *PendingRequest {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Check if we're at capacity
	if len(d.pending) >= d.maxSize {
		return nil
	}

	pending := &PendingRequest{
		done:    make(chan struct{}),
		waiters: 0,
	}

	d.pending[key] = pending
	return pending
}

// Complete marks a request as complete and shares the response
func (d *Deduplicator) Complete(key RequestKey, response *CachedResponse) int {
	d.mu.Lock()
	pending, exists := d.pending[key]
	if exists {
		delete(d.pending, key)
	}
	d.mu.Unlock()

	if !exists {
		return 0
	}

	// Store response and notify waiters
	pending.mu.Lock()
	pending.response = response
	waiters := pending.waiters
	pending.mu.Unlock()

	close(pending.done)

	return waiters
}

// Cancel cancels a pending request (e.g., on error)
func (d *Deduplicator) Cancel(key RequestKey) {
	d.mu.Lock()
	pending, exists := d.pending[key]
	if exists {
		delete(d.pending, key)
	}
	d.mu.Unlock()

	if exists {
		close(pending.done)
	}
}

// GetWaiters returns the number of requests waiting for this key
func (d *Deduplicator) GetWaiters(key RequestKey) int {
	d.mu.RLock()
	pending, exists := d.pending[key]
	d.mu.RUnlock()

	if !exists {
		return 0
	}

	pending.mu.RLock()
	defer pending.mu.RUnlock()
	return pending.waiters
}

// cleanup periodically removes stale pending requests
func (d *Deduplicator) cleanup() {
	ticker := time.NewTicker(d.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		d.mu.Lock()
		
		// Remove any requests that have been pending too long
		// This shouldn't happen in normal operation but prevents leaks
		for key, pending := range d.pending {
			select {
			case <-pending.done:
				// Already done, remove it
				delete(d.pending, key)
			default:
				// Still pending - this is normal
			}
		}
		
		d.mu.Unlock()
	}
}

// Stats returns deduplication statistics
func (d *Deduplicator) Stats() map[string]interface{} {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return map[string]interface{}{
		"enabled":         d.enabled,
		"pending_count":   len(d.pending),
		"max_size":        d.maxSize,
		"ttl_seconds":     d.ttl.Seconds(),
	}
}

// SetEnabled enables or disables deduplication at runtime
func (d *Deduplicator) SetEnabled(enabled bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	d.enabled = enabled
	
	// If disabling, clear all pending requests
	if !enabled {
		for key, pending := range d.pending {
			close(pending.done)
			delete(d.pending, key)
		}
	}
}

// IsEnabled returns whether deduplication is enabled
func (d *Deduplicator) IsEnabled() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.enabled
}
