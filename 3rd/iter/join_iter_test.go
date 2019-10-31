package iter_test

import (
	"container/list"
	"github.com/zllangct/RockGO/3rd/assert"
	"github.com/zllangct/RockGO/3rd/iter"
	"testing"
)

func testFixture() (iter.Iter, iter.Iter) {
	container := list.New()
	container.PushBack(100)
	container.PushBack(110)
	container.PushBack(111)
	i1 := iter.FromList(container)

	container2 := list.New()
	container.PushBack(200)
	container.PushBack(220)
	container.PushBack(222)
	i2 := iter.FromList(container2)

	return i1, i2
}

func TestJoin(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
		i1, i2 := testFixture()
		i3 := iter.Join(i1, i2)

		total := 0
		for val, err := i3.Next(); err == nil; val, err = i3.Next() {
			total += val.(int)
		}

		T.Assert(total == (321 + 642))


		_, err := i1.Next()
		T.Assert(err != nil)

		_, err = i2.Next()
		T.Assert(err != nil)

		_, err = i3.Next()
		T.Assert(err != nil)
	})
}
