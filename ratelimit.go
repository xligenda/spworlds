package spworlds

import (
	"context"
	"sync"
	"time"
)

const (
	defaultDelta = 250 * time.Millisecond
	window       = 1 * time.Minute
)

type RateLimiter struct {
	mu          sync.Mutex
	limit       int // 0 means unlimited
	delta       time.Duration
	lastRequest time.Time
	timestamps  []time.Time
}

func NewRateLimiter(n int) *RateLimiter {
	return NewRateLimiterWithDelta(n, defaultDelta)
}

func NewRateLimiterWithDelta(n int, delta time.Duration) *RateLimiter {
	if n < 0 {
		n = 0 // unlimited
	}
	if delta <= 0 {
		delta = defaultDelta
	}
	return &RateLimiter{
		limit: n,
		delta: delta,
	}
}

func (r *RateLimiter) Wait(ctx context.Context) error {
	for {
		r.mu.Lock()
		wait := r.waitDuration(time.Now())
		if wait <= 0 {
			now := time.Now()
			r.lastRequest = now
			r.timestamps = append(r.timestamps, now)
			r.mu.Unlock()
			return nil
		}
		r.mu.Unlock()

		timer := time.NewTimer(wait)
		select {
		case <-timer.C:
			continue
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		}
	}
}

func (r *RateLimiter) waitDuration(now time.Time) time.Duration {
	if r.limit == 0 {
		return 0
	}

	r.purgeExpired(now)

	var wait time.Duration

	if !r.lastRequest.IsZero() {
		if d := r.delta - now.Sub(r.lastRequest); d > wait {
			wait = d
		}
	}

	if len(r.timestamps) >= r.limit {
		if d := r.timestamps[0].Add(window).Sub(now); d > wait {
			wait = d
		}
	}

	return wait
}

func (r *RateLimiter) purgeExpired(now time.Time) {
	cutoff := now.Add(-window)
	i := 0
	for i < len(r.timestamps) && r.timestamps[i].Before(cutoff) {
		i++
	}
	if i > 0 {
		r.timestamps = r.timestamps[i:]
	}
}

func (r *RateLimiter) Limit() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.limit
}

func (r *RateLimiter) Remaining() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.purgeExpired(time.Now())

	return max(r.limit-len(r.timestamps), 0)
}
