package Component

import (
	"container/list"
	"sync"
)

type IStart interface {
	Start(context *Context)
}

type StartSystem struct {
	SystemBase
	wg *sync.WaitGroup
	components *list.List
	temp *list.List
}

func (this *StartSystem)Init(runtime *Runtime)  {
	this.name="start"
	this.components = list.New()
	this.temp = list.New()
	this.wg = &sync.WaitGroup{}
	this.runtime=runtime
}

func (this *StartSystem)Name() string {
	this.locker.RLock()
	defer this.locker.RUnlock()
	return this.name
}

func (this *StartSystem)UpdateFrame()  {
	this.locker.Lock()
	this.components,this.temp=this.temp,this.components
	this.locker.Unlock()

	for c:=this.temp.Front(); c!=nil; c=this.temp.Front() {
		this.wg.Add(1)
		ctx:=&Context{
			Runtime:this.runtime,
		}

		v:=c.Value.(IStart)
		this.runtime.workers.Run(func() {
			v.Start(ctx)
		}, func() {
			this.wg.Done()
		})
		this.temp.Remove(c)
	}
	this.wg.Wait()
}

func (this *StartSystem)Filter(component IComponent)  {
	this.locker.Lock()
	defer this.locker.Unlock()

	s,ok:=component.(IStart)
	if ok{
		this.components.PushBack(s)
	}
}
func (this *StartSystem)IndependentFilter(op int,component IComponent)  {
	this.Filter(component)
}