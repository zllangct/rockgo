package Component

import (
	errors2 "errors"
	"fmt"
	"github.com/zllangct/RockGO/3rd/errors"
	"github.com/zllangct/RockGO/3rd/iter"
	"github.com/zllangct/RockGO/logger"
	"github.com/zllangct/RockGO/utils/UUID"
	"reflect"
	"runtime/debug"
	"strings"
	"sync"
)

// Node is a game object type.
type Object struct {
	id         string
	name       string
	runtime    *Runtime
	components []*componentInfo // The set of components attached to this node
	children   []*Object        // The set of child objects attached to this node
	parent     *Object
	writeLock  *sync.Mutex
	locked     bool
}

// New returns a new Node
func NewObject(names ...string) *Object {
	name := ""
	if len(names) > 0 {
		name = names[0]
	}
	return &Object{
		id:         UUID.Next(),
		name:       name,
		runtime:    nil,
		components: make([]*componentInfo, 0),
		children:   make([]*Object, 0),
		writeLock:  &sync.Mutex{}}
}

func NewObjectWithComponent(component IComponent, names ...string) *Object {
	o := NewObject(names...)
	return o.AddComponent(component)

}

// Add a behaviour to a node
func (o *Object) AddComponent(component IComponent) *Object {
	info := newComponentInfo(component, o)
	err := o.WithLock(func() error {
		if info.Unique != nil {
			switch info.Unique.IsUnique() {
			case UNIQUE_TYPE_GLOBAL:
				_, err := o.Root().GetComponentsInChildren(reflect.TypeOf(component)).Next()
				if err == nil {
					return errors.Fail(ErrUniqueComponent{}, nil, "This component is unique global,one object already has a same component")
				}
			case UNIQUE_TYPE_LOCAL:
				if o.HasComponent(reflect.TypeOf(component)) {
					return errors.Fail(ErrUniqueComponent{}, nil, "This component is unique,this object already has a same component")
				}
			}
		}
		if info.Require != nil {
			for obj, requires := range info.Require.GetRequire() {
				for _, require := range requires {
					if !obj.HasComponent(require) {
						return errors.Fail(ErrMissingComponent{}, nil, "This component require other components,some are missing")
					}
				}
			}
		}
		o.components = append(o.components, info)
		if info.Awake != nil {
			info.Active += 1
			return info.Awake.Awake()
		}
		return nil
	})
	if err != nil {
		logger.Error(err)
	}
	return o
}

// remove the first component finded
func (o *Object) RemoveComponent(component IComponent) {
	err := o.WithLock(func() error {
		index := -1
		for i, v := range o.components {
			if v.Component == component {
				index = i
				break
			}
		}
		if index != -1 {
			o.components = append(o.components[:index], o.components[index+1:]...)
		}
		return nil
	})
	logger.Error(err)
}

//remove all components
func (o *Object) RemoveComponentsByType(t reflect.Type) {
	err := o.WithLock(func() error {
		for i := 0; i < len(o.components); i++ {
		A:
			if o.components[i].Type == t {
				o.components = append(o.components[:i], o.components[i+1:]...)
				if i < len(o.components) {
					goto A
				}
			}
		}
		return nil
	})
	logger.Error(err)
}

// Add a child object
func (o *Object) AddObject(object *Object) error {
	var err error
	err = object.Move(nil)
	if err != nil {
		return err
	}
	err = o.WithLock(func() error {
		if o == object || o.HasParent(object) {
			return errors.Fail(ErrBadObject{}, nil, "Circular object references are not permitted")
		} else {
			// Move the object into the new parent 'o'; this will lock the child and the old parent.
			err := object.Move(o)
			if err != nil {
				return err
			}
			// Now assign a new reference to this object
			o.children = append(o.children, object)
		}
		return nil
	})
	return err
}

// Move the object into a new parent object, which may also be nil
func (o *Object) Move(parent *Object) (err error) {
	oldParent := o.Parent()
	if oldParent != nil {
		if err = oldParent.RemoveObject(o); err != nil {
			return
		}
	}
	err = o.WithLock(func() error {
		o.parent = parent
		if parent != nil {
			o.runtime = parent.runtime // see Runtime()
		}
		return nil
	})
	return
}

