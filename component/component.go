package Component

import (
	"reflect"
)

const (
	UNIQUE_TYPE_NONE    =iota //non-uniqueness
	UNIQUE_TYPE_LOCAL         //unique within this parent object
	UNIQUE_TYPE_GLOBAL        //unique global

)

type IComponent interface {
	Init(typ reflect.Type,runtime *Runtime,parent *Object)
	Type() reflect.Type
	Runtime()*Runtime
	Parent()*Object
}

type IUnique interface {
	IsUnique() int
}

type IRequire interface {
	GetRequire()(requires map[*Object][]reflect.Type)
}

type IPersist interface {
	Serialize() (interface{}, error)
	Deserialize(data interface{}) error
}

type Context struct {
	Object    *Object
	DeltaTime float32
	Runtime   *Runtime
}

type Base struct {
	parent  *Object
	runtime *Runtime
	typ     reflect.Type
}

func (this *Base)Init(typ reflect.Type,runtime *Runtime,parent *Object)  {
	this.typ =typ
	this.runtime =runtime
	this.parent=parent
}

func (this *Base) Type() reflect.Type {
	return this.typ
}

func (this *Base)Runtime()*Runtime  {
	return this.runtime
}

func (this *Base)Parent()*Object  {
	return this.parent
}
