package Component

import "reflect"

type componentInfo struct {
	Type      reflect.Type // The components type, cached
	Component Component    // The component instance
	Active    int          // Number of frames this component has been active for
	Start     Start        // Start interface for component, if any
	Update    Update       // Update interface for component, if any
	Destroy   Destroy
	Persist   Persist // Persist interface for component, if any
	Uniqual   Unique  // Is it unique
}

func newComponentInfo(cmp Component) *componentInfo {
	rtn := &componentInfo{
		Type: cmp.Type(),
		Component: cmp,
		Active : 0}
	if rtn.Type.Implements(reflect.TypeOf((*Start)(nil)).Elem()) {
		rtn.Start = rtn.Component.(Start)
	}
	if rtn.Type.Implements(reflect.TypeOf((*Update)(nil)).Elem()) {
		rtn.Update = rtn.Component.(Update)
	}
	if rtn.Type.Implements(reflect.TypeOf((*Persist)(nil)).Elem()) {
		rtn.Persist = rtn.Component.(Persist)
	}
	if rtn.Type.Implements(reflect.TypeOf((*Destroy)(nil)).Elem()) {
		rtn.Destroy = rtn.Component.(Destroy)
	}
	if rtn.Type.Implements(reflect.TypeOf((*Unique)(nil)).Elem()) {
		rtn.Uniqual = rtn.Component.(Unique)
	}
	return rtn
}

// Update a single component
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