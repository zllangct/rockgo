package Component

import (
	"log"
	"reflect"
)


type IComponent interface {
	Type() reflect.Type
}

type IUnique interface {
	IsUnique() bool
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
	Init(parent *Object)
}

type Base struct {
	Parent *Object
}

func (this *Base)Init(parent *Object)  {
	this.Parent=parent
}

func (this *Base)GetParent()*Object  {
	return this.Parent
}

func (this *Base) Type() reflect.Type {
	return reflect.TypeOf(this)
}