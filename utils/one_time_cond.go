package utils

import "sync"

// Synchronous condition which can be signalled only once
type OneTimeCond struct {
	signalled bool
	lock      sync.Mutex
	subcond   *sync.Cond
}

// Create a new OneTimeCond
func NewOneTimeCond() *OneTimeCond {
	cond := &OneTimeCond{}
	cond.subcond = sync.NewCond(&cond.lock)
	return cond
}

// Wait for condition to be signalled.
//
// If the condition was signalled before Wait, Wait returns immediately without blocking.
func (cond *OneTimeCond) Wait() {
	cond.lock.Lock()
	if cond.signalled {
		cond.lock.Unlock()
		return
	}

	cond.subcond.Wait()
	cond.lock.Unlock()
}

// Signal the condition
//
// All goroutines waiting for the condition will be notified to continue
func (cond *OneTimeCond) Signal() {
	cond.lock.Lock()
	cond.signalled = true
	cond.subcond.Broadcast()
	cond.lock.Unlock()
}

// Return if the condition is already signalled
func (cond *OneTimeCond) IsSignalled() (signalled bool) {
	cond.lock.Lock()
	signalled = cond.signalled
	cond.lock.Unlock()
	return
}
