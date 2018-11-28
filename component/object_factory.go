package Component

import (
	"reflect"
	"fmt"
	"strings"
	"github.com/zllangct/RockGO/3rd/errors"
)
// ComponentProvider maps between component instances and component templates
type ComponentProvider interface {
	Type() reflect.Type
	New() IComponent
}

// ObjectFactory is the overseer that can be used to convert between objects and object templates
type ObjectFactory struct {
	handlers map[string]ComponentProvider
}

// NewObjectFactory returns a new object factory
func NewObjectFactory() *ObjectFactory {
	return &ObjectFactory{handlers: make(map[string]ComponentProvider)}
}

// Register a ComponentProvider that can be used to serialize and deserialize objects
func (factory *ObjectFactory) Register(provider ComponentProvider) {
	factory.handlers[typeName(provider.Type())] = provider
}

// Serialize converts an object into an ObjectTemplate
func (factory *ObjectFactory) Serialize(object *Object) (*ObjectTemplate, error) {
	obj := &ObjectTemplate{Name: object.name}

	// Assign each component
	for i := 0; i < len(object.components); i++ {
		c, err := factory.serializeComponent(object.components[i])
		if err != nil {
			return nil, err
		}
		obj.Components = append(obj.Components, *c)
	}

	// Assign each object
	for i := 0; i < len(object.children); i++ {
		o, err := factory.Serialize(object.children[i])
		if err != nil {
			return nil, err
		}
		obj.Objects = append(obj.Objects, *o)
	}

	return obj, nil
}

// Deserialize converts an ObjectTemplate into an object
func (factory *ObjectFactory) Deserialize(template *ObjectTemplate) (*Object, error) {
	obj := NewObject(template.Name)

	// Add components
	for i := 0; i < len(template.Components); i++ {
		c, err := factory.deserializeComponent(&template.Components[i])
		if err != nil {
			return nil, err
		}
		obj.AddComponent(c)
	}

	// Add children
	for i := 0; i < len(template.Objects); i++ {
		child, err := factory.Deserialize(&template.Objects[i])
		if err != nil {
			return nil, err
		}
		obj.AddObject(child)
	}

	return obj, nil
}

// deserializeComponent turns a component template into a component
func (factory *ObjectFactory) deserializeComponent(template *ComponentTemplate) (IComponent, error) {
	for k, v := range factory.handlers {
		if k == template.Type {
			component := v.New()
			if component.Type().Implements(reflect.TypeOf((*IPersist)(nil)).Elem()) {
				err := component.(IPersist).Deserialize(template.Data)
				if err != nil {
					return nil, err
				}
			}
			return component, nil
		}
	}
	return nil, errors.Fail(ErrUnknownComponent{}, nil, fmt.Sprintf("IComponent type %s is not registered with the factory", template.Type))
}

// serializeComponent converts a component into a template
func (factory *ObjectFactory) serializeComponent(component *componentInfo) (*ComponentTemplate, error) {
	template := &ComponentTemplate{
		Type: typeName(component.Type)}
	if component.Persist != nil {
		data, err := component.Persist.Serialize()
		if err != nil {
			return nil, err
		}
		template.Data = data
	}
	return template, nil
}

// typeName returns the name for a specific type
func typeName(T reflect.Type) string {
	pkgPath := ""
	typeName := ""
	isPtr := false
	if T.Kind() == reflect.Ptr {
		isPtr = true
		pkgPath = fmt.Sprintf("%s", T.Elem().PkgPath())
		typeName = fmt.Sprintf("%s", T.Elem().Name())
	} else {
		pkgPath = fmt.Sprintf("%s", T.PkgPath())
		typeName = fmt.Sprintf("%s", T.Name())
	}
	pkgPath = strings.TrimPrefix(pkgPath, "vendor/")
	rtn := fmt.Sprintf("%s.%s", pkgPath, typeName)
	if isPtr {
		rtn = "*" + rtn
	}
	return rtn
}
