package Component_test

import (
	"reflect"

	"github.com/zllangct/RockGO/Component"
)

type ComponentTemplate struct {
	parent *Component.Object
}

func (c *ComponentTemplate) New() Component.Component {
	return &ComponentTemplate{}
}

func (c *ComponentTemplate) Type() reflect.Type {
	return reflect.TypeOf(c)
}

func (c *ComponentTemplate) Awake(parent *Component.Object) {
	c.parent = parent
}

func (c *ComponentTemplate) Update(context *Component.Context) {
}

func (c *ComponentTemplate) Serialize() (interface{}, error) {
	return "", nil
}

func (c *ComponentTemplate) Deserialize(raw interface{}) error {
	return nil
}
