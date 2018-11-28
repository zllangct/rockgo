package errors_test

import (
	"ntoolkit/assert"
	"ntoolkit/errors"
	"testing"
)

type ErrCode1 struct{}
type ErrCode2 struct{}

func dummy(value int) (int, error) {
	if value == 0 {
		return 0, errors.Fail(ErrCode1{}, nil, "Invalid value: %d", value)
	}
	return value + 1, nil
}

func dummy_safe(value interface{}) (rval int, rerr error) {
	defer func() {
		if err := recover(); err != nil {
			rval = 0
			rerr = errors.Fail(ErrCode2{}, err.(error), "Wrapped inner error")
		}
	}()
	return value.(int) + 1, nil
}

func TestNoError(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
		v, err := dummy(1)
		T.Assert(v == 2)
		T.Assert(err == nil)
	})
}

func TestError(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
		v, err := dummy(0)
		T.Assert(v == 0)
		T.Assert(err != nil)
		T.Assert(errors.Is(err, ErrCode1{}))
		T.Assert(!errors.Is(err, ErrCode2{}))
	})
}

func TestRecover(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
		v, err := dummy_safe(0)
		T.Assert(v == 1)
		T.Assert(err == nil)

		v, err = dummy_safe("hi")
		T.Assert(v == 0)
		T.Assert(err != nil)
		T.Assert(errors.Is(err, ErrCode2{}))
		T.Assert(err.Error() == "(errors_test.ErrCode2) Wrapped inner error: interface conversion: interface {} is string, not int")
	})
}

func TestInner(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
		err := errors.Fail(ErrCode1{}, errors.Fail(ErrCode2{}, nil, "Inner"), "Output")
		inner, ok := errors.Inner(err)
		T.Assert(ok)
		T.Assert(errors.Is(inner, ErrCode2{}))
	})
}

func TestErrData(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
    err := errors.Fail(ErrCode1{}, errors.Data("Some error specific data"), "Output")
    output := err.Error()
    T.Assert(output == "(errors_test.ErrCode1) Output: ErrData: Some error specific data")
	})
}
