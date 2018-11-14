package utils

import "sync"

type NewlessPool struct {
	lock      sync.Mutex
	availCond *sync.Cond
	items     []interface{}
}

func NewNewlessPool() *NewlessPool {
	pool := &NewlessPool{
		items: nil,
	}
	pool.availCond = sync.NewCond(&pool.lock)
	return pool
}

func (pool *NewlessPool) Get() (v interface{}) {
	pool.lock.Lock()
	for len(pool.items) == 0 {
		pool.availCond.Wait()
	}
	// pop the last item
	n := len(pool.items)
	v, pool.items[n-1] = pool.items[n-1], nil
	pool.items = pool.items[:n-1]
	pool.lock.Unlock()
	return
}

func (pool *NewlessPool) TryGet() (v interface{}) {
	pool.lock.Lock()
	n := len(pool.items)
	if n > 0 {
		// pop the last item
		v, pool.items[n-1] = pool.items[n-1], nil
		pool.items = pool.items[:n-1]
	}

	pool.lock.Unlock()
	return
}

func (pool *NewlessPool) Put(v interface{}) {
	pool.lock.Lock()
	pool.items = append(pool.items, v)
	pool.availCond.Signal()
	pool.lock.Unlock()
}
