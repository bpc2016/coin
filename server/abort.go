// Package coin implements distributed bitcoin mining.
// this part has utilities
package main

import "sync"

// Abort is a secured channel
type Abort struct {
	sync.Mutex
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
func (ab *Abort) Cancel() bool {
	ab.Lock()
	defer ab.Unlock()
	if !ab.isClosed() {
		close(ab.ch)
		return false
	}
	return true
}

// Revive regenerates our channel
func (ab *Abort) Revive() {
	ab.Lock()
	if ab.isClosed() {
		ab.ch = make(chan struct{})
	}
	ab.Unlock()
}

// Chan exposes our channel for use
func (ab *Abort) Chan() chan struct{} {
	return ab.ch
}

// New initialises our channel
func (ab *Abort) New() {
	ab.ch = make(chan struct{})
}
