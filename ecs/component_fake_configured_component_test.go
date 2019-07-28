package ecs_test

import (
	"github.com/zllangct/RockGO/ecs"
	"reflect"
)

type FakeConfiguredComponentData struct {
	Items []FakeConfiguredComponentItem
}

type FakeConfiguredComponentItem struct {
	Id    string
	Count int
}

type FakeConfiguredComponent struct {
	ecs.ComponentBase
	Data FakeConfiguredComponentData
}

func (fake *FakeConfiguredComponent) Type() reflect.Type {
	return reflect.TypeOf(fake)
}

func (fake *FakeConfiguredComponent) New() ecs.IComponent {
	return &FakeConfiguredComponent{}
}

func (fake *FakeConfiguredComponent) Serialize() (interface{}, error) {
	return ecs.SerializeState(&fake.Data)
}

func (fake *FakeConfiguredComponent) Deserialize(raw interface{}) error {
	var data FakeConfiguredComponentData
	if err := ecs.DeserializeState(&data, raw); err != nil {
		return err
	}
	fake.Data = data
	return nil
}
