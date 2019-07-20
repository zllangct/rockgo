package Component

import (
	"container/list"
	"github.com/zllangct/RockGO/3rd/iter"
	"sync"
)

const (
	SYSTEM_OP_UPDATE_ADD = iota
	SYSTEM_OP_UPDATE_REMOVE
)

type IUpdate interface {
	Update(context *Context)
}

type UpdateSystem struct {
	SystemBase
	wg         *sync.WaitGroup
	components *list.List
	push       *list.List
	pop        *list.List
}

func (this *UpdateSystem) Init(runtime *Runtime) {
	this.name = "update"
	this.components = list.New()
	this.push = list.New()
	this.pop = list.New()
	this.wg = &sync.WaitGroup{}
	this.runtime = runtime
}

func (this *UpdateSystem) Name() string {
	this.locker.RLock()
	defer this.locker.RUnlock()
	return this.name
}

func (this *UpdateSystem) UpdateFrame() {
	this.locker.Lock()
	//添加上一帧新添加的组件
	if this.push.Len() > 0 {
		this.components.PushBackList(this.push)
		this.push = list.New()
	}
	//移除上一帧删除的组件，添加在前，移除在后，同一帧添加又移除，不执行update
	if this.pop.Len() > 0 {
		for p := this.pop.Front(); p != nil; p = p.Next() {
			for c := this.components.Front(); c != nil; c = c.Next() {
				if c.Value == p.Value {
					this.components.Remove(c)
				}
			}
		}
		this.pop = list.New()
	}
	this.locker.Unlock()

	Iter := iter.FromList(this.components)
	for c, err := Iter.Next(); err == nil; c, err = Iter.Next() {
		this.wg.Add(1)
		ctx := &Context{
			Runtime: this.runtime,
		}
		v := c.(IUpdate)
		this.runtime.workers.Run(func() {
			v.Update(ctx)
		}, func() {
			this.wg.Done()
		})
	}
	this.wg.Wait()
}

func (this *UpdateSystem) Filter(component IComponent) {
	s, ok := component.(IUpdate)
	if ok {
		this.locker.Lock()
		this.push.PushBack(s)
		this.locker.Unlock()
	}
}

func (this *UpdateSystem) IndependentFilter(op int, component IComponent) {
	switch op {
	case SYSTEM_OP_UPDATE_ADD:
		this.Filter(component)
	case SYSTEM_OP_UPDATE_REMOVE:
		s, ok := component.(IUpdate)
		if !ok {
			return
		}
		this.locker.Lock()
		this.pop.PushBack(s)
		this.locker.Unlock()
	}
}
