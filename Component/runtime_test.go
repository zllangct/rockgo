package Component_test

import (
	"reflect"
	"testing"

	"github.com/zllangct/RockGO/3RD/assert"
	"github.com/zllangct/RockGO/3RD/iter"
	"github.com/zllangct/RockGO/Component"
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

		runtime.Update(1.0)
		runtime.Update(1.0)

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
		obj.AddComponent(&FakeComponent{})
		obj.AddComponent(&FakeComponent{})
		obj.AddComponent(&FakeComponent{})

		root := Component.NewObject()
		root.AddObject(obj)

		count, _ := iter.Count(root.ObjectsInChildren())
		T.Assert(count == 1)

		count, err := iter.Count(root.GetComponentsInChildren(reflect.TypeOf((*FakeComponent)(nil))))
		T.Assert(err == nil)
		T.Assert(count == 3)

		runtime.Update(1.0)
		runtime.Update(1.0)

		components := root.GetComponentsInChildren(reflect.TypeOf((*FakeComponent)(nil)).Elem())
		for val, err := components.Next(); err == nil; val, err = components.Next() {
			T.Assert(val.(*FakeComponent).Count == 2)
		}
	})
}