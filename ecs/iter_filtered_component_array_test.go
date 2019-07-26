package ecs_test

import (
	"testing"

	"github.com/zllangct/RockGO/3rd/assert"
	"github.com/zllangct/RockGO/3rd/iter"
	"github.com/zllangct/RockGO/ecs"
	"reflect"
)

func TestGetComponents(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
		obj := ecs.NewObject("Object 1")
		obj.AddComponent(&FakeComponent{Id: "1"})
		obj.AddComponent(&FakeComponent{Id: "1"})
		os := obj.GetComponents(reflect.TypeOf((ecs.IComponent)(nil)))
		ci, err := iter.Collect(os)
		t2 := reflect.TypeOf((*FakeComponent)(nil))
		_ = t2
		//ci, err := iter.Collect(obj.GetComponents(reflect.TypeOf((*FakeComponent)(nil))))
		T.Assert(err == nil)
		T.Assert(len(ci) == 1)
		T.Assert(ci[0].(*FakeComponent).Id == "1")
	})
}

func TestGetComponentsInChildren(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
		obj := ecs.NewObject()
		obj2 := ecs.NewObject()
		obj3 := ecs.NewObject()
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
