package iter

import (
	"container/list"
)

// JoinIter is a chained list of iterators that can be enumerated as a single sequence.
type JoinIter struct {
	iterators *list.List
	cursor    *list.Element
	current   Iter
	err       error
}

// Join joins two iterators and returns the combined iterator
func Join(a Iter, b Iter) *JoinIter {
	rtn := &JoinIter{iterators: list.New()}
	rtn.Add(a)
	rtn.Add(b)
	return rtn
}

// Next increments the iterator cursor
func (iterator *JoinIter) Next() (interface{}, error) {
	if iterator.err != nil {
		return nil, iterator.err
	}
	return iterator.nextValue()
}

// Get next iterator from the iterator set if any
func (iterator *JoinIter) nextIterator() (Iter, error) {
	if iterator.cursor == nil {
		iterator.cursor = iterator.iterators.Front()
	} else {
		iterator.cursor = iterator.cursor.Next()
	}

	if iterator.cursor == nil {
		return nil, ErrEndIteration
	}

	next := iterator.cursor.Value.(Iter)
	return next, nil
}

// Get next value from the current iterator or error.
// If there is no current iterator, try to get the next one.
func (iterator *JoinIter) nextValue() (interface{}, error) {
	if iterator.current == nil {
		it, err := iterator.nextIterator()
		if err != nil {
			iterator.err = err
			return nil, iterator.err
		}
		iterator.current = it
	}

	value, err := iterator.current.Next()
	if err != nil {
		if err == ErrEndIteration {
			iterator.current = nil
			value, err = iterator.nextValue()
			if err != nil {
				iterator.err = err
				return nil, iterator.err
			}
		}
	}

	return value, nil
}

// Add a new iterator to this chain of iterators and return the JoinIter object.
func (iterator *JoinIter) Add(target Iter) *JoinIter {
	iterator.iterators.PushBack(target)
	return iterator
}
