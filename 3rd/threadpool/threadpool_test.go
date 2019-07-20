package threadpool_test

import (
	"github.com/zllangct/RockGO/3rd/assert"
	"github.com/zllangct/RockGO/3rd/threadpool"
	"strconv"
	"sync"
	"testing"
	"time"
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

type Hello struct {
}

func (this *Hello) Hello(str string) {
	sum := 0
	for i := 0; i < 10000; i++ {
		sum = sum + i
	}
	//println("sum:",sum,str)
}
func TestTTT(T *testing.T) {

	tasklist := make([]*Hello, 10000)
	for i := 0; i < 10000; i++ {
		tasklist = append(tasklist, &Hello{})
	}

	//====================== pool
	pool := threadpool.New()
	pool.MaxThreads = 50

	t1 := time.Now()
	wg1 := sync.WaitGroup{}
	for i := 0; i < 10000; i++ {
		wg1.Add(1)
		pool.Run(func() {
			tasklist[i].Hello(strconv.Itoa(1))
			wg1.Done()
		})

	}

	wg1.Done()
	elapsed1 := time.Since(t1)
	println("pool:", elapsed1)

	//========================== traditional

	t2 := time.Now()
	wg := sync.WaitGroup{}
	wg.Add(10000)
	for j := 0; j < 50; j++ {
		go func() {
			for i := 0; i < 200; i++ {
				tasklist[i].Hello(strconv.Itoa(2))
				wg.Done()
			}

		}()
	}
	wg.Wait()
	elapsed2 := time.Since(t2)
	println("traditional:", elapsed2)

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
