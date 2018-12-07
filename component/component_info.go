package Component

import "reflect"

type componentInfo struct {
	Type      reflect.Type // The components type, cached
	Component IComponent   // The component instance
	Require   IRequire
	Parent    *Object      // The component mount root
	Active    int          // Number of frames this component has been active for
	Awake     IAwake
	Start     IStart       // IStart interface for component, if any
	Update    IUpdate      // IUpdate interface for component, if any
	Destroy   IDestroy
	Persist   IPersist // IPersist interface for component, if any

	Unique   IUnique  // Is it unique
}

func newComponentInfo(cmp IComponent,root *Object) *componentInfo {
	t:=reflect.TypeOf(cmp)
	rtn := &componentInfo{
		Type: t,
		Component: cmp,
		Parent:root,
		Active : 0}
	if rtn.Type.Implements(reflect.TypeOf((*IComponentBase)(nil)).Elem()) {
		rtn.Component.(IComponentBase).Init(root,t)
	}
	if rtn.Type.Implements(reflect.TypeOf((*IRequire)(nil)).Elem()) {
		rtn.Require = rtn.Component.(IRequire)
	}
	if rtn.Type.Implements(reflect.TypeOf((*IAwake)(nil)).Elem()) {
		rtn.Awake = rtn.Component.(IAwake)
	}
	if rtn.Type.Implements(reflect.TypeOf((*IStart)(nil)).Elem()) {
		rtn.Start = rtn.Component.(IStart)
	}
	if rtn.Type.Implements(reflect.TypeOf((*IUpdate)(nil)).Elem()) {
		rtn.Update = rtn.Component.(IUpdate)
	}
	if rtn.Type.Implements(reflect.TypeOf((*IPersist)(nil)).Elem()) {
		rtn.Persist = rtn.Component.(IPersist)
	}
	if rtn.Type.Implements(reflect.TypeOf((*IDestroy)(nil)).Elem()) {
		rtn.Destroy = rtn.Component.(IDestroy)
	}
	if rtn.Type.Implements(reflect.TypeOf((*IUnique)(nil)).Elem()) {
		rtn.Unique = rtn.Component.(IUnique)
	}

	return rtn
}

// IUpdate a single component
func (info *componentInfo) updateComponent(step float32, runtime *Runtime, context *Context) {
	if info.Active == 0 && info.Start != nil {
		runtime.workers.Run(func() {
			info.Start.Start(context)
			info.Active += 1
		})
	} else if info.Update != nil {
		runtime.workers.Run(func() {
			info.Update.Update(context)
			info.Active += 1
		})
	}
}