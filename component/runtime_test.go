package Component_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/zllangct/RockGO/3rd/assert"
	"github.com/zllangct/RockGO/3rd/iter"
	"github.com/zllangct/RockGO/component"
)

func TestNew(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
		Component.NewRuntime(Component.Config{
			ThreadPoolSize: 3})
	})
}

func TestUpdate(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
		runtime := Component.NewRuntime(Component.Config{
			ThreadPoolSize: 50})

		obj := Component.NewObject()
		obj.AddComponent(&FakeComponent{})
		obj.AddComponent(&FakeComponent{})
		obj.AddComponent(&FakeComponent{})

		count, err := iter.Count(obj.GetComponents(reflect.TypeOf((*FakeComponent)(nil))))
		T.Assert(err == nil)
		T.Assert(count == 3)

		components := obj.GetComponents(reflect.TypeOf((*FakeComponent)(nil)))
		for val, err := components.Next(); err == nil; val, err = components.Next() {
			T.Assert(val.(*FakeComponent).Count == 0)
		}

		runtime.Root().AddObject(obj)

		runtime.UpdateFrameByInterval(time.Second * 1)

		components = obj.GetComponents(reflect.TypeOf((*FakeComponent)(nil)))
		for val, err := components.Next(); err == nil; val, err = components.Next() {
			T.Assert(val.(*FakeComponent).Count == 2)
		}
	})
}

func TestComponentsAreUpdated(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
		runtime := Component.NewRuntime(Component.Config{
			ThreadPoolSize: 50})

		obj := Component.NewObject()
		obj.AddComponent(&FakeComponent{Id: "1"})
		obj.AddComponent(&FakeComponent{Id: "2"})
		obj.AddComponent(&FakeComponent{Id: "3"})

		root := Component.NewObject()
		root.AddComponent(&FakeComponent{Id: "4"})

		root.AddObject(obj)

		//obj1:=Component.NewObject()
		//obj1.AddComponent(&FakeComponent{Id:"5"})
		//
		//obj.AddObject(obj1)

		count, _ := iter.Count(root.ObjectsInChildren())
		T.Assert(count == 1)

		count, err := iter.Count(root.GetComponentsInChildren(reflect.TypeOf((*FakeComponent)(nil))))
		T.Assert(err == nil)
		T.Assert(count == 3)

		runtime.UpdateFrameByInterval(time.Second * 1)

		components := root.GetComponentsInChildren(reflect.TypeOf((*FakeComponent)(nil)).Elem())
		for val, err := components.Next(); err == nil; val, err = components.Next() {
			T.Assert(val.(*FakeComponent).Count == 2)
		}

		root.Destroy()
	})
}
