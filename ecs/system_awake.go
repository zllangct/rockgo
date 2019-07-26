package ecs

import (
	"container/list"
	"sync"
)

type IAwake interface {
	Awake(context *Context)
}

type AwakeSystem struct {
	SystemBase
	wg         *sync.WaitGroup
	components *list.List
	temp       *list.List
}

func (this *AwakeSystem) Init(runtime *Runtime) {
	this.name = "awake"
	this.components = list.New()
	this.temp = list.New()
	this.wg = &sync.WaitGroup{}
	this.runtime = runtime
}

func (this *AwakeSystem) Name() string {
	this.locker.RLock()
	defer this.locker.RUnlock()
	return this.name
}

func (this *AwakeSystem) UpdateFrame() {
	this.locker.Lock()
	this.components, this.temp = this.temp, this.components
	this.locker.Unlock()

	for c := this.temp.Front(); c != nil; c = this.temp.Front() {
		this.wg.Add(1)
		ctx := &Context{
			Runtime: this.runtime,
		}

		v := c.Value.(IAwake)
		//name:=c.Value.(IComponent).Type().String()
		this.runtime.workers.Run(func() {
			//logger.Debug("awake: "+name)
			v.Awake(ctx)
		}, func() {
			this.wg.Done()
		})
		this.temp.Remove(c)
	}
	this.wg.Wait()
}

func (this *AwakeSystem) Filter(component IComponent) {
	this.locker.Lock()
	defer this.locker.Unlock()

	s, ok := component.(IAwake)
	if ok {
		this.components.PushBack(s)
	}
}

func (this *AwakeSystem) IndependentFilter(op int, component IComponent) {
	this.Filter(component)
}
