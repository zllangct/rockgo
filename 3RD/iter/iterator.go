package iter

// Iter is the basic iterator interface that objects can implement.
type Iter interface {

	// Value return raw the current value of the iterator or nil an error
	// If the iterator has run out of values, return iter.ErrEndIteration.
	Next() (interface{}, error)
}
