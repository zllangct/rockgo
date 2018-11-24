package Component

import (
	"log"
	"reflect"
)

//TODO 组件的Destroy
//TODO Object 的 Destroy ,当对象销毁时,需要触发对象孩子的销毁,组件的销毁

type Component interface {
	Type() reflect.Type
}

type Unique interface {
	IsUnique() bool
}

type Persist interface {
	Serialize() (interface{}, error)
	Deserialize(data interface{}) error
}

type Start interface {
	Start(context *Context)
}

type Update interface {
	Update(context *Context)
}

//TODO 添加组件销毁事件
type Destroy interface {
	Destroy()
}

type Context struct {
	Object    *Object     // The object the component is attached to.
	DeltaTime float32     // The delta step in global time for the update.
	Logger    *log.Logger // The runtime logger.
	Runtime   *Runtime
}
