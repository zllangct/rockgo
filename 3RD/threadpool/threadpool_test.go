package threadpool_test

import (
	"github.com/zllangct/RockGO/3RD/assert"
	"github.com/zllangct/RockGO/3RD/threadpool"
	"testing"
)

func TestRun(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
		pool := threadpool.New()
		value := 0

		pool.Run(func() { value += 1 })
		pool.Wait()

		T.Assert(value == 1)
	})
}

func TestBusy(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
		pool := threadpool.New()
		pool.MaxThreads = 2

		value := 0

		pool.Run(func() { value += 1 })
		pool.Run(func() { value += 1 })
		pool.Run(func() { value += 1 })

		pool.Wait()
		T.Assert(value == 3)
	})
}

func TestRunAndWait(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
		pool := threadpool.New()
		pool.MaxThreads = 10

		value := 0
		procs := 0

		inc := pool.Locker(func() {
			value += 1
		})

		for i := 0; i < 50; i++ {
			pool.Run(inc.Invoke)
			active := pool.Active()
			if active > procs {
				procs = active
			}
		}

		pool.Wait()
		T.Assert(procs == 10)
		T.Assert(value == 50)
	})
}

func TestResolveAsync(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
		pool := threadpool.New()
		pool.MaxThreads = 10

		resolved := 0
		rejected := false

		resolveToValue := pool.LockerWith(func(value interface{}) {
			intValue, _ := value.(int)
			resolved = intValue
		})

		rejectValue := pool.Locker(func() {
			rejected = true
		})

		pool.Run(func() {
			resolveToValue.InvokeWith(100)
		})

		pool.Wait()
		T.Assert(resolved == 100)
		T.Assert(rejected == false)

		resolved = 0
		rejected = false

		pool.Run(func() {
			rejectValue.Invoke()
		})

		pool.Wait()
		T.Assert(resolved == 0)
		T.Assert(rejected == true)
	})
}
