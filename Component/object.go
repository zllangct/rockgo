package Component

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"

	"github.com/zllangct/RockGO/3RD/iter"
	"github.com/zllangct/RockGO/3RD/errors"
)

// Node is a game object type.
type Object struct {
	id         int64
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
		id:         makeObjectId(),
		name:       name,
		runtime:    nil,
		components: make([]*componentInfo, 0),
		children:   make([]*Object, 0),
		writeLock:  &sync.Mutex{}}
}

func NewObjectWithComponent(component IComponent,names ...string) *Object {
	o:=NewObject(names...)
	return o.AddComponent(component)

}

// Add a behaviour to a node
func (o *Object) AddComponent(component IComponent) *Object {
	info := newComponentInfo(component,o)
	o.WithLock(func() error {
		if info.Uniqual != nil && info.Uniqual.IsUnique() {
			if o.HasComponent(component) {
				return errors.Fail(ErrUniqueComponent{}, nil, "This componet is unique,the object already has a same component")
			}
		}
		o.components = append(o.components, info)
		if info.Awake!=nil{
			info.Awake.Awake()
		}
		return nil
	})
	return o
}

// remove the first component finded
func (o *Object) RemoveComponent(component IComponent) {
	o.WithLock(func() error {
		index:=-1
		for i,v:=range o.components{
			if v.Component == component{
				index=i
				break
			}
		}
		if index!=-1{
			o.components = append(o.components[:index],o.components[index+1:]...)
		}
		return nil
	})
}
//remove all components
func (o *Object) RemoveComponentsByType(t reflect.Type) {
	o.WithLock(func() error {
		for i:=0;i<len(o.components) ; i++ {
			A:
			if o.components[i].Type == t{
				o.components = append(o.components[:i],o.components[i+1:]...)
				if i < len(o.components){
					goto A
				}
			}
		}
		return nil
	})
}
// Add a child object
func (o *Object) AddObject(object *Object) error {
	object.Move(nil)
	return o.WithLock(func() error {
		if o == object || o.HasParent(object) {
			return errors.Fail(ErrBadObject{}, nil, "Circular object references are not permitted")
		} else {
			// Move the object into the new parent 'o'; this will lock the child and the old parent.
			object.Move(o)

			// Now assign a new reference to this object
			o.children = append(o.children, object)
		}
		return nil
	})
}

// Move the object into a new parent object, which may also be nil
func (o *Object) Move(parent *Object) error {
	oldParent := o.Parent()
	if oldParent != nil {
		if err := oldParent.RemoveObject(o); err != nil {
			return err
		}
	}
	return o.WithLock(func() error {
		o.parent = parent
		o.runtime = parent.runtime // see Runtime()
		return nil
	})
}

// Remove a child object
func (o *Object) RemoveObject(object *Object) error {
	if o == object {
		return errors.Fail(ErrBadObject{}, nil, "Cannot remove object from itself")
	}
	return o.WithLock(func() error {
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
		object.WithLock(func() error {
			for _, cpt := range object.components {
				if cpt.Destroy != nil {
					cpt.Destroy.Destroy()
				}
			}
			object.parent = nil
			object.runtime = nil
			return nil
		})

		return nil
	})
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
		Logger:    activeRuntime.logger,
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
	o.WithLock(func() error {
		o.name = name
		return nil
	})
}

// Return the unique id of this object.
func (o *Object) ID() int64 {
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
//TODO 需测试验证
func (o *Object) HasComponent(component interface{}) bool {
	componentType := reflect.TypeOf(component).Elem()
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

func (o *Object) Logger() *log.Logger {
	runtime := o.runtime
	if runtime != nil {
		return runtime.logger
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
			err = r.(error)
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
