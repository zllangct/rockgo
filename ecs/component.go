package ecs

import (
	"reflect"
)

const (
	UNIQUE_TYPE_NONE   = iota //non-uniqueness
	UNIQUE_TYPE_LOCAL         //unique within this parent object
	UNIQUE_TYPE_GLOBAL        //unique global

)

type IComponent interface {
	Init(typ reflect.Type, runtime *Runtime, parent *Object)
	Type() reflect.Type
	Runtime() *Runtime
	Parent() *Object
	Root() *Object
}

//组件唯一性
type IUnique interface {
	IsUnique() int
}

//组件依赖检查
type IRequire interface {
	GetRequire() (requires map[*Object][]reflect.Type)
}

//持久化接口
type IPersist interface {
	Serialize() (interface{}, error)
	Deserialize(data interface{}) error
}

//Init 会立即执行，等同于构造函数，用于保证顺序
type IInit interface {
	Initialize() error
}

type Context struct {
	Object    *Object
	DeltaTime float32
	Runtime   *Runtime
}

//组件基类
type ComponentBase struct {
	parent  *Object
	runtime *Runtime
	typ     reflect.Type
}

func (this *ComponentBase) Init(typ reflect.Type, runtime *Runtime, parent *Object) {
	this.typ = typ
	this.runtime = runtime
	this.parent = parent
}

func (this *ComponentBase) Type() reflect.Type {
	return this.typ
}

func (this *ComponentBase) Runtime() *Runtime {
	return this.runtime
}

func (this *ComponentBase) Parent() *Object {
	return this.parent
}

func (this *ComponentBase) Root() *Object {
	return this.runtime.Root()
}

func (this *ComponentBase) GetComponent(cpt interface{}) error {
	return this.Parent().Find(cpt)
}

func (this *ComponentBase) AddComponent(cpt IComponent) *Object {
	return this.Parent().AddComponent(cpt)
}
