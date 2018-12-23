package iter

import (
	"container/list"
)

// ListIter implements Iterator for list.List
type ListIter struct {
	container *list.List
	cursor    *list.Element
	err       error
}

// FromList returns a new list iterator for a list
func FromList(container *list.List) *ListIter {
	return &ListIter{container: container}
}

// Next increments the iterator cursor
func (iterator *ListIter) Next() (interface{}, error) {
	if iterator.err != nil {
		return nil, iterator.err
	}

	if iterator.cursor == nil {
		iterator.cursor = iterator.container.Front()
	} else {
		iterator.cursor = iterator.cursor.Next()
	}

	if iterator.cursor == nil {
		iterator.err = ErrEndIteration
		return nil, iterator.err
	}

	return iterator.cursor.Value, nil
}
