package assert

import (
	"fmt"
	"runtime/debug"
	"testing"
)

// T is a light weight testing helper
type T struct {
	test   *testing.T
	failed bool
}

// Test executes the given testcase passing a new testing
// instance into it.
func Test(t *testing.T, testcase func(*T)) {
	defer func() {
		r := recover()
		if r != nil {
			fmt.Printf("Panic running test: %s\n", r)
			debug.PrintStack()
			fmt.Printf("\n")
			t.Fail()
		}
	}()
	tester := New(t)
	testcase(tester)
	tester.Commit()
}

// New constructs a new testing instance; however, the test case will
// not fail the test until `Commit` is invoked; you should use `Assert`
// for this purpose unless doing something complex.
func New(t *testing.T) *T {
	return &T{t, false}
}

// Failed returns true if the test has failed
func (T *T) Failed() bool {
	return T.failed
}

// Commit fails the current test case if an assertion has failed.
func (T *T) Commit() {
	if T.failed {
		T.test.Fail()
	}
}

// Assert marks the test case failed if not true
func (T *T) Assert(truth bool) {
	if !truth {
		T.failed = true
		fmt.Printf("Assertion failed:\n")
		debug.PrintStack()
		fmt.Printf("\n")
	}
}

// Unreachable fails the test if it gets hit.
func (T *T) Unreachable() {
	T.failed = true
	fmt.Printf("Unreachable bound hit:\n")
	debug.PrintStack()
	fmt.Printf("\n")
}
