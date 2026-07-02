package spworlds

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestNewRateLimiter_DefaultDelta(t *testing.T) {
	r := NewRateLimiter(5)
	if r.limit != 5 {
		t.Fatalf("limit = %d, want 5", r.limit)
	}
	if r.delta != defaultDelta {
		t.Fatalf("delta = %v, want %v", r.delta, defaultDelta)
	}
}

func TestNewRateLimiterWithDelta_NegativeN_BecomesUnlimited(t *testing.T) {
	r := NewRateLimiterWithDelta(-10, time.Second)
	if r.limit != 0 {
		t.Fatalf("limit = %d, want 0 (unlimited)", r.limit)
	}
}

func TestNewRateLimiterWithDelta_ZeroN_IsUnlimited(t *testing.T) {
	r := NewRateLimiterWithDelta(0, time.Second)
	if r.limit != 0 {
		t.Fatalf("limit = %d, want 0 (unlimited)", r.limit)
	}
}

func TestNewRateLimiterWithDelta_InvalidDelta_FallsBackToDefault(t *testing.T) {
	cases := []time.Duration{0, -1 * time.Second}
	for _, d := range cases {
		r := NewRateLimiterWithDelta(3, d)
		if r.delta != defaultDelta {
			t.Fatalf("delta for input %v = %v, want default %v", d, r.delta, defaultDelta)
		}
	}
}

func TestNewRateLimiterWithDelta_ValidDelta_Kept(t *testing.T) {
	r := NewRateLimiterWithDelta(3, 10*time.Second)
	if r.delta != 10*time.Second {
		t.Fatalf("delta = %v, want 10s", r.delta)
	}
}

func TestWaitDuration_UnlimitedAlwaysZero(t *testing.T) {
	r := NewRateLimiterWithDelta(0, time.Hour)
	now := time.Now()
	r.lastRequest = now
	if got := r.waitDuration(now); got != 0 {
		t.Fatalf("waitDuration = %v, want 0 for unlimited limiter", got)
	}
}

func TestWaitDuration_EnforcesDelta(t *testing.T) {
	r := NewRateLimiterWithDelta(100, 250*time.Millisecond)
	now := time.Now()
	r.lastRequest = now

	got := r.waitDuration(now)
	if got <= 0 || got > r.delta {
		t.Fatalf("waitDuration = %v, want in (0, %v]", got, r.delta)
	}

	later := now.Add(r.delta + time.Millisecond)
	if got := r.waitDuration(later); got != 0 {
		t.Fatalf("waitDuration after delta = %v, want 0", got)
	}
}

func TestWaitDuration_EnforcesWindowLimit(t *testing.T) {
	r := NewRateLimiterWithDelta(2, time.Nanosecond)
	base := time.Now()

	r.timestamps = []time.Time{base, base.Add(time.Second)}
	r.lastRequest = base.Add(time.Second)

	now := base.Add(2 * time.Second)
	got := r.waitDuration(now)
	want := r.timestamps[0].Add(window).Sub(now)
	if got != want {
		t.Fatalf("waitDuration = %v, want %v", got, want)
	}
	if got <= 0 {
		t.Fatalf("expected positive wait while window limit exceeded, got %v", got)
	}

	after := base.Add(window + 2*time.Second)
	if got := r.waitDuration(after); got != 0 {
		t.Fatalf("waitDuration after window expiry = %v, want 0", got)
	}
}

func TestWaitDuration_TakesMaxOfDeltaAndWindowWait(t *testing.T) {
	r := NewRateLimiterWithDelta(1, time.Minute)
	base := time.Now()
	r.lastRequest = base
	r.timestamps = []time.Time{base}

	now := base.Add(time.Second)
	got := r.waitDuration(now)

	deltaWait := r.delta - now.Sub(r.lastRequest)
	windowWait := r.timestamps[0].Add(window).Sub(now)

	want := deltaWait
	if windowWait > want {
		want = windowWait
	}
	if got != want {
		t.Fatalf("waitDuration = %v, want max(delta,window) = %v", got, want)
	}
}

func TestPurgeExpired_RemovesOldTimestamps(t *testing.T) {
	r := &RateLimiter{}
	now := time.Now()
	r.timestamps = []time.Time{
		now.Add(-2 * window),
		now.Add(-90 * time.Second),
		now.Add(-30 * time.Second),
		now,
	}
	r.purgeExpired(now)

	if len(r.timestamps) != 2 {
		t.Fatalf("timestamps left = %d, want 2, got %v", len(r.timestamps), r.timestamps)
	}
	cutoff := now.Add(-window)
	for _, ts := range r.timestamps {
		if ts.Before(cutoff) {
			t.Fatalf("expired timestamp %v was not purged", ts)
		}
	}
}

