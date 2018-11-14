package Component

import (

	"container/list"
	"github.com/zllangct/RockGO/3RD/iter"
	"github.com/zllangct/RockGO/3RD/errors"
)

// ObjectIter implements Iterator for []Object
type ObjectIter struct {
	values *list.List
	decend bool
	err    error
}

// fromObjectArray returns a new list iterator for a list
func fromObject(root *Object, decend bool) *ObjectIter {
	rtn := &ObjectIter{values: list.New(), decend: decend}
	if root == nil {
		rtn.err = errors.Fail(ErrNullValue{}, nil, "Invalid root object")
	} else if len(root.children) == 0 {
		rtn.err = errors.Fail(iter.ErrEndIteration{}, nil, "No more values")
	} else {
		for i := 0; i < len(root.children); i++ {
			rtn.values.PushBack(root.children[i])
		}
	}
	return rtn
}

// Next increments the iterator cursor
func (iterator *ObjectIter) Next() (interface{}, error) {
	if iterator.err != nil {
		return nil, iterator.err
	}

	obj := iterator.nextObject()
	if obj == nil {
		return nil, iterator.err
	}

	return obj, nil
}

// Attempt to get the next set if its null.
// Each set is the set of objects from a different parent.
// Returns true if an error is set.
func (iterator *ObjectIter) nextObject() *Object {
	el := iterator.values.Front()
	if el != nil {
		iterator.values.Remove(el)
		obj := el.Value.(*Object)
		if iterator.decend {
			for i := 0; i < len(obj.children); i++ {
				iterator.values.PushBack(obj.children[i])
			}
		}
		return obj
	} else {
		iterator.err = errors.Fail(iter.ErrEndIteration{}, nil, "No more values")
		return nil
	}
}