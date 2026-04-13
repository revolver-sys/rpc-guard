package main

import (
	"sync"
	"time"
)

type Endpoint struct {
	URL string

	mu sync.RWMutex

	avgLatency   time.Duration
	successes    int
	failures     int
	consecFails  int
	openedUntil  time.Time
	lastError    string
	lastUsedAt   time.Time
	lastSuccess  time.Time
	lastFailure  time.Time
}

func NewEndpoint(url string) *Endpoint {
	return &Endpoint{
		URL:        url,
		avgLatency: 500 * time.Millisecond,
	}
}

func (e *Endpoint) IsAvailable(now time.Time) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return !now.Before(e.openedUntil)
}

func (e *Endpoint) RecordSuccess(latency time.Duration, now time.Time) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.successes++
	e.consecFails = 0
	e.lastUsedAt = now
	e.lastSuccess = now

	// Exponential moving average with a simple weight.
	e.avgLatency = (e.avgLatency*4 + latency) / 5
}

func (e *Endpoint) RecordFailure(err string, now time.Time, openFor time.Duration) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.failures++
	e.consecFails++
	e.lastError = err
	e.lastUsedAt = now
	e.lastFailure = now

	if e.consecFails >= 2 {
		e.openedUntil = now.Add(openFor)
	}
}

func (e *Endpoint) Score(now time.Time) float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	score := 1000.0

	// Penalize endpoints in open-circuit state very heavily.
	if now.Before(e.openedUntil) {
		score -= 10000
	}

	// Lower latency = better score.
	score -= float64(e.avgLatency.Milliseconds())

	// Penalize failures.
	score -= float64(e.failures * 20)
	score -= float64(e.consecFails * 100)

	// Reward successes a bit.
	score += float64(e.successes * 5)

	return score
}

func (e *Endpoint) Snapshot(now time.Time) map[string]any {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return map[string]any{
		"url":          e.URL,
		"avg_latency":  e.avgLatency.String(),
		"successes":    e.successes,
		"failures":     e.failures,
		"consec_fails": e.consecFails,
		"available":    !now.Before(e.openedUntil),
		"opened_until": e.openedUntil.Format(time.RFC3339),
		"last_error":   e.lastError,
	}
}
