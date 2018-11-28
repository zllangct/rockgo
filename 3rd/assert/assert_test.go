package assert_test

import "testing"
import "github.com/zllangct/RockGO/3rd/assert"

func TestNew(T *testing.T) {
	assert.New(T)
}

func TestAssert(T *testing.T) {
	assert := assert.New(T)
	assert.Assert(true)
	if assert.Failed() {
		T.Fail()
	}
	assert.Assert(false)
	if !assert.Failed() {
		T.Fail()
	}
}

func TestUnreachable(T *testing.T) {
	assert := assert.New(T)
	assert.Unreachable()
	if !assert.Failed() {
		T.Fail()
	}
}

func TestTest(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
		T.Assert(1+1 == 2)
		T.Assert(true)
		if false {
			T.Unreachable()
		}
	})
}
