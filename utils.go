package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"math"
	"math/rand/v2"
	"time"
)

func StrHash(s1 string) string {
	hash := sha256.New()
	hash.Write([]byte(s1))
	hashedBytes := hash.Sum(nil)
	return hex.EncodeToString(hashedBytes)
}

type backoff struct {
	initial      time.Duration
	current      time.Duration
	maxInterval  time.Duration
	multiplier   float64
	maxRetries   int
	currentRetry int
	jitter       int
}

func newBackoff(initial time.Duration, maxInterval time.Duration, multiplier float64, maxRetries int) *backoff {
	return &backoff{
		initial * time.Second,
		initial * time.Second,
		maxInterval * time.Second,
		multiplier,
		maxRetries,
		0,
		4,
	}
}

func (b *backoff) Backoff() error {
	/* error if max retries already breached */
	if b.currentRetry >= b.maxRetries {
		return fmt.Errorf("RetryCountBreeched")
	}

	/* add 25% jitter to the current interval to avoid stampeding herd */
	jitter := b.current / 4
	wait := b.current + time.Duration(rand.Float64()*float64(jitter))
	/* sleep for the current interval */
	slog.Info("Waiting", "time", wait, "attempt", b.currentRetry+1, "of", b.maxRetries)
	time.Sleep(wait)

	/* Increment currentRetry */
	b.currentRetry++

	/* calculate next interval */
	b.current = b.initial * time.Duration(math.Pow(b.multiplier, float64(b.currentRetry)))

	if b.current > b.maxInterval {
		slog.Info("Next calculated backoff interval is greater than max interval, reverting to max interval")
		b.current = b.maxInterval
	}
	slog.Info("Next backup interval calculated", "interval", b.current.Seconds())

	return nil
}

func (b *backoff) Reset() {
	b.current = b.initial
	b.currentRetry = 0
}

func (b *backoff) GetRetries() int {
	return b.currentRetry
}
