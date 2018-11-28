package errors

import (
	"fmt"
	"reflect"
)

// Error is a basic common error type that wraps a message, a code and
// an optional inner error.
type Error struct {
	Reason string
	Inner  error
	Type   reflect.Type
}

// Error interface
func (err Error) Error() string {
	if err.Inner == nil {
		return fmt.Sprintf("(%v) %s", err.Type, err.Reason)
	}
	return fmt.Sprintf("(%v) %s: %s", err.Type, err.Reason, err.Inner)
}

// Fail returns an error with a message.
func Fail(etype interface{}, inner error, format string, args ...interface{}) error {
	return Error{
		fmt.Sprintf(format, args...),
		inner,
		reflect.TypeOf(etype),
	}
}

// Is checks for the error type on a standard error
func Is(err error, etype interface{}) bool {
	stderr, ok := err.(Error)
	if ok {
		return stderr.Type == reflect.TypeOf(etype)
	}
	return false
}

// Inner returns the inner error or err and if there was one
func Inner(err error) (error, bool) {
	stderr, ok := err.(Error)
  if ok {
    return stderr.Inner, true
  }
  return nil, false
}

// ErrData is an error wrapper for an arbitrary data object
type ErrData struct {
	Data interface{}
}

// Error interface
func (err ErrData) Error() string {
	return fmt.Sprintf("ErrData: %v", err.Data)
}

// Data returns a new ErrData error holding the given data only.
// You can use this as an inner exception; for example:
// errors.Fail(ErrMyErr{}, errors.Data(dataObject), "Some info")
func Data(data interface{}) error {
  return ErrData{Data: data}
}
