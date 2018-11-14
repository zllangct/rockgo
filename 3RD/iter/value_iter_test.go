package iter_test

import (
	"github.com/zllangct/RockGO/3RD/assert"
	"github.com/zllangct/RockGO/3RD/errors"
	"github.com/zllangct/RockGO/3RD/iter"
	"testing"
)

func TestFromValue(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
		total := 0

		i := iter.FromValue(1)
		for val, err := i.Next(); err == nil; val, err = i.Next() {
			total += val.(int)
		}

		T.Assert(total == 1)

		_, err := i.Next()
		T.Assert(err != nil)
		T.Assert(errors.Is(err, iter.ErrEndIteration{}))
	})
}
