package Component

import (
	"sync"
)

type ISystem interface {
	Init(runtime *Runtime)
	UpdateFrame()
	Filter(component IComponent)
	IndependentFilter(op int,component IComponent)
	Name()string
}

type SystemBase struct {
	locker sync.RWMutex
	runtime   *Runtime
	name  string
}



