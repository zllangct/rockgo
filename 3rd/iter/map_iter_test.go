package iter_test

import (
	"errors"
	"github.com/zllangct/RockGO/3rd/assert"
	"testing"
)

type MapIter struct {
	values map[string]int
	keys   []string
	offset int
	err    error
}

type MapIterRecord struct {
	Key   string
	Value int
}

func NewMapIter(values map[string]int) *MapIter {
	rtn := &MapIter{values: values}
	rtn.Init()
	return rtn
}

func (iterator *MapIter) Init() {
	iterator.offset = -1
	iterator.keys = make([]string, 0, len(iterator.values))
	for k := range iterator.values {
		iterator.keys = append(iterator.keys, k)
	}
}

func (iterator *MapIter) Next() (interface{}, error) {
	if iterator.err != nil {
		return nil, iterator.err
	}

	iterator.offset += 1
	if iterator.offset >= len(iterator.keys) {
		iterator.err = errors.New("no more values")
		return nil, iterator.err
	}

	keyValuePair := &MapIterRecord{
		iterator.keys[iterator.offset],
		iterator.values[iterator.keys[iterator.offset]]}
	return keyValuePair, nil
}

func TestMapIter(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
		container := make(map[string]int)
		container["one"] = 1
		container["two"] = 2
		container["three"] = 3
		container["four"] = 4

		total := 0

		i := NewMapIter(container)
		for val, err := i.Next(); err == nil; val, err = i.Next() {
			keyValue := val.(*MapIterRecord)
			total += keyValue.Value
		}

		T.Assert(total == (1 + 2 + 3 + 4))

		_, err := i.Next()
		T.Assert(err != nil)
	})
}
