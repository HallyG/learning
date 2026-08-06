package main

import (
	"fmt"
	"sync"
	"time"
)

type GRCARateLimiter struct {
	mu sync.Mutex // todo, could shared this as currently different IDs block each other

	interval time.Duration
	burst    int64

	bucket map[string]time.Time
	now    func() time.Time
}

func NewRateLimiter(rate float64, burst int64) *GRCARateLimiter {
	return &GRCARateLimiter{
		interval: time.Duration(float64(time.Second) / rate),
		burst:    burst,
		bucket:   make(map[string]time.Time),
		now:      time.Now,
	}
}

func (r *GRCARateLimiter) Allow(id string) (allowed bool) {
	start := r.now()

	lockStart := r.now()
	r.mu.Lock()
	lockDuration := time.Since(lockStart)

	defer func() {
		r.mu.Unlock()

		fmt.Println(allowed, time.Since(start), lockDuration)
	}()

	now := r.now()
	tat, ok := r.bucket[id]
	if !ok {
		r.bucket[id] = now.Add(r.interval)
		return true
	}

	limit := tat.Add(-time.Duration(r.burst) * r.interval)

	if now.Before(limit) {
		return false
	}

	// Advance the theoretical arrival time
	if now.After(tat) {
		tat = now
	}

	r.bucket[id] = tat.Add(r.interval)
	return true
}
