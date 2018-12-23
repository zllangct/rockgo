package Component

import (
	"errors"
	"fmt"
	"github.com/zllangct/RockGO/logger"
	"sync"

	"github.com/zllangct/RockGO/3rd/iter"
	"github.com/zllangct/RockGO/3rd/threadpool"
)


type Config struct {
	ThreadPoolSize int
	Factory        *ObjectFactory
}

type Runtime struct {
	root       *Object
	workers    *threadpool.ThreadPool
	updateLock *sync.Mutex
	factory    *ObjectFactory
}

func NewRuntime(config Config) *Runtime {
	validateConfig(&config)
	runtime := &Runtime{
		root:       NewObject("runtime"),
		updateLock: &sync.Mutex{},
		workers:    threadpool.New(),
		factory:    config.Factory}
	runtime.root.runtime = runtime
	runtime.workers.MaxThreads = config.ThreadPoolSize
	return runtime
}

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

func (runtime *Runtime)SetMaxThread(maxThread int){
	if runtime.workers!=nil && maxThread > 0{
		runtime.workers.MaxThreads =maxThread
	}else{
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

func (runtime *Runtime) Update(step float32) {
	runtime.updateLock.Lock()
	defer runtime.updateLock.Unlock()
	runtime.updateObject(step, runtime.root)
	objects := runtime.Objects()
	if objects != nil {
		var val interface{}
		var err error
		for val, err = objects.Next(); err == nil; val, err = objects.Next() {
			obj := val.(*Object)
			runtime.updateObject(step, obj)
		}
		runtime.workers.Wait()
	}
}

func (runtime *Runtime) updateObject(step float32, obj *Object) {
	runtime.workers.Run(func() {
		if obj.runtime == nil {
			obj.runtime = runtime
		}
		obj.Update(step, runtime)
	})
}
