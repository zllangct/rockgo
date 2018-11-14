package Component_test

import (
	"testing"

	"github.com/zllangct/RockGO/3RD/assert"
	"github.com/zllangct/RockGO/3RD/iter"
	"github.com/zllangct/RockGO/Component"
)

func TestSingleChildIterator(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
		obj := Component.NewObject("Object 1")
		obj2 := Component.NewObject("Object 2")
		obj.AddObject(obj2)

		results, err := iter.Collect(obj.ObjectsInChildren())
		T.Assert(err == nil)
		T.Assert(len(results) == 1)
		T.Assert(results[0] == obj2)
	})
}

func TestDepth3ChildIterator(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
		obj := Component.NewObject()
		obj2 := Component.NewObject()
		obj3 := Component.NewObject()
		obj.AddObject(obj2)
		obj2.AddObject(obj3)

		results, err := iter.Collect(obj.ObjectsInChildren())
		T.Assert(err == nil)
		T.Assert(len(results) == 2)
		T.Assert(results[0] == obj2)
		T.Assert(results[1] == obj3)
	})
}