func (o *Object) Destroy() (err error){
	err= o.WithLock(func() error {
		for _, cpt := range o.components {
			if cpt.Destroy != nil {
				err:=cpt.Destroy.Destroy()
				if err!=nil {
					logger.Error(err)
				}
			}
		}
		o.parent = nil
		o.runtime = nil
		for _, child := range o.children {
			err:=child.Destroy()
			if err!=nil {
				logger.Error(err)
			}
		}
		return nil
	})
	return err
}

// Remove a child object
func (o *Object) RemoveObject(object *Object) (err error) {
	if o == object {
		err = errors.Fail(ErrBadObject{}, nil, "Cannot remove object from itself")
		return
	}
	err = o.WithLock(func() error {
		offset := -1
		for i := 0; i < len(o.children); i++ {
			if o.children[i] == object {
				offset = i
				break
			}
		}
		if offset >= 0 {
			o.children = append(o.children[:offset], o.children[offset+1:]...)
		}
		err=object.Destroy()
		return err
	})
	return
}

// Check if an object has a parent
func (o *Object) HasParent(object *Object) bool {
	root := o
	for root != nil {
		root = root.Parent()
		if root == object {
			return true
		}
	}
	return false
}

// Return the parent of this object
func (o *Object) Parent() *Object {
	return o.parent
}

// Return the root object in the current object tree
func (o *Object) Root() *Object {
	i := o
	for {
		j := i.parent
		if j == nil {
			break
		}
		i = j
	}
	return i
}

// Objects returns an iterator of all immediate child objects on a game object
func (o *Object) Objects() iter.Iter {
	return fromObject(o, false)
}

// ObjectsInChildren returns an iterator of all the child objects on a game object
func (o *Object) ObjectsInChildren() iter.Iter {
	return fromObject(o, true)
}

// GetComponents returns an iterator of all components matching the given type.
func (o *Object) GetComponents(T reflect.Type) iter.Iter {
	return fromComponentArray(&o.components, T)
}

func (o *Object) AllComponents() iter.Iter {
	return fromComponentArray(&o.components, nil)
}

// GetComponentsInChildren returns an iterator of all components matching the given type in all children.
func (o *Object) GetComponentsInChildren(T reflect.Type) iter.Iter {
	cIter := fromComponentArray(nil, T)
	objIter := o.ObjectsInChildren()
	var val interface{} = nil
	var err error = nil
	for val, err = objIter.Next(); err == nil; val, err = objIter.Next() {
		componentList := &val.(*Object).components
		if len(*componentList) > 0 {
			cIter.Add(componentList)
		}
	}
	return cIter
}

// IUpdate all components in this object
func (o *Object) Update(step float32, runtime ...*Runtime) {
	activeRuntime := o.runtime
	if len(runtime) > 0 {
		activeRuntime = runtime[0]
	}
	clone := o.components
	context := o.NewContext(step, activeRuntime)
	for i := 0; i < len(clone); i++ {
		clone[i].updateComponent(step, activeRuntime, context)
	}
}

// Return a context for an object
func (o *Object) NewContext(delta float32, runtime ...*Runtime) *Context {
	activeRuntime := o.runtime
	if len(runtime) > 0 {
		activeRuntime = runtime[0]
	}
	return &Context{
		Object:    o,
		DeltaTime: delta,
		Runtime:   activeRuntime,
	}
}

// Extend an existing iterator with more objects
func (o *Object) addChildren(iterator *ObjectIter) {
	if len(o.children) > 0 {
		iterator.values.PushBack(&o.children)
	}
}

// Return the name for this object.
func (o *Object) Name() string {
	return o.name
}

// Rename the object
func (o *Object) Rename(name string) {
	err := o.WithLock(func() error {
		o.name = name
		return nil
	})
	logger.Error(err)
}

