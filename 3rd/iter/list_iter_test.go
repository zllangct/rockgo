package iter_test

import (
	"container/list"
	"github.com/zllangct/RockGO/3rd/assert"
	"github.com/zllangct/RockGO/3rd/iter"
	"testing"
)

func TestFromList(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
		container := list.New()
		container.PushBack(100)
		container.PushBack(110)
		container.PushBack(111)

		total := 0

		i := iter.FromList(container)
		for val, err := i.Next(); err == nil; val, err = i.Next() {
			total += val.(int)
		}

		T.Assert(total == 321)

		_, err := i.Next()
		T.Assert(err != nil)
	})
}
