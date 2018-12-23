package iter

import "errors"

// ErrEndIteration is raised when an iterator is out of values.
var ErrEndIteration =errors.New("iterator is out of values")
