package Component

import (
	"errors"
	"fmt"
	"github.com/zllangct/RockGO/3rd/iter"
	"github.com/zllangct/RockGO/logger"
	"github.com/zllangct/RockGO/utils"
	"github.com/zllangct/RockGO/utils/UUID"
	"reflect"
	"runtime/debug"
	"strings"
	"sync"
)

type Object struct {
	id         string
	name       string
	runtime    *Runtime
	components []IComponent
	children   []*Object 
	parent     *Object
	locker     *sync.RWMutex
}

// 新建一个实体对象
func NewObject(names ...string) *Object {
	name := ""
	if len(names) > 0 {
		name = names[0]
	}
	return &Object{
		id:         UUID.Next(),
		name:       name,
		runtime:    nil,
		components: make([]IComponent, 0),
		children:   make([]*Object, 0),
		locker:     &sync.RWMutex{}}
}



//添加组件
func (o *Object) AddComponent(component IComponent) *Object {
	component.Init(reflect.TypeOf(component),o.Runtime(),o)
	err := o.WithLock(func() error {
		if unique,ok:=component.(IUnique);ok{
			switch unique.IsUnique() {
			case UNIQUE_TYPE_GLOBAL:
				_, err := o.Root().GetComponentsInChildren(reflect.TypeOf(component)).Next()
				if err == nil {
					return ErrUniqueComponent
				}
			case UNIQUE_TYPE_LOCAL:
				if o.HasComponent(reflect.TypeOf(component)) {
					return ErrUniqueComponent
				}
			}
		}
		if require,ok:=component.(IRequire);ok{
			for obj, requires := range require.GetRequire() {
				for _, require := range requires {
					if !obj.HasComponent(require) {
						return ErrMissingComponent
					}
				}
			}
		}
		o.components = append(o.components, component)

		runtime:= o._runtime()
		if runtime != nil{
			runtime.SystemFilter(component)
		}
		return nil
	})
	if err != nil {
		logger.Error(err)
	}
	utils.Try(func() {
		if init,ok:=component.(IInit);ok{
			init.Initialize()
		}
	})
	return o
}

//移除找到的第一个对应组件
func (o *Object) RemoveComponent(component IComponent) {
	err := o.WithLock(func() error {
		index := -1
		for i, v := range o.components {
			if v == component {
				index = i
				break
			}
		}
		if index != -1 {
			o.components = append(o.components[:index], o.components[index+1:]...)
		}
		runtime:=o.runtime
		if runtime ==nil{
			return errors.New("this object has no runtime")
		}
		err:=runtime.SystemOperate("update",SYSTEM_OP_UPDATE_REMOVE,component)
		if err!=nil {
			logger.Error(err)
		}
		err=runtime.SystemOperate("destroy",SYSTEM_OP_DESTROY_ADD,component)
		if err!=nil {
			logger.Error(err)
		}
		return nil
	})
	if err!=nil{
		logger.Error(err)
	}
}

//移除该类型的所有组件
func (o *Object) RemoveComponentsByType(t reflect.Type) {
	err := o.WithLock(func() error {
		for index, component := range o.components {
			if component.Type() == t {
				o.components = append(o.components[:index], o.components[index+1:]...)
				runtime:=o.Runtime()
				if runtime ==nil{
					return errors.New("this object has no runtime")
				}
				err:=runtime.SystemOperate("update",SYSTEM_OP_UPDATE_REMOVE,component)
				if err!=nil {
					logger.Error(err)
				}
				err=runtime.SystemOperate("destroy",SYSTEM_OP_DESTROY_ADD,component)
				if err!=nil {
					logger.Error(err)
				}
			}
		}
		return nil
	})
	logger.Error(err)
}

func (o *Object) AddObjectWithComponent(object *Object,component IComponent) error {
 	err:=o.AddObject(object)
	if err!=nil {
		return  err
	}
 	object.AddComponent(component)
	return nil
}

func (o *Object) AddObjectWithComponents(object *Object,components []IComponent) error {
	err:=o.AddObject(object)
	if err!=nil {
		return  err
	}
	for _, component:= range components {
		object.AddComponent(component)
	}
	return nil
}

func (o *Object) AddNewObjectWithComponent(component IComponent,name ...string)(*Object,error) {
	obj:=NewObject(name...)
	return obj, o.AddObjectWithComponent(obj,component)
}

func (o *Object) AddNewbjectWithComponents(components []IComponent,name ...string)(*Object,error){
	obj:=NewObject(name...)
	return obj, o.AddObjectWithComponents(obj,components)
}

