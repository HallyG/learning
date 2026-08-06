package main

import (
	"sync/atomic"
	"time"
)

type LimiterMetrics struct {
	Allowed       atomic.Int64
	Denied        atomic.Int64
	LockWaitNanos atomic.Int64
	DurationNanos atomic.Int64
}

func (m *LimiterMetrics) RecordDecision(allowed bool) {
	if allowed {
		m.Allowed.Add(1)
	} else {
		m.Denied.Add(1)
	}
}

func (m *LimiterMetrics) RecordDuration(d time.Duration) {
	m.DurationNanos.Add(d.Nanoseconds())
}

func (m *LimiterMetrics) RecordLockWait(d time.Duration) {
	m.LockWaitNanos.Add(d.Nanoseconds())
}
