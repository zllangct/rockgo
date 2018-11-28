package threadpool

import (
	"log"
	"sync"
)

// Locker binds a single mutex locked update
type Locker struct {
	lock   *sync.Mutex
	action func(data interface{})
}

// Invoke executes the action on the locker in mutex locked context
func (locker *Locker) Invoke() {
	locker.lock.Lock()
	defer (func() {
		if r := recover(); r != nil {
			log.Printf("%s", r)
		}
		locker.lock.Unlock()
	})()
	locker.action(nil)
}

// InvokeWith executes the action on the locker in mutex locked context with an argument
func (locker *Locker) InvokeWith(data interface{}) {
	locker.lock.Lock()
	defer (func() {
		if r := recover(); r != nil {
			log.Printf("%s", r)
		}
		locker.lock.Unlock()
	})()
	locker.action(data)
}
