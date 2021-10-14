package ecs_test

import (
	"reflect"

	"github.com/zllangct/rockgo/ecs"
)

type ComponentTemplate struct {
	ecs.ComponentBase
	parent *ecs.Object
}

func (c *ComponentTemplate) New() ecs.IComponent {
	return &ComponentTemplate{}
}

func (c *ComponentTemplate) Type() reflect.Type {
	return reflect.TypeOf(c)
}

func (c *ComponentTemplate) Awake(parent *ecs.Object) {
	c.parent = parent
}

func (c *ComponentTemplate) Update(context *ecs.Context) {
}

func (c *ComponentTemplate) Serialize() (interface{}, error) {
	return "", nil
}

func (c *ComponentTemplate) Deserialize(raw interface{}) error {
	return nil
}