//添加子对象
func (o *Object) AddObject(object *Object) error {
	var err error
	if object.Parent() != nil {
		return ErrBadObject
	}
	err = o.WithLock(func() error {
		if o == object || o.HasParent(object) {
			return ErrBadObject
		} else {
			// Move the object into the new parent 'o'; this will lock the child and the old parent.
			object.parent=o
			object.runtime=o.runtime
			// Now assign a new reference to this object
			o.children = append(o.children, object)
		}
		return nil
	})
	return err
}
//释放所有引用
func (o *Object) _dispose(){
	o.name=""
	o.runtime=nil
	o.components=nil
	o.children=nil
}

//销毁实体本身及其子对象，无法解除父对象对自己的引用
func (o *Object) _destroy() {
	err:= o.WithLock(func() error {
		for _, component := range o.components {
			runtime:=o.runtime
			if runtime ==nil{
				continue
			}
			err:=runtime.SystemOperate("update",SYSTEM_OP_UPDATE_REMOVE,component)
			if err!=nil {
				logger.Error(err)
			}
			err=runtime.SystemOperate("destroy",SYSTEM_OP_DESTROY_ADD,component)
			if err!=nil {
				logger.Error(err)
			}
		}
		for _, child := range o.children {
			child._destroy()
		}
		o._dispose()
		return nil
	})
	if err!=nil {
		logger.Error(err)
	}
}

// 销毁实体
func (o *Object) Destroy() (err error) {
	p:=o.Parent()
	if  p!=nil {
		err=p.RemoveObject(o)
		if err != nil && err == ErrNoThisChild {}else{
			return
		}
	}
	o._destroy()
	return
}

// 移除实体
func (o *Object) RemoveObject(object *Object) (err error) {
	if o == object {
		err = ErrBadObject
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
		if offset < 0 {
			return nil
		}
		o.children = append(o.children[:offset], o.children[offset+1:]...)
		object._destroy()
		return nil
	})
	return
}

// 判断实体是否有父节点
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

// 获取实体父节点
func (o *Object) Parent() *Object {
	return o.parent
}

// 获取实体根节点
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

// 获取实体对应类型组件，当T为nil时匹配所有对象
func (o *Object) GetComponents(T reflect.Type) iter.Iter {
	return fromComponentArray(&o.components, T)
}

//获取所有组件
func (o *Object) AllComponents() iter.Iter {
	return fromComponentArray(&o.components, nil)
}

//获取子对象中所对应的组件，当T为nil时匹配所有对象
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

func (o *Object) Name() string {
	o.locker.RLock()
	defer o.locker.RUnlock()
	return o.name
}

func (o *Object) Rename(name string) {
	err := o.WithLock(func() error {
		o.name = name
		return nil
	})
	logger.Error(err)
}

func (o *Object) ID() string {
	o.locker.RLock()
	defer o.locker.RUnlock()
	return o.id
}

//查找组件，component为要查询组件的指针，query 查询条件为实体名
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

//实体是否拥有该组件
func (o *Object) HasComponent(componentType reflect.Type) bool {
	_, err := o.GetComponents(componentType).Next()
	if err != nil {
		return false
	}
	return true
}

//获取运行时
func (o *Object) Runtime() *Runtime {
	o.locker.RLock()
	defer o.locker.RUnlock()

	return o._runtime()
}
func (o *Object) _runtime() *Runtime {
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

//查找实体
func (o *Object) FindObject(query ...string) (*Object, error) {
	if len(query) == 0 {
		return nil,ErrBadValue
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

//获取实体，通过实体名
func (o *Object) GetObject(name string) (*Object, error) {
	for i := 0; i < len(o.children); i++ {
		if o.children[i].name == name {
			return o.children[i], nil
		}
	}
	return nil, ErrNoMatch
}

//判断是否有该实体
func (o *Object) HasObject(name string) bool {
	for i := 0; i < len(o.children); i++ {
		if o.children[i].name == name {
			return true
		}
	}
	return false
}

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
			rtn += fmt.Sprintf("! %s\n", typeName(o.components[i].Type()))
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

func (o *Object) WithLock(action func() error) (err error) {
	defer (func() {
		if r := recover(); r != nil {
			var str string
			switch r.(type) {
			case error:
				str =r.(error).Error()
			case string:
				str = r.(string)
			}
			err = errors.New(str+ string(debug.Stack()))
		}
		o.locker.Unlock()
	})()

	o.locker.Lock()
	err = action()
	return err
}
