package utils

import (
	"runtime"
	"sync/atomic"
)

type SpinLock struct {
	locked int32
}

func (sl *SpinLock) Lock() {
	for !atomic.CompareAndSwapInt32(&sl.locked, 0, 1) {
		runtime.Gosched()
	}
}

func (sl *SpinLock) Unlock() {
	atomic.StoreInt32(&sl.locked, 0)
}
