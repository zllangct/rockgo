package Component_test

import (
	"reflect"
	"github.com/zllangct/RockGO/component"
)

type FakeConfiguredComponentData struct {
	Items []FakeConfiguredComponentItem
}

type FakeConfiguredComponentItem struct {
	Id    string
	Count int
}

type FakeConfiguredComponent struct {
	Component.Base
	Data FakeConfiguredComponentData
}

func (fake *FakeConfiguredComponent) Type() reflect.Type {
	return reflect.TypeOf(fake)
}

func (fake *FakeConfiguredComponent) New() Component.IComponent {
	return &FakeConfiguredComponent{}
}

func (fake *FakeConfiguredComponent) Serialize() (interface{}, error) {
	return Component.SerializeState(&fake.Data)
}

func (fake *FakeConfiguredComponent) Deserialize(raw interface{}) error {
	var data FakeConfiguredComponentData
	if err := Component.DeserializeState(&data, raw); err != nil {
		return err
	}
	fake.Data = data
	return nil
}