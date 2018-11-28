package iter_test

import (
	"container/list"
	"github.com/zllangct/RockGO/3rd/assert"
	"github.com/zllangct/RockGO/3rd/iter"
	"testing"
)

func TestCount(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
		container := list.New()
		container.PushBack(100)
		container.PushBack(110)
		container.PushBack(111)

		i := iter.FromList(container)
		count, err := iter.Count(i)
		T.Assert(err == nil)
		T.Assert(count == 3)

		count, err = iter.Count(i)
		T.Assert(err == nil)
		T.Assert(count == 0)
	})
}

func TestCollect(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
		container := list.New()
		container.PushBack(100)
		container.PushBack(110)
		container.PushBack(111)

		i := iter.FromList(container)
		all, err := iter.Collect(i)
		T.Assert(err == nil)
		T.Assert(all != nil)
		T.Assert(len(all) == 3)

		all, err = iter.Collect(i)
		T.Assert(err == nil)
		T.Assert(len(all) == 0)
	})
}
