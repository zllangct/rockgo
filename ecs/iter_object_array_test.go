package ecs_test

import (
	"testing"

	"github.com/zllangct/rockgo/3rd/assert"
	"github.com/zllangct/rockgo/3rd/iter"
	"github.com/zllangct/rockgo/ecs"
)

func TestSingleChildIterator(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
		obj := ecs.NewObject("Object 1")
		obj2 := ecs.NewObject("Object 2")
		obj.AddObject(obj2)

		results, err := iter.Collect(obj.ObjectsInChildren())
		T.Assert(err == nil)
		T.Assert(len(results) == 1)
		T.Assert(results[0] == obj2)
	})
}

func TestDepth3ChildIterator(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
		obj := ecs.NewObject()
		obj2 := ecs.NewObject()
		obj3 := ecs.NewObject()
		obj.AddObject(obj2)
		obj2.AddObject(obj3)

		results, err := iter.Collect(obj.ObjectsInChildren())
		T.Assert(err == nil)
		T.Assert(len(results) == 2)
		T.Assert(results[0] == obj2)
		T.Assert(results[1] == obj3)
	})
}
