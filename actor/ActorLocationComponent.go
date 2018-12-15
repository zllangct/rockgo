package Actor

import (
	"errors"
	"fmt"
	"github.com/zllangct/RockGO/cluster"
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/configComponent"
	"reflect"
	"sync"
)

type ActorLocationComponent struct {
	Component.Base
	locker        sync.RWMutex
	nodeComponent *Cluster.NodeComponent
	actors        *sync.Map // [role,*ActorIDGroup]
}

func (this *ActorLocationComponent) IsUnique() int {
	return Component.UNIQUE_TYPE_GLOBAL
}

func (this *ActorLocationComponent) GetRequire() map[*Component.Object][]reflect.Type {
	requires := make(map[*Component.Object][]reflect.Type)
	requires[this.Parent.Root()] = []reflect.Type{
		reflect.TypeOf(&Config.ConfigComponent{}),
		reflect.TypeOf(&Cluster.NodeComponent{}),
	}
	return requires
}

func (this *ActorLocationComponent) Awake() error {
	this.actors = &sync.Map{}
	err := this.Parent.Root().Find(&this.nodeComponent)
	if err != nil {
		return err
	}
	err = this.nodeComponent.Register(this)
	if err != nil {
		return err
	}
	return nil
}

func (this *ActorLocationComponent)ServiceInvalidChecking() {

}

func (this *ActorLocationComponent) ServiceInquiry(role string, reply *ActorIDGroup) error {
	if g, ok := this.actors.Load(role); ok {
		*reply = *(g.(*ActorIDGroup))
	} else {
		return errors.New(fmt.Sprintf("no any actor of this role :%s", role))
	}
	return nil
}

type ActorService struct {
	Role    string
	ActorID ActorID
}

func (this *ActorLocationComponent) ServiceRegister(args ActorService, reply *bool) error {
	if g, ok := this.actors.LoadOrStore(args.Role, &ActorIDGroup{
		Actors: []ActorID{args.ActorID},
	}); ok {
		g.(*ActorIDGroup).Add(args.ActorID)
	}
	return nil
}

func (this *ActorLocationComponent) ServiceUnregister(args ActorService, reply *bool) error {
	this.actors.Range(func(key, value interface{}) bool {
		value.(*ActorIDGroup).Sub(args.ActorID)
		return true
	})
	return nil
}


