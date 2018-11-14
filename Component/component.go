package Component

import (
	"reflect"
	"log"
)

//TODO 组件的Destroy
//TODO Object 的 Destroy ,当对象销毁时,需要触发对象孩子的销毁,组件的销毁

// Component is a unit of functionality that can be attached to objects.
type Component interface {
	// Type returns the Type information for this component
	Type() reflect.Type

}

// Persist should be implemented by components that need to persist state across serialization.
type Persist interface {
	// Serialize returns a string serialization of the data for the component
	Serialize() (interface{}, error)

	// Deserialize loads the string serialization of the data for the component
	Deserialize(data interface{}) error
}

// Awake components are invoked when a component is assigned to an object.
type Awake interface {
	// Awake is invoked immediately after a component is attached to an object.
	Awake(parent *Object)
}

// Start components are invoked the first frame they are active.
type Start interface {
	// Start is invoked the first frame a component is active
	Start(context *Context)
}

// Update components are updated every frame.
type Update interface {
	// Update the component this frame.
	Update(context *Context)
}

//TODO 添加组件销毁事件
type Destroy interface {
	Destroy(context *Context)
}

// Context provides a reference back to the owning game object and runtime state for a component
type Context struct {
	Object    *Object     // The object the component is attached to.
	DeltaTime float32     // The delta step in global time for the update.
	Logger    *log.Logger // The runtime logger.
	Runtime   *Runtime
}
