package ecs_test

import (
	"github.com/zllangct/rockgo"
	"github.com/zllangct/rockgo/3rd/assert"
	"github.com/zllangct/rockgo/ecs"
	"testing"
)

func TestCannotMakeRecursiveObjects(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
		o1 := ecs.NewObject("A")
		o2 := ecs.NewObject("B")

		o1.AddObject(o2)
	})
}

func TestFindObject(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
		o1 := ecs.NewObject("A")
		o2 := ecs.NewObject("B")
		o3 := ecs.NewObject("C")
		o4 := ecs.NewObject("D1")
		o5 := ecs.NewObject("D2")

		o1.AddObject(ecs.NewObject())
		o1.AddObject(o2)
		o1.AddObject(ecs.NewObject())

		o2.AddObject(o3)

		o3.AddObject(o4)
		o3.AddObject(ecs.NewObject())
		o3.AddObject(o5)

		r, err := o1.FindObject("B")
		T.Assert(err == nil)
		T.Assert(r == o2)

		r, err = o1.FindObject("B", "C")
		T.Assert(err == nil)
		T.Assert(r == o3)

		r, err = o1.FindObject("B", "C", "D1")
		T.Assert(err == nil)
		T.Assert(r == o4)

		r, err = o1.FindObject("B", "C", "D2")
		T.Assert(err == nil)
		T.Assert(r == o5)
	})
}

func TestFindComponent(T *testing.T) {
	o1 := ecs.NewObject("A")
	o2 := ecs.NewObject("B")
	o3 := ecs.NewObject("C")
	o4 := ecs.NewObject("D")
	c1 := &FakeComponent{Id: "IComponent"}

	o1.AddObject(o2)
	o2.AddObject(o3)
	o3.AddObject(o4)
	o4.AddComponent(c1)

	var c2 *FakeComponent
	err := o1.Find(&c2, "B", "C", "D")
	_ = err
}

func TestFindComponentOnRoot(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
		o1 := ecs.NewObject("A")
		o2 := ecs.NewObject("B")
		o3 := ecs.NewObject("C")
		o4 := ecs.NewObject("D")
		c1 := &FakeComponent{Id: "IComponent"}

		o1.AddObject(o2)
		o2.AddObject(o3)
		o3.AddObject(o4)
		o1.AddComponent(c1)

		var c2 *FakeComponent
		err := o1.Find(&c2)

		T.Assert(err == nil)
		T.Assert(c2.Id == "IComponent")
	})
}

func TestModifyComponent(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
		o1 := ecs.NewObject("A")
		o2 := ecs.NewObject("B")
		o3 := ecs.NewObject("C")
		o4 := ecs.NewObject("D")
		c1 := &FakeComponent{Id: "IComponent"}

		o1.AddObject(o2)
		o2.AddObject(o3)
		o3.AddObject(o4)
		o4.AddComponent(c1)

		var c2 *FakeComponent
		err := o1.Find(&c2, "B", "C", "D")

		T.Assert(err == nil)
		T.Assert(c2.Id == "IComponent")

		c2.Count = 100

		var c3 *FakeComponent
		o1.Find(&c3, "B", "C", "D")

		T.Assert(err == nil)
		T.Assert(c2.Id == "IComponent")
		T.Assert(c3.Count == 100)
	})
}

func TestRemoveChild(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
		o1 := ecs.NewObject("A")
		o2 := ecs.NewObject("B")
		o3 := ecs.NewObject("C")
		o4 := ecs.NewObject("D")

		o1.AddObject(o2)
		o2.AddObject(o3)
		o3.AddObject(o4)
		o3.RemoveObject(o4)

		_, err := o3.GetObject("D")
		T.Assert(err != nil)
	})
}

func TestRemoveChildWithChildren(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
		o1 := ecs.NewObject("A")
		o2 := ecs.NewObject("B")
		o3 := ecs.NewObject("C")
		o4 := ecs.NewObject("D")

		o1.AddObject(o2)
		o2.AddObject(o3)
		o3.AddObject(o4)
		o2.RemoveObject(o3)

		_, err := o2.GetObject("C")
		T.Assert(err != nil)

		_, err = o3.GetObject("D")
		T.Assert(err == nil)
	})
}
