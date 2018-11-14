package Actor

import (
	"reflect"
	"github.com/zllangct/RockGO/Component"
	"sync"
)

/*
	Actor Manager Component
*/

type ActorManagerComponent struct {
	Dic sync.Map  // key,value = int64,*Component.Object
}

func (this *ActorManagerComponent) Type() reflect.Type {
	return reflect.TypeOf(this)
}

func (this *ActorManagerComponent) Update(ctx *Component.Context) {

}

func (this *ActorManagerComponent) New() Component.Component {
	return &ActorManagerComponent{}
}

func (this *ActorManagerComponent) Serialize() (interface{}, error) {
	return nil, nil
}

func (this *ActorManagerComponent) Deserialize(raw interface{}) error {
	return nil
}

func (this *ActorManagerComponent) Add(entity *Component.Object)  {
	this.Dic.Store(entity.ID(),entity)
}

func (this *ActorManagerComponent) Remove(id int64)  {
	this.Dic.Delete(id)
}

func (this *ActorManagerComponent) Get(id int64) (*Component.Object,bool) {
	e,ok:=this.Dic.Load(id)
	return e.(*Component.Object),ok
}