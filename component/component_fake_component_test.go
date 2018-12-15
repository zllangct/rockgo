package Component_test

import (
	"reflect"
	"fmt"
	"strings"
	"strconv"
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/3rd/errors"
)

type FakeComponent struct {
	Id    string
	Count int
}

func (fake *FakeComponent) Type() reflect.Type {
	return reflect.TypeOf(fake)
}

func (fake *FakeComponent) Update(_ *Component.Context) {
	fake.Count += 1
}

func (fake *FakeComponent) Destroy()error {
	println(fake.Id)
	return nil
}

func (fake *FakeComponent) New() Component.IComponent {
	return &FakeComponent{}
}

func (fake *FakeComponent) Serialize() (interface{}, error) {
	return fmt.Sprintf("%s,%d", fake.Id, fake.Count), nil
}

func (fake *FakeComponent) Deserialize(raw interface{}) error {
	if raw == nil {
		return nil
	}
	data := raw.(string)
	if len(data) > 0 {
		parts := strings.Split(data, ",")
		if len(parts) != 2 {
			return errors.Fail(Component.ErrBadValue{}, nil, "Bad component data")
		}
		fake.Id = parts[0]
		count, err := strconv.Atoi(parts[1])
		if err != nil {
			return err
		}
		fake.Count = count
	}

	return nil
}
