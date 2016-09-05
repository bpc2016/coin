// Package coin implements distributed bitcoin mining.
// this part has utilities
package coin

import "sync"

// Abort is a secured channel
type Abort struct {
	mu sync.Mutex
	ch chan struct{}
}

// IsClosed returns a function that tests whether our channel is closed
func (ab *Abort) IsClosed() func() bool {
	f := func() bool {
		select {
		case <-ab.ch:
			return true
		default:
			return false
		}
	}
	return f
}

func (ab *Abort) isClosed() bool {
	select {
	case <-ab.ch:
		return true
	default:
		return false
	}
}

// Cancel safely closes our channel
func (ab *Abort) Cancel() {
	ab.mu.Lock()
	if !ab.isClosed() {
		close(ab.ch)
	}
	ab.mu.Unlock()
}

// Revive regenerates our channel
func (ab *Abort) Revive() {
	ab.mu.Lock()
	if ab.isClosed() {
		ab.ch = make(chan struct{})
	}
	ab.mu.Unlock()
}

// Chan exposes our channel for use
func (ab *Abort) Chan() chan struct{} {
	return ab.ch
}

// New initialises our channel
func (ab *Abort) New() {
	ab.ch = make(chan struct{})
}
