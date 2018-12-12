package Component

import (
	"log"
	"reflect"
)

const (
	UNIQUE_TYPE_NONE    =iota //non-uniqueness
	UNIQUE_TYPE_LOCAL         //unique within this parent object
	UNIQUE_TYPE_GLOBAL        //unique global

)

type IComponent interface {
	Type() reflect.Type
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

type IAwake interface {
	Awake()
}

type IStart interface {
	Start(context *Context)
}

type IUpdate interface {
	Update(context *Context)
}

type IDestroy interface {
	Destroy()
}

type Context struct {
	Object    *Object     // The object the component is attached to.
	DeltaTime float32     // The delta step in global time for the update.
	Logger    *log.Logger // The runtime logger.
	Runtime   *Runtime
}


type IComponentBase interface {
	Init(parent *Object,t reflect.Type)
}

type Base struct {
	Parent *Object
	t reflect.Type
}

func (this *Base)Init(parent *Object,t reflect.Type)  {
	this.Parent=parent
	this.t=t
}

func (this *Base)GetParent()*Object  {
	return this.Parent
}

func (this *Base) Type() reflect.Type {
	return this.t
}

