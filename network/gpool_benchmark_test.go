package network

import (
	"sync"
	"testing"
	"time"
)

const (
	runTimes  = 1000000
	poolSize  = 50000
	queueSize = 5000
)

func demoTask() {
	time.Sleep(time.Millisecond * 10)
}

//BenchmarkGoroutine benchmark the goroutine doing tasks.
func BenchmarkGoroutine(b *testing.B) {
	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(runTimes)

		for j := 0; j < runTimes; j++ {
			go func() {
				defer wg.Done()
				demoTask()
			}()
		}

		wg.Wait()
	}
}

//BenchmarkGpool benchmarks the goroutine pool.
func BenchmarkGpool(b *testing.B) {
	pool := NewPool(poolSize, queueSize)
	defer Release()
	var wg sync.WaitGroup

	for i := 0; i < b.N; i++ {
		wg.Add(runTimes)
		for j := 0; j < runTimes; j++ {
			AddJobParallel(func(args ...interface{}) {
				defer wg.Done()
				demoTask()
			}, nil, -1, nil)
		}
	}
}