func TestPurgeExpired_NoneExpired_NoOp(t *testing.T) {
	r := &RateLimiter{}
	now := time.Now()
	r.timestamps = []time.Time{now.Add(-10 * time.Second), now}
	r.purgeExpired(now)
	if len(r.timestamps) != 2 {
		t.Fatalf("timestamps left = %d, want 2", len(r.timestamps))
	}
}

func TestWait_UnlimitedNeverBlocks(t *testing.T) {
	r := NewRateLimiter(0)
	start := time.Now()
	for i := 0; i < 50; i++ {
		if err := r.Wait(context.Background()); err != nil {
			t.Fatalf("Wait returned error: %v", err)
		}
	}
	if elapsed := time.Since(start); elapsed > 100*time.Millisecond {
		t.Fatalf("unlimited Wait took too long: %v", elapsed)
	}
}

func TestWait_RespectsDeltaBetweenCalls(t *testing.T) {
	delta := 100 * time.Millisecond
	r := NewRateLimiterWithDelta(1000, delta)

	if err := r.Wait(context.Background()); err != nil {
		t.Fatalf("first Wait error: %v", err)
	}
	start := time.Now()
	if err := r.Wait(context.Background()); err != nil {
		t.Fatalf("second Wait error: %v", err)
	}
	elapsed := time.Since(start)

	if elapsed < delta-10*time.Millisecond {
		t.Fatalf("second Wait returned too early: elapsed=%v, delta=%v", elapsed, delta)
	}
}

func TestWait_ContextCancelledWhileWaiting(t *testing.T) {
	r := NewRateLimiterWithDelta(1000, 500*time.Millisecond)

	if err := r.Wait(context.Background()); err != nil {
		t.Fatalf("first Wait error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := r.Wait(ctx)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected context error, got nil")
	}
	if err != context.DeadlineExceeded {
		t.Fatalf("err = %v, want context.DeadlineExceeded", err)
	}
	if elapsed > 200*time.Millisecond {
		t.Fatalf("Wait should have returned promptly after ctx timeout, took %v", elapsed)
	}
}

func TestWait_AlreadyCancelledContext(t *testing.T) {
	r := NewRateLimiterWithDelta(1, 500*time.Millisecond)
	if err := r.Wait(context.Background()); err != nil {
		t.Fatalf("first Wait error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := r.Wait(ctx)
	if err != context.Canceled {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestWait_ConcurrentCallsAreSerializedByDelta(t *testing.T) {
	delta := 30 * time.Millisecond
	r := NewRateLimiterWithDelta(1000, delta)

	const n = 5
	var wg sync.WaitGroup
	start := time.Now()
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = r.Wait(context.Background())
		}()
	}
	wg.Wait()
	elapsed := time.Since(start)

	minExpected := time.Duration(n-1) * delta
	if elapsed < minExpected-10*time.Millisecond {
		t.Fatalf("elapsed = %v, want at least ~%v for %d serialized calls", elapsed, minExpected, n)
	}
}

func TestLimit_ReturnsConfiguredValue(t *testing.T) {
	r := NewRateLimiter(7)
	if got := r.Limit(); got != 7 {
		t.Fatalf("Limit() = %d, want 7", got)
	}
}

func TestRemaining_DecreasesWithUsage(t *testing.T) {
	r := NewRateLimiterWithDelta(3, time.Nanosecond)

	if got := r.Remaining(); got != 3 {
		t.Fatalf("Remaining() initial = %d, want 3", got)
	}

	for i := 0; i < 2; i++ {
		if err := r.Wait(context.Background()); err != nil {
			t.Fatalf("Wait error: %v", err)
		}
	}

	if got := r.Remaining(); got != 1 {
		t.Fatalf("Remaining() after 2 requests = %d, want 1", got)
	}
}

func TestRemaining_NeverNegative(t *testing.T) {
	r := NewRateLimiterWithDelta(1, time.Nanosecond)
	now := time.Now()
	r.timestamps = []time.Time{now, now, now}

	if got := r.Remaining(); got < 0 {
		t.Fatalf("Remaining() = %d, want >= 0", got)
	}
	if got := r.Remaining(); got != 0 {
		t.Fatalf("Remaining() = %d, want 0 when over limit", got)
	}
}

func TestRemaining_UnlimitedIsAlwaysLimitValue(t *testing.T) {
	r := NewRateLimiter(0)
	for range 5 {
		_ = r.Wait(context.Background())
	}

	if got := r.Remaining(); got != 0 {
		t.Fatalf("Remaining() for unlimited = %d, want 0 (limit-len clamps to 0)", got)
	}
}
