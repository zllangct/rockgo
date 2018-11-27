package Component

import (
	"errors"
	"fmt"
	"github.com/zllangct/RockGO/logger"
	"sync"

	"github.com/zllangct/RockGO/3RD/iter"
	"github.com/zllangct/RockGO/3RD/threadpool"
)

// Config configures a runtime.
type Config struct {
	ThreadPoolSize int
	Factory        *ObjectFactory
}

// Runtime is the basic operating unit of the mud.
// A Runtime executes the main game loop on objects.
type Runtime struct {
	root       *Object                // The root object for this runtime.
	workers    *threadpool.ThreadPool // The thread ConnPool for updating objects
	updateLock *sync.Mutex            // The thread safe lock for updates.
	factory    *ObjectFactory         // The serialization factory
}

// New returns a new Runtime instance
func NewRuntime(config Config) *Runtime {
	validateConfig(&config)
	runtime := &Runtime{
		root:       NewObject(),
		updateLock: &sync.Mutex{},
		workers:    threadpool.New(),
		factory:    config.Factory}
	runtime.root.runtime = runtime
	runtime.workers.MaxThreads = config.ThreadPoolSize
	return runtime
}

// Configure sensible defaults if none are provided
func validateConfig(config *Config) {
	if config.ThreadPoolSize <= 0 {
		config.ThreadPoolSize = 10
	}
	if config.Factory == nil {
		config.Factory = NewObjectFactory()
	}
}

// Return a reference to the root object for the runtime
func (runtime *Runtime) Root() *Object {
	return runtime.root
}

// Factory returns the object factory for the runtime
func (runtime *Runtime) Factory() *ObjectFactory {
	return runtime.factory
}

func (runtime *Runtime)SetMaxThread(maxThread int) error {
	if runtime.workers!=nil && maxThread > 0{
		runtime.workers.MaxThreads =maxThread
	}else{
		return errors.New("max thread must > 0")
	}
	return nil
}

// Extract creates a deep copy of the object and then removes it from the runtime.
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

// Insert converts the template into an object and attaches it as a child of the given parent.
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

// Return the set of objects as an iterator, including root.
func (runtime *Runtime) Objects() iter.Iter {
	return runtime.root.ObjectsInChildren()
}

// Schedules a task to execute between the next update loops.
// Return immediately, but the task will only be executed after
// the current IUpdate() loop finishes.
// This effectively blocks until the current loop ends, then runs; then
// finally returns.
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

// Execute the update step of all components on all objects in worker threads
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

// Execute a single object update
func (runtime *Runtime) updateObject(step float32, obj *Object) {
	runtime.workers.Run(func() {
		if obj.runtime == nil {
			obj.runtime = runtime
		}
		obj.Update(step, runtime)
	})
}
