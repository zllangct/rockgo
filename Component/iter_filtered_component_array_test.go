package Component_test

import (
	"testing"

	"reflect"
	"github.com/zllangct/RockGO/3RD/assert"
	"github.com/zllangct/RockGO/Component"
	"github.com/zllangct/RockGO/3RD/iter"
)

func TestGetComponents(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
		obj := Component.NewObject("Object 1")
		obj.AddComponent(&FakeComponent{Id: "1"})
		obj.AddComponent(&FakeComponent{Id: "1"})
		ci, err := iter.Collect(obj.GetComponents(reflect.TypeOf((*FakeComponent)(nil))))
		T.Assert(err == nil)
		T.Assert(len(ci) == 1)
		T.Assert(ci[0].(*FakeComponent).Id == "1")
	})
}

func TestGetComponentsInChildren(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
		obj := Component.NewObject()
		obj2 := Component.NewObject()
		obj3 := Component.NewObject()
		obj.AddObject(obj2)
		obj2.AddObject(obj3)

		obj3.AddComponent(&FakeComponent{Id: "1"})
		obj3.AddComponent(&FakeComponent{Id: "2"})

		ci, err := iter.Collect(obj.GetComponents(reflect.TypeOf((*FakeComponent)(nil))))
		T.Assert(err == nil)
		T.Assert(len(ci) == 0)

		ci, err = iter.Collect(obj.GetComponentsInChildren(reflect.TypeOf((*FakeComponent)(nil))))
		T.Assert(err == nil)
		T.Assert(len(ci) == 2)
		T.Assert(ci[0].(*FakeComponent).Id == "1")
		T.Assert(ci[1].(*FakeComponent).Id == "2")
	})
}