// Return the unique id of this object.
func (o *Object) ID() string {
	return o.id
}

// Find returns the first matching component on the object tree given by the name sequence or nil
// component should be a pointer to store the output component into.
// eg. If *FakeComponent implements IComponent, pass **FakeComponent to Find.
func (o *Object) Find(component interface{}, query ...string) error {
	componentType := reflect.TypeOf(component).Elem()

	obj := o
	var err error
	if len(query) != 0 {
		obj, err = o.FindObject(query...)
		if err != nil {
			return err
		}
	}

	cmp, err := obj.GetComponents(componentType).Next()
	if err != nil {
		return err
	}

	reflect.ValueOf(component).Elem().Set(reflect.ValueOf(cmp))
	return nil
}

func (o *Object) HasComponent(componentType reflect.Type) bool {
	_, err := o.GetComponents(componentType).Next()
	if err != nil {
		return false
	}
	return true
}

func (o *Object) Runtime() *Runtime {
	if o.runtime != nil {
		return o.runtime
	}
	marker := o.parent
	for marker != nil {
		if marker.runtime != nil {
			o.runtime = marker.runtime
			return o.runtime
		}
		marker = marker.parent
	}
	return nil
}

// FindObject returns the first matching child object on the object tree given by the name sequence or nil
func (o *Object) FindObject(query ...string) (*Object, error) {
	if len(query) == 0 {
		return nil, errors.Fail(ErrBadValue{}, nil, "Invalid query length of zero")
	}

	cursor := o
	queryCursor := 0

	var rtn *Object = nil
	for rtn == nil {
		next, err := cursor.GetObject(query[queryCursor])
		if err != nil {
			return nil, err
		} else {
			cursor = next
		}

		queryCursor += 1
		if queryCursor == len(query) {
			rtn = cursor
		}
	}

	return rtn, nil
}

func (o *Object) GetObject(name string) (*Object, error) {
	for i := 0; i < len(o.children); i++ {
		if o.children[i].name == name {
			return o.children[i], nil
		}
	}
	return nil, errors.Fail(ErrNoMatch{}, nil, fmt.Sprintf("No match for object '%s' on parent '%s'", name, o.name))
}

// HasObject is a single return value test for named object existence.
func (o *Object) HasObject(name string) bool {
	for i := 0; i < len(o.children); i++ {
		if o.children[i].name == name {
			return true
		}
	}
	return false
}

// Debug prints out a summary of the object and its components
func (o *Object) Debug(indents ...int) string {
	indent := 0
	if len(indents) > 0 {
		indent = indents[0]
	}

	name := o.name
	if len(name) == 0 {
		name = "Untitled"
	}

	rtn := fmt.Sprintf("object: %s (%d / %d)\n", name, len(o.children), len(o.components))
	if len(o.components) > 0 {
		for i := 0; i < len(o.components); i++ {
			rtn += fmt.Sprintf("! %s\n", typeName(o.components[i].Type))
		}
	}

	if len(o.children) > 0 {
		for i := 0; i < len(o.children); i++ {
			rtn += o.children[i].Debug(indent+1) + "\n"
		}
	}

	lines := strings.Split(rtn, "\n")
	prefix := strings.Repeat("  ", indent)
	if indent != 0 {
		prefix += " "
	}
	output := ""
	for i := 0; i < len(lines); i++ {
		if len(strings.Trim(lines[i], " ")) != 0 {
			output += prefix
			output += lines[i]
			if i != (len(lines) - 1) {
				output += "\n"
			}
		}
	}

	return output
}

// Lock allows you to safely perform some action without worrying about sync issues.
func (o *Object) WithLock(action func() error) (err error) {
	defer (func() {
		if r := recover(); r != nil {
			errs := r.(error).Error()
			err = errors2.New(errs + "\n" + string(debug.Stack()))
		}
		o.locked = false
		o.writeLock.Unlock()
	})()

	// fmt.Printf("Lock: %s\n", o.name)
	// debug.PrintStack()

	o.writeLock.Lock()
	o.locked = true
	err = action()
	return err
}
