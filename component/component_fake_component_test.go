package Component_test

import (
	"fmt"
	"github.com/zllangct/RockGO/component"
	"strconv"
	"strings"
)

type FakeComponent struct {
	Component.Base
	Id    string
	Count int
}

func (fake *FakeComponent) Update(context *Component.Context) {
	fake.Count += 1
}

func (fake *FakeComponent) Destroy(context *Component.Context) {
	println(fake.Id)
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
			return Component.ErrBadValue
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
