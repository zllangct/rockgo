package ecs_test

import (
	"fmt"
	"github.com/zllangct/rockgo/ecs"
	"strconv"
	"strings"
)

type FakeComponent struct {
	ecs.ComponentBase
	Id    string
	Count int
}

func (fake *FakeComponent) Update(context *ecs.Context) {
	fake.Count += 1
}

func (fake *FakeComponent) Destroy(context *ecs.Context) {
	println(fake.Id)
}

func (fake *FakeComponent) New() ecs.IComponent {
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
			return ecs.ErrBadValue
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
