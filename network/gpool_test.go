package network

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
)

func init() {
	println("using MAXPROC")
	numCPUs := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPUs)
}

//TestNewPool test new goroutine pool
func TestNewPool(t *testing.T) {
	pool := NewPool(1000, 10000)
	defer Release()

	iterations := 1000000
	var counter uint64

	wg := sync.WaitGroup{}
	wg.Add(iterations)
	for i := 0; i < iterations; i++ {
		arg := uint64(1)
		AddJobParallel(func(args ...interface{}) {
			defer wg.Done()
			atomic.AddUint64(&counter, arg)
		}, nil, -1, nil)
	}
	wg.Wait()

	counterFinal := atomic.LoadUint64(&counter)
	if uint64(iterations) != counterFinal {
		t.Errorf("iterations %v is not equal counterFinal %v", iterations, counterFinal)
	}
}
