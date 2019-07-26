package ecs

import (
	"container/list"
	"github.com/zllangct/RockGO/3rd/iter"
	"reflect"
)

// FilterComponentArrayIter implements Iterator for components with a type filter.
type FilterComponentArrayIter struct {
	target  reflect.Type
	values  *list.List
	current *[]IComponent
	offset  int
	err     error
}

// fromComponentArray returns a new list iterator for a list
func fromComponentArray(values *[]IComponent, T reflect.Type) *FilterComponentArrayIter {
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
	var cmp IComponent = nil
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
		if iterator.target == nil {
			cmp = value
			break
		}
		//支持通过实现过的接口类型查询,接口查询效率低下，慎重使用。
		cmpTye := value.Type()
		if cmpTye == iterator.target || (cmpTye.Kind() == reflect.Interface && cmpTye.Implements(iterator.target.Elem())) {
			cmp = value
			break
		}
	}
	return cmp, nil
}

// Add another set of components to search through
func (iterator *FilterComponentArrayIter) Add(values *[]IComponent) {
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
		iterator.current = el.Value.(*[]IComponent)
		iterator.offset = 0
	} else {
		iterator.err = iter.ErrEndIteration
		return true
	}
	return false
}
