// packag coin implements distributed bitcoin mining.
// this part has utilities
package coin

import "sync"

/*
// refresh the channel

    getSet.mu.Lock()
    getSet.on.theLine = make(chan struct{})
    getSet.mu.Unlock()

// safe close
    watchNet.mu.Lock()
	if !weWonOutAlready() {
		close(watchNet.weWon) // cancel previous getNewBlocks
	}
	watchNet.mu.Unlock()

// new should also generate this test function

    func weWonOutAlready() bool {
	select {
	case <-watchNet.weWon:
		return true
	default:
		return false
	}
}

*/

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

// Close safely closes our channel
func (ab *Abort) Close() {
	ab.mu.Lock()
	if !ab.isClosed() {
		ab.Close()
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
