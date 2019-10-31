package ecs

import (
	"container/list"
	"sync"
)

const (
	SYSTEM_OP_DESTROY_ADD = iota
	SYSTEM_OP_DESTROY_REMOVE
)

type IDestroy interface {
	Destroy(context *Context)
}

type DestroySystem struct {
	SystemBase
	wg         *sync.WaitGroup
	components *list.List
	temp       *list.List
}

func (this *DestroySystem) Init(runtime *Runtime) {
	this.name = "destroy"
	this.components = list.New()
	this.temp = list.New()
	this.wg = &sync.WaitGroup{}
	this.runtime = runtime
}

func (this *DestroySystem) Name() string {
	this.locker.RLock()
	defer this.locker.RUnlock()
	return this.name
}

func (this *DestroySystem) UpdateFrame() {
	this.locker.Lock()
	this.components, this.temp = this.temp, this.components
	this.locker.Unlock()

	for c := this.temp.Front(); c != nil; c = this.temp.Front() {
		this.wg.Add(1)
		ctx := &Context{
			Runtime: this.runtime,
		}

		v := c.Value.(IDestroy)
		this.runtime.workers.Run(func() {
			v.Destroy(ctx)
			this.wg.Done()
		})
		this.temp.Remove(c)
	}
	this.wg.Wait()
}

func (this *DestroySystem) Filter(component IComponent) {

}

func (this *DestroySystem) IndependentFilter(op int, component IComponent) {
	this.locker.Lock()
	defer this.locker.Unlock()

	s, ok := component.(IDestroy)
	if ok {
		this.components.PushBack(s)
	}
}
