package threadpool

import (
	"container/list"
	"log"
	"sync"
)

// ThreadPool is a common high level interface for a single producer multi-consumer pattern.
type ThreadPool struct {
	MaxThreads int
	active     int
	lock       *sync.Mutex
	any        *sync.Mutex
	pending    *list.List
}

// New returns a new empty ThreadPool
func New() *ThreadPool {
	return &ThreadPool{
		MaxThreads: -1,
		active:     0,
		pending:    list.New(),
		lock:       &sync.Mutex{},
		any:        &sync.Mutex{}}
}

// Locker returns a new locker with the given action
func (pool *ThreadPool) Locker(action func()) *Locker {
	return pool.LockerWith(func(_ interface{}) { action() })
}

// LockerWith returns a new locker with the given action that takes an argument
func (pool *ThreadPool) LockerWith(action func(interface{})) *Locker {
	return &Locker{
		lock:   &sync.Mutex{},
		action: action}
}

// Run starts a new task, or puts the task on the queue of tasks to run.
func (pool *ThreadPool) Run(task func()) {
	pool.run(task, true)
}

// run starts a new task, or puts the task on the queue of tasks to run.
func (pool *ThreadPool) run(task func(), requireLock bool) {
	if requireLock {
		pool.lock.Lock()
	}
	if pool.activeUp() {
		go func() {
			defer func() {
				pool.lock.Lock()
				if r := recover(); r != nil {
					log.Printf("%s", r)
				}
				pool.activeDown()
				pool.nextTask()
				pool.lock.Unlock()
			}()
			task()
		}()
	} else {
		pool.pending.PushBack(task)
	}
	if requireLock {
		pool.lock.Unlock()
	}
}

// nextTask runs a task if there is any pending task
func (pool *ThreadPool) nextTask() {
	if pool.pending.Len() > 0 {
		task, _ := pool.pending.Remove(pool.pending.Front()).(func())
		pool.run(task, false)
	}
}

// Active returns a count of active threads.
func (pool *ThreadPool) Active() int {
	return pool.active
}

// Wait blocks until all tasks are completed.
func (pool *ThreadPool) Wait() {
	pool.any.Lock()
	pool.any.Unlock()
}

// Update the active count and lock state
func (pool *ThreadPool) activeUp() bool {
	if pool.MaxThreads < 0 || pool.active < pool.MaxThreads {
		pool.active++
		if pool.active == 1 {
			pool.any.Lock()
		}
		return true
	}
	return false
}

// Update the active count and lock state
func (pool *ThreadPool) activeDown() {
	pool.active--
	if pool.active == 0 {
		pool.any.Unlock()
	}
}
