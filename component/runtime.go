package Component

import (
	"errors"
	"fmt"
	"github.com/zllangct/RockGO/logger"
	"sync"
	"time"

	"github.com/zllangct/RockGO/3rd/iter"
	"github.com/zllangct/RockGO/3rd/threadpool"
)

type Config struct {
	ThreadPoolSize int
	Factory        *ObjectFactory
}

type Runtime struct {
	locker        *sync.RWMutex
	root          *Object
	workers       *threadpool.ThreadPool
	updateLock    *sync.Mutex
	factory       *ObjectFactory
	innerSystems  []ISystem
	customSystems []ISystem
}

func NewRuntime(config Config) *Runtime {
	validateConfig(&config)
	runtime := &Runtime{
		root:         NewObject("runtime"),
		updateLock:   &sync.Mutex{},
		workers:      threadpool.New(),
		factory:      config.Factory,
		innerSystems: []ISystem{&AwakeSystem{}, &StartSystem{}, &UpdateSystem{}, &DestroySystem{}},
		locker:       &sync.RWMutex{},
	}
	runtime.root.runtime = runtime
	runtime.workers.MaxThreads = config.ThreadPoolSize
	for _, s := range runtime.innerSystems {
		s.Init(runtime)
	}
	return runtime
}

func (this *Runtime) UpdateFrameByInterval(duration time.Duration) chan<- struct{} {
	shutdown := make(chan struct{})
	c := time.Tick(duration)
	go func() {
		ticking := false
		for {
			select {
			case <-shutdown:
				return
			case <-c:
				if ticking {
					continue
				}
				//上一帧还未执行完毕时，跳过一帧，避免帧滚雪球
				ticking = true
				this.UpdateFrame()
				ticking = false
			}
		}
	}()
	return shutdown
}

func (this *Runtime) UpdateFrame() {
	this.locker.RLock()
	defer this.locker.RUnlock()
	//内部系统间是有序的,awake->start->update->destroy
	for _, s := range this.innerSystems {
		s.UpdateFrame()
	}
	//自定义系统整体顺序在内部系统之后,是否需要同系统独立执行，在updateFrame接口实现
	for _, s := range this.customSystems {
		s.UpdateFrame()
	}
}

//注册自定义system
func (this *Runtime) RegisterSystem(system ISystem) {
	this.locker.Lock()
	defer this.locker.Unlock()
	//过滤重复系统
	for _, s := range this.customSystems {
		if s.Name() == system.Name() {
			return
		}
	}
	this.customSystems = append(this.customSystems, system)
}

func (this *Runtime) SystemFilter(component IComponent) {
	this.locker.RLock()
	defer this.locker.RUnlock()

	for _, ss := range this.innerSystems {
		ss.Filter(component)
	}
	for _, ss := range this.customSystems {
		ss.Filter(component)
	}
}

func (this *Runtime) SystemOperate(name string, op int, component IComponent) error {
	this.locker.RLock()
	defer this.locker.RUnlock()

	var s ISystem
	var ok bool
	for _, value := range this.innerSystems {
		if value.Name() == name {
			s = value
			ok = true
		}
	}
	if !ok {
		for _, value := range this.customSystems {
			if value.Name() == name {
				s = value
				ok = true
			}
		}
	}
	if ok && s != nil {
		s.IndependentFilter(op, component)
		return nil
	}
	return errors.New("system not found")
}

//----------------------------------------------------------------------------------------------------
func validateConfig(config *Config) {
	if config.ThreadPoolSize <= 0 {
		config.ThreadPoolSize = 10
	}
	if config.Factory == nil {
		config.Factory = NewObjectFactory()
	}
}

func (runtime *Runtime) Root() *Object {
	return runtime.root
}

func (runtime *Runtime) Factory() *ObjectFactory {
	return runtime.factory
}

func (runtime *Runtime) SetMaxThread(maxThread int) {
	if runtime.workers != nil && maxThread > 0 {
		runtime.workers.MaxThreads = maxThread
	} else {
		logger.Error(errors.New("max thread must > 0"))
	}
}

func (runtime *Runtime) Extract(object *Object) (*ObjectTemplate, error) {
	rtn, err := runtime.factory.Serialize(object)
	if err != nil {
		return nil, err
	}
	if object.parent != nil {
		if err = object.parent.RemoveObject(object); err != nil {
			return nil, err
		}
	}
	return rtn, nil
}

func (runtime *Runtime) Insert(template *ObjectTemplate, parent *Object) (*Object, error) {
	rtn, err := runtime.factory.Deserialize(template)
	if err != nil {
		return nil, err
	}
	if err := parent.AddObject(rtn); err != nil {
		return nil, err
	}
	return rtn, nil
}

func (runtime *Runtime) Objects() iter.Iter {
	return runtime.root.ObjectsInChildren()
}

func (runtime *Runtime) ScheduleTask(task func()) {
	go (func() {
		defer (func() {
			r := recover()
			if r != nil {
				err, ok := r.(error)
				if ok {
					logger.Error(fmt.Sprintf("Failed to execute scheduled task: %s", err.Error()))
				} else {
					logger.Error(fmt.Sprintf("Failed to execute scheduled task: %s", r))
				}
			}
		})()
		runtime.updateLock.Lock()
		task()
		defer runtime.updateLock.Unlock()
	})()
}
