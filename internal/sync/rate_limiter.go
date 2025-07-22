// Rate Limiter for bandwidth control in ZohoSync
// Author: bdstest

package sync

import (
	"context"
	"sync"
	"time"
)

// RateLimiter controls bandwidth usage for sync operations
type RateLimiter struct {
	bytesPerSecond int64
	bucket         int64
	lastUpdate    time.Time
	mutex         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(bytesPerSecond int64) *RateLimiter {
	return &RateLimiter{
		bytesPerSecond: bytesPerSecond,
		bucket:         bytesPerSecond,
		lastUpdate:    time.Now(),
	}
}

// WaitForCapacity waits for available bandwidth capacity
func (rl *RateLimiter) WaitForCapacity(ctx context.Context) error {
	if rl.bytesPerSecond <= 0 {
		return nil // No rate limiting
	}
	
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	now := time.Now()
	elapsed := now.Sub(rl.lastUpdate)
	
	// Refill bucket based on elapsed time
	tokensToAdd := int64(elapsed.Seconds() * float64(rl.bytesPerSecond))
	rl.bucket += tokensToAdd
	
	// Cap at maximum capacity
	if rl.bucket > rl.bytesPerSecond {
		rl.bucket = rl.bytesPerSecond
	}
	
	rl.lastUpdate = now
	return nil
}

// ConsumeCapacity consumes bandwidth capacity
func (rl *RateLimiter) ConsumeCapacity(bytes int64) {
	if rl.bytesPerSecond <= 0 {
		return // No rate limiting
	}
	
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	rl.bucket -= bytes
	if rl.bucket < 0 {
		rl.bucket = 0
	}
}