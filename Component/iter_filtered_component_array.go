package Component

import (

	"container/list"
	"reflect"
	"github.com/zllangct/RockGO/3RD/iter"
	"github.com/zllangct/RockGO/3RD/errors"
)

// FilterComponentArrayIter implements Iterator for components with a type filter.
type FilterComponentArrayIter struct {
	target  reflect.Type
	values  *list.List
	current *[]*componentInfo
	offset  int
	err     error
}

// fromComponentArray returns a new list iterator for a list
func fromComponentArray(values *[]*componentInfo, T reflect.Type) *FilterComponentArrayIter {
	rtn := &FilterComponentArrayIter{values: list.New(), offset: -1, target: T}
	rtn.Add(values)
	return rtn
}

// Next increments the iterator cursor
func (iterator *FilterComponentArrayIter) Next() (interface{}, error) {
	if iterator.err != nil {
		return nil, iterator.err
	}

	// Look for a matching type
	var cmp Component = nil
	for iterator.err == nil {
		if iterator.err != nil {
			return nil, iterator.err
		}

		if iterator.current == nil {
			if iterator.nextGroup() {
				return nil, iterator.err
			}
		} else {
			iterator.offset += 1
		}

		if iterator.offset >= len(*iterator.current) {
			if iterator.nextGroup() {
				return nil, iterator.err
			}
		}

		value := (*iterator.current)[iterator.offset]
		if value.Type == iterator.target {
			cmp = value.Component
			break
		}
	}
	return cmp, nil
}

// Add another set of components to search through
func (iterator *FilterComponentArrayIter) Add(values *[]*componentInfo) {
	if values != nil {
		iterator.values.PushBack(values)
	}
}

// Attempt to get the next set if its null.
// Each set is the set of objects from a different parent.
// Returns true if an error is set.
func (iterator *FilterComponentArrayIter) nextGroup() bool {
	el := iterator.values.Front()
	if el != nil {
		iterator.values.Remove(el)
		iterator.current = el.Value.(*[]*componentInfo)
		iterator.offset = 0
	} else {
		iterator.err = errors.Fail(iter.ErrEndIteration{}, nil, "No more values")
		return true
	}
	return false